/*******************************************************************************
 * Copyright 2020 Samsung Electronics All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *******************************************************************************/

package client

import (
	"bufio"
	"errors"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
	"net"
	"os"
	"sync"
	"time"

	restclient "github.com/lf-edge/edge-home-orchestration-go/src/restinterface/client"
	//"controller/discoverymgr"
	networkhelper "github.com/lf-edge/edge-home-orchestration-go/src/common/networkhelper"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/connectionutil"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/tunmgr"

	"github.com/songgao/water"
)

const (
	packetSize  = 1024
	channelSize = 200

	logTag = "[mnedcclient]"
)

//NetPacket defines the packet struct
type NetPacket struct {
	Packet []byte
}

//Client defines the MNEDC client struct
type Client struct {
	conn            net.Conn
	incomingChannel chan *NetPacket
	outgoingChannel chan *NetPacket
	isAlive         bool
	isConnected     bool
	isSecure        bool
	intf            *water.Interface
	virtualIP       net.IP
	netMask         *net.IPNet
	mutexLock       sync.Mutex
	serverIP        string
	serverPort      string
	deviceID        string
	configPath      string
	clientAPI       restclient.Clienter
}

var (
	clientIns      *Client
	tunIns         tunmgr.Tun
	networkUtilIns connectionutil.NetworkUtil
	networkIns     networkhelper.Network
	//discoveryIns   discoverymgr.Discovery
	log            = logmgr.GetInstance()
)

const (
	waitDelay                = 150 * time.Millisecond
	retryDelay               = 5 * time.Second
	mnedcBroadcastServerPort = 3333
)

//MNEDCClient declares methods related to MNEDC client
type MNEDCClient interface {
	Run()
	CreateClient(string, string, bool) (*Client, error)
	Close() error
	StartSendRoutine()
	StartRecvRoutine()
	NotifyClose()
	ConnectionReconciled()
	HandleError(error)
	ParseVirtualIP(string) error
	TunReadRoutine()
	TunWriteRoutine()
	NotifyBroadcastServer(configPath string) error
	SetClient(clientAPI restclient.Clienter)
}

func init() {
	clientIns = &Client{}
	tunIns = tunmgr.GetInstance()
	networkUtilIns = connectionutil.GetInstance()
	networkIns = networkhelper.GetInstance()
	//discoveryIns = discoverymgr.GetInstance()
}

//GetInstance returns MNEDCClient interface instance
func GetInstance() MNEDCClient {
	return clientIns
}

//CreateClient creates the MNEDC client
func (c *Client) CreateClient(deviceID, configPath string, isSecure bool) (*Client, error) {
	logPrefix := logTag + "[CreateClient]"

	for {
		servAddr, servPort, err := getMNEDCServerAddress(configPath)
		if err != nil {
			return nil, errors.New("Cannot read config file, " + err.Error())
		}
		conn, err := networkUtilIns.ConnectToHost(servAddr, servPort, isSecure) //register to MNEDC server
		if err != nil {
			log.Println(logPrefix, "Dial failed", err.Error(), ", retrying")
			time.Sleep(retryDelay)
			continue
		}

		buf := []byte(deviceID) //send the key in packet(Currently hardcoded)
		err = networkUtilIns.WriteTo(conn, buf)
		if err != nil {
			log.Println(logPrefix, "Secret Write error", err.Error(), ", retrying")
			conn.Close()
			time.Sleep(retryDelay)
			continue
		}

		log.Println("Write Successful")
		//conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		readBufSize, readBuf, err := networkUtilIns.ReadFrom(conn)

		if err != nil {
			log.Println(logPrefix, "Read Error: ", err.Error(), ", retrying")
			time.Sleep(retryDelay)
			conn.Close()
			continue
		}

		intf, err := tunIns.CreateTUN()
		if err != nil {
			c.Close()
			return nil, errors.New(logPrefix + " TUN error: " + err.Error())
		}

		c.conn = conn
		c.incomingChannel = make(chan *NetPacket, channelSize)
		c.outgoingChannel = make(chan *NetPacket, channelSize)
		c.isAlive = true
		c.isConnected = true
		c.intf = intf
		c.serverIP = servAddr
		c.serverPort = servPort
		c.deviceID = deviceID
		c.isSecure = isSecure
		c.configPath = configPath

		err = c.ParseVirtualIP(string(readBuf[0:readBufSize])) //unique TUN ip sent by server
		if err == nil {
			setIPError := tunIns.SetTUNIP(c.intf.Name(), c.virtualIP, c.netMask, true)
			if setIPError != nil {
				err = setIPError
			}
			setStatusError := tunIns.SetTUNStatus(c.intf.Name(), true, true)
			if setStatusError != nil {
				err = setStatusError
			}
		}

		if err != nil {
			c.Close()
		}

		return c, err
	}
}

//Run starts the MNEDC client
func (c *Client) Run() {
	go c.StartSendRoutine()
	go c.StartRecvRoutine()
	go c.TunReadRoutine()
	go c.TunWriteRoutine()
}

//StartSendRoutine reads from incomingChannel and writes on server connection
func (c *Client) StartSendRoutine() {

	for c.isAlive {

		for c.isConnected {
			pkt, ok := <-c.incomingChannel
			if !ok {
				break
			}

			err := networkUtilIns.WriteTo(c.conn, pkt.Packet)
			if err != nil {
				log.Println("error in send: ", err)
				c.mutexLock.Lock()
				c.isConnected = false
				c.mutexLock.Unlock()
				c.HandleError(err)
				break
			}
		}
		time.Sleep(waitDelay)
		dropSendBuffer(c.incomingChannel)
	}
}

//StartRecvRoutine reads from server connection and writes on outgoingChannel
func (c *Client) StartRecvRoutine() {
	for c.isAlive {

		for c.isConnected {
			_, vpnbuf, err := networkUtilIns.ReadFrom(c.conn)
			if err != nil {
				log.Println("Read error", err.Error())
				c.mutexLock.Lock()
				c.isConnected = false
				c.mutexLock.Unlock()
				c.HandleError(err)
				break
			}
			c.outgoingChannel <- &NetPacket{Packet: vpnbuf}
		}
		time.Sleep(waitDelay)
	}
}

func dropSendBuffer(buffer chan *NetPacket) {
	for {
		select {
		case <-buffer:
		default:
			return
		}
	}
}

//ParseVirtualIP parses the parameters sent by server
func (c *Client) ParseVirtualIP(parameters string) error {
	virtualIP, virtualNetMask, err := net.ParseCIDR(parameters + "/24")
	if err != nil {
		return errors.New(parameters + "/24" + "is Invalid network/mask " + err.Error())
	}

	c.netMask = virtualNetMask
	c.virtualIP = virtualIP
	return nil
}

// Close shuts down the client, reversing configuration changes to the system.
func (c *Client) Close() error {

	if !c.isAlive {
		return errors.New("Client not alive")
	}

	c.isAlive = false
	var err error
	if c.conn != nil {
		err = c.conn.Close()
	}

	if c.intf != nil {
		err = c.intf.Close()
	}

	c.NotifyClose()

	return err
}

//HandleError handles the error occurred in MNEDC client connection
func (c *Client) HandleError(err error) {
	logPrefix := "[hadError]"
	log.Println(logTag, logPrefix, err.Error(), ", Connection problem detected. Re-connecting.")

	if !c.isAlive {
		return
	}

	c.mutexLock.Lock()
	defer c.mutexLock.Unlock()

	c.isConnected = false
	c.conn.Close()
	if c.intf != nil {
		c.intf.Close()
	}

	c.NotifyClose()

	deviceID := c.deviceID

	for {
		servAddr, servPort, err := getMNEDCServerAddress(c.configPath)
		if err != nil {
			log.Println(errors.New("Cannot read config file," + err.Error()))
			return
		}
		conn, err := networkUtilIns.ConnectToHost(servAddr, servPort, c.isSecure) //register to MNEDC server
		if err != nil {
			log.Println(logPrefix, "Dial failed", err.Error(), ", retrying")
			time.Sleep(retryDelay)
			continue
		}

		buf := []byte(deviceID)
		err = networkUtilIns.WriteTo(conn, buf)
		if err != nil {
			log.Println(logPrefix, "Secret Write error", err.Error(), ", retrying")
			conn.Close()
			time.Sleep(retryDelay)
			continue
		}

		n, readBuf, err := networkUtilIns.ReadFrom(conn)
		if err != nil {
			log.Println(logPrefix, "Read Error: ", err.Error(), ", retrying")
			time.Sleep(retryDelay)
			conn.Close()
			continue
		}

		intf, err := tunIns.CreateTUN()
		if err != nil {
			log.Println(logPrefix, "TUN error:", err.Error())
			return
		}

		c.intf = intf
		err = c.ParseVirtualIP(string(readBuf[0:n])) //unique TUN ip sent by server
		if err != nil {
			log.Println(logPrefix, "Parse Virtual IP error:", err.Error())
			return
		}

		err = tunIns.SetTUNIP(c.intf.Name(), c.virtualIP, c.netMask, true)
		if err != nil {
			log.Println(logPrefix, "TUN error:", err.Error())
			return
		}

		err = tunIns.SetTUNStatus(c.intf.Name(), true, true)
		if err != nil {
			log.Println(logPrefix, "TUN error:", err.Error())
			return
		}

		c.serverIP = servAddr
		c.serverPort = servPort
		c.conn = conn
		c.isConnected = true

		time.Sleep(3 * time.Second)

		go c.ConnectionReconciled()
		return
	}
}

//TunReadRoutine reads from TUN interface and writes on incomingChannel
func (c *Client) TunReadRoutine() {

	for c.isAlive {
		for c.isConnected {
			packet := make([]byte, packetSize)
			n, err := c.intf.Read(packet)
			if err != nil {
				continue
			}
			p := &NetPacket{
				Packet: packet[:n],
			}
			c.incomingChannel <- p
		}
		time.Sleep(waitDelay)
	}
}

//TunWriteRoutine reads from outgoingChannel and writes on TUN interface
func (c *Client) TunWriteRoutine() {

	for c.isAlive {
		for c.isConnected {
			pkt := <-c.outgoingChannel
			w, err := c.intf.Write(pkt.Packet)
			if err != nil {
				continue
			}
			if w != len(pkt.Packet) {
				log.Println(logTag, "Write to tun interface has mismatched len:", w, "!=", len(pkt.Packet))
			}
		}
		time.Sleep(waitDelay)
	}
}

//NotifyClose handles the case when MNEDC connection is closed
func (c *Client) NotifyClose() {
	logPrefix := "[NotifyClose]"
	log.Println(logPrefix, "MNEDC connection closed")
	//discoveryIns.MNEDCClosedCallback()
}

//ConnectionReconciled handles the case when MNEDC connection is re-established
func (c *Client) ConnectionReconciled() {
	logPrefix := "[connectionReIstablish]"
	log.Println(logPrefix, "MNEDC connection reistablished")
	//discoveryIns.MNEDCReconciledCallback()
	//notifyBroadcastServer()
	c.NotifyBroadcastServer(c.configPath)
}

//NotifyBroadcastServer sends request to broadcast server
func (c *Client) NotifyBroadcastServer(configPath string) error {
	logPrefix := "[RegisterBroadcast]"
	log.Println(logTag, "Registering to Broadcast server")
	c.configPath = configPath
	virtualIP, err := networkIns.GetVirtualIP()
	if err != nil {
		log.Println(logPrefix, "Cant register to Broadcast server, virtual IP error", err.Error())
		return err
	}

	privateIP, err := networkIns.GetOutboundIP()
	if err != nil {
		log.Println(logPrefix, "Cant register to Broadcast server, outbound IP error", err.Error())
		return err
	}

	file, err := os.Open(configPath)

	if err != nil {
		log.Println(logPrefix, "cant read config file from", configPath, err.Error())
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	scanner.Scan()
	serverIP := scanner.Text()

	go func() {

		if c.clientAPI == nil {
			log.Println(logPrefix, "Client is nil, returning")
			err = errors.New("Client is nil")
			return
		}
		err = c.clientAPI.DoNotifyMNEDCBroadcastServer(serverIP, mnedcBroadcastServerPort, c.deviceID, privateIP, virtualIP)
		if err != nil {
			log.Println(logPrefix, "Cannot register to Broadcast server", err.Error())
		}
	}()

	time.Sleep(5 * time.Second)
	if err != nil {
		return err
	}

	return nil
}

//SetClient sets the rest client
func (c *Client) SetClient(clientAPI restclient.Clienter) {
	c.clientAPI = clientAPI
}

func getMNEDCServerAddress(path string) (string, string, error) {

	var serverIP, serverPort string

	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	scanner.Scan()
	serverIP = scanner.Text()
	scanner.Scan()
	serverPort = scanner.Text()

	return serverIP, serverPort, nil
}
