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

package server

import (
	"errors"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
	"net"
	"strconv"
	"sync"

	"github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/connectionutil"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/tunmgr"

	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
)

//NetPacket defines the packet struct
type NetPacket struct {
	Packet []byte
}

//NetPacketIP defines the IP packet struct
type NetPacketIP struct {
	Packet   *NetPacket
	ClientID string
}

//IPTypes defines struct that holds virtual IP and privateIP
type IPTypes struct {
	VirtualIP string
	PrivateIP string
}

const (
	logTag               = "[mnedcserver]"
	serverVirtualAddress = "10.0.0.1/24"
	virtualIPPrefix      = "10.0.0."
	channelSize          = 200
	packetSize           = 1024
)

var (
	serverIns      *Server
	tunIns         tunmgr.Tun
	networkUtilIns connectionutil.NetworkUtil
	log            = logmgr.GetInstance()
)

//Server defines MNEDC server struct
type Server struct {
	listener                net.Listener
	intf                    *water.Interface
	virtualIP               net.IP
	netMask                 *net.IPNet
	isAlive                 bool
	clientCount             int
	clients                 map[string]*clientConnection
	clientsLock             sync.Mutex
	clientAddressByDeviceID map[string]string
	clientIDByAddress       map[string]string
	incomingChannel         chan *NetPacket
	outgoingChannel         chan *NetPacket
	incomingIPPacketChan    chan *NetPacketIP
	clientIPInfoByDeviceID  map[string]IPTypes
}

//MNEDCServer declares methods related to MNEDC server
type MNEDCServer interface {
	CreateServer(string, string, bool) (*Server, error)
	Run()
	AcceptRoutine()
	HandleConnection(net.Conn)
	SetVirtualIP(string) string
	DispatchRoutine()
	Route(*NetPacket)
	SetClientAddress(string, string)
	SetClientIP(string, string, string)
	RemoveClient(string)
	GetClientIPMap() map[string]IPTypes
	TunReadRoutine()
	TunWriteRoutine()
	Close() error
}

func init() {
	serverIns = &Server{}
	tunIns = tunmgr.GetInstance()
	networkUtilIns = connectionutil.GetInstance()
}

//GetInstance returns server instance
func GetInstance() MNEDCServer {
	return serverIns
}

//CreateServer creates a Server structure and returns it
func (s *Server) CreateServer(address, port string, isSecure bool) (*Server, error) {
	logPrefix := logTag + "[NewServer]"

	src := address + ":" + port

	listener, err := networkUtilIns.ListenIP(src, isSecure)
	if err != nil {
		log.Println(logPrefix, "Listen error", err.Error())
		return nil, err
	}

	virtualIP, virtualNetMask, err := net.ParseCIDR(serverVirtualAddress)
	if err != nil {
		return nil, errors.New(logPrefix + " Invalid network/mask:" + err.Error())
	}

	s.listener = listener
	s.virtualIP = virtualIP
	s.netMask = virtualNetMask
	s.isAlive = true
	s.clientCount = 1
	s.clients = map[string]*clientConnection{}
	s.clientIDByAddress = map[string]string{}
	s.clientAddressByDeviceID = map[string]string{}
	s.clientIPInfoByDeviceID = map[string]IPTypes{}
	s.incomingChannel = make(chan *NetPacket, channelSize)
	s.outgoingChannel = make(chan *NetPacket, channelSize)
	s.incomingIPPacketChan = make(chan *NetPacketIP, channelSize)

	intf, err := tunIns.CreateTUN()
	if err != nil {
		s.Close()
		return nil, errors.New(logPrefix + " TUN error: " + err.Error())
	}
	s.intf = intf

	setIPError := tunIns.SetTUNIP(s.intf.Name(), s.virtualIP, s.netMask, true)
	if setIPError != nil {
		err = setIPError
	}
	setStatusError := tunIns.SetTUNStatus(s.intf.Name(), true, true)
	if setStatusError != nil {
		err = setStatusError
	}

	if err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}

//Run starts the server and starts to accept clients
func (s *Server) Run() {
	go s.AcceptRoutine()   //handle new Client connection
	go s.DispatchRoutine() //handle packets and route them to proper place
	go s.TunReadRoutine()  //read from tun interface
	go s.TunWriteRoutine() //write to tun interface

	log.Println(logTag, "Server started")
}

//AcceptRoutine accepts client connections
func (s *Server) AcceptRoutine() {
	logPrefix := logTag + "[acceptRoutine]"

	for s.isAlive {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println(logPrefix, "Some connection error:", err.Error())
		}
		if conn != nil {
			s.HandleConnection(conn)
		}
	}
}

//HandleConnection handles new client registration
func (s *Server) HandleConnection(conn net.Conn) {

	logPrefix := logTag + "[handleConnection]"

	remoteAddr := conn.RemoteAddr().String()
	log.Println(logPrefix, "Connection request from"+remoteAddr)
	buf := make([]byte, packetSize)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println(logPrefix, "error in read", err.Error())
		conn.Close()
		return
	}

	deviceID := string(buf[0:n])

	log.Println(logPrefix, "Client connected! IP:", remoteAddr)
	clientVirtualIP := s.SetVirtualIP(deviceID)

	_, err = conn.Write([]byte(clientVirtualIP))

	if err != nil {
		log.Println(logPrefix, "parameters sending failed", err.Error())
		conn.Close()
		return
	}

	c := clientConnection{
		conn: conn,
	}

	s.clientsLock.Lock()

	c.deviceID = deviceID
	s.clients[c.deviceID] = &c
	s.clientIDByAddress[clientVirtualIP] = c.deviceID

	s.clientsLock.Unlock()

	c.initClient(s)
}

//SetVirtualIP builds the parameters to be sent to client
func (s *Server) SetVirtualIP(deviceID string) string {

	var ip string

	if val, ok := s.clientAddressByDeviceID[deviceID]; ok {
		ip = val
	} else {
		s.clientCount = s.clientCount + 1
		ip = virtualIPPrefix + strconv.Itoa(s.clientCount)
		s.clientsLock.Lock()
		s.clientAddressByDeviceID[deviceID] = ip
		s.clientsLock.Unlock()
	}
	log.Println(logTag, "[NewConnection]", "The ip given is", ip)

	return ip
}

//DispatchRoutine sends packets from inboundIPPkts/inboundDevPkts to client/TUN.
func (s *Server) DispatchRoutine() {
	for s.isAlive {
		select {
		case pkt := <-s.incomingIPPacketChan:
			s.Route(pkt.Packet)
		case pkt := <-s.incomingChannel:
			s.Route(pkt)
		}
	}
}

//Route channels the packet to appropriate destination
func (s *Server) Route(pkt *NetPacket) {
	logPrefix := logTag + "[route]"

	dest := waterutil.IPv4Destination(pkt.Packet)

	s.clientsLock.Lock()
	destClientID, canRouteDirectly := s.clientIDByAddress[dest.String()]
	if canRouteDirectly {
		destClient, clientExists := s.clients[destClientID]
		if clientExists {
			destClient.queueIP(pkt)
		} else {
			log.Println(logPrefix, "WARN: Attempted to route packet to clientID", destClientID, "which does not exist. Dropping")
		}
	}
	s.clientsLock.Unlock()
	if !canRouteDirectly {
		s.outgoingChannel <- pkt
	}
}

//SetClientAddress puts the device ID in the map
func (s *Server) SetClientAddress(deviceID string, addr string) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	s.clientIDByAddress[addr] = deviceID
}

//SetClientIP saves new device ipInfo from broadcast server
func (s *Server) SetClientIP(deviceID, privateIP, virtualIP string) {
	ipInfo := IPTypes{
		PrivateIP: privateIP,
		VirtualIP: virtualIP,
	}
	s.clientIPInfoByDeviceID[deviceID] = ipInfo
}

//RemoveClient removes a client connection
func (s *Server) RemoveClient(deviceID string) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	//delete from the clientIDByAddress map if it exists
	var toDeleteAddrs []string
	for dest, itemID := range s.clientIDByAddress {
		if itemID == deviceID {
			toDeleteAddrs = append(toDeleteAddrs, dest)
		}
	}
	for _, addr := range toDeleteAddrs {
		delete(s.clientIDByAddress, addr)
	}
	delete(s.clients, deviceID)
	delete(s.clientIPInfoByDeviceID, deviceID)
}

//GetClientIPMap returns clientIPInfoByDeviceID map to proxy server
func (s *Server) GetClientIPMap() map[string]IPTypes {
	return s.clientIPInfoByDeviceID
}

//TunReadRoutine reads from tun virtual interface and sends the packet to server
func (s *Server) TunReadRoutine() {
	for s.isAlive {
		vpnbuf := make([]byte, packetSize)
		n, err := s.intf.Read(vpnbuf)
		if n > 0 && err == nil {

			pkt := &NetPacket{
				Packet: vpnbuf,
			}
			s.incomingChannel <- pkt
		}
	}
}

//TunWriteRoutine reads the outgoing packets from server and writes on tun virtual Interface
func (s *Server) TunWriteRoutine() {
	for s.isAlive {
		pkt := <-s.outgoingChannel
		vpnbuf := []byte(pkt.Packet)

		_, _ = s.intf.Write(vpnbuf)
	}
}

// Close shuts down the server, reversing configuration changes to the system.
func (s *Server) Close() error {

	if !s.isAlive {
		return errors.New("Server not alive")
	}

	s.isAlive = false
	err := s.listener.Close()
	if err != nil {
		return err
	}
	if s.intf != nil {
		err = s.intf.Close()
	}

	return err
}
