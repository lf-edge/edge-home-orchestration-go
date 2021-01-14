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
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/songgao/water"

	networkUtilMocks "github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/connectionutil/mocks"
	tunMocks "github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/tunmgr/mocks"
)

var (
	tunIntf         *water.Interface
	mockTun         *tunMocks.MockTun
	mockNetworkUtil *networkUtilMocks.MockNetworkUtil
	runningServer   *Server
	listener        net.Listener
)

const (
	edgeDir          = "/var/edge-orchestration/"
	deviceIDFilePath = edgeDir + "/device/orchestration_deviceID.txt"
	defaultID        = "edge-orchestration-dummy"
	defaultIP        = "2.2.2.2"
	defaultVirtualIP = "10.0.0.2"
	correctPassword  = "test"
	wrongPassword    = "wrong"
	edgePrefix       = "edge-orchestration-"
	localhost        = "127.0.0.1"
	defaultMsg       = "dummy"

	defaultTCPPort = "8000"
	defaultTunPort = "9999"
)

func init() {
	tunIntf = &water.Interface{}
	startConnection()
}

func startConnection() {
	list, err := net.Listen("tcp", ":"+defaultTunPort)
	if err != nil {
		log.Println("Listen error in test, free port", defaultTunPort)
		return
	}
	listener = list
}

func makeConnection() (net.Conn, error) {
	conn, err := net.Dial("tcp", ":"+defaultTunPort)
	if err != nil {
		log.Println("Some connection error:", err.Error())
		return nil, err
	}
	return conn, nil
}

func acceptConnection() {
	conn, err := listener.Accept()
	if err != nil {
		log.Println("Some connection error:", err.Error())
		return
	}
	tunIntf.ReadWriteCloser = conn
}
func closeConnection() {
	log.Println("Close called")
	if listener != nil {
		listener.Close()
	}
}

func TestCreateServer(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	go acceptConnection()
	time.Sleep(2 * time.Second)

	createMockIns(ctrl)

	_, err := makeConnection()
	if err != nil {
		log.Println("Cannot Make connection")
		return
	}

	defaultListener, err := net.Listen("tcp", ":"+defaultTCPPort)
	if err != nil {
		log.Println("Listen error in test, free port", defaultTCPPort)
		return
	}

	t.Run("ListenError", func(t *testing.T) {
		mockNetworkUtil.EXPECT().ListenIP(gomock.Any(), gomock.Any()).Return(nil, errors.New("Listen error"))
		serverInstance := GetInstance()
		_, err := serverInstance.CreateServer("", "", false)
		if err == nil {
			t.Error("Server should not be started")
		}
	})

	t.Run("TunError", func(t *testing.T) {
		mockNetworkUtil.EXPECT().ListenIP(gomock.Any(), gomock.Any()).Return(defaultListener, nil)
		mockTun.EXPECT().CreateTUN().Return(nil, errors.New("TUN error"))
		serverInstance := GetInstance()
		_, err := serverInstance.CreateServer("", "", false)
		if err == nil {
			t.Error("Server should not be started")
		}
	})

	t.Run("TunIPError", func(t *testing.T) {

		mockNetworkUtil.EXPECT().ListenIP(gomock.Any(), gomock.Any()).Return(defaultListener, nil)
		mockTun.EXPECT().CreateTUN().Return(tunIntf, nil)
		mockTun.EXPECT().SetTUNIP(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockTun.EXPECT().SetTUNStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("TUN Errror"))

		serverInstance := GetInstance()
		_, err := serverInstance.CreateServer("", "", false)
		if err == nil {
			t.Error("Server should not be started")
		}
	})

	t.Run("Success", func(t *testing.T) {

		mockNetworkUtil.EXPECT().ListenIP(gomock.Any(), gomock.Any()).Return(defaultListener, nil)
		mockTun.EXPECT().CreateTUN().Return(tunIntf, nil)
		mockTun.EXPECT().SetTUNIP(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockTun.EXPECT().SetTUNStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		serverInstance := GetInstance()
		server, err := serverInstance.CreateServer("", "", false)
		if err != nil {
			t.Error("Server not started")
		}
		_ = server.Close()
	})
}

func TestClientMaps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("Success", func(t *testing.T) {

		serverInstance := &Server{
			clientIPInfoByDeviceID:  map[string]IPTypes{},
			clientIDByAddress:       map[string]string{},
			clientAddressByDeviceID: map[string]string{},
			clientCount:             1,
		}

		serverInstance.SetClientIP(defaultID, defaultIP, defaultVirtualIP)

		serverInstance.SetClientAddress(defaultID, defaultIP)

		_ = serverInstance.GetClientIPMap()
		delete(serverInstance.clientIPInfoByDeviceID, defaultID)
		delete(serverInstance.clientIDByAddress, defaultIP)

		_ = serverInstance.SetVirtualIP(defaultID)
		_ = serverInstance.SetVirtualIP(defaultID)

		delete(serverInstance.clientAddressByDeviceID, defaultID)

	})
}

func TestNewConnection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("ErrorCases", func(t *testing.T) {
		_, err := makeConnection()
		if err != nil {
			log.Println("Cannot Make connection")
			return
		}
		defaultListener, err := net.Listen("tcp", ":8001")
		if err != nil {
			log.Println("Listen error in test, free port", defaultTCPPort)
			return
		}
		mockNetworkUtil.EXPECT().ListenIP(gomock.Any(), gomock.Any()).Return(defaultListener, nil)
		mockTun.EXPECT().CreateTUN().Return(tunIntf, nil)
		mockTun.EXPECT().SetTUNIP(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockTun.EXPECT().SetTUNStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		serverInstance := GetInstance()
		server, err := serverInstance.CreateServer("", "", false)
		if err != nil {
			t.Error("Server not started")
			return
		}

		go server.AcceptRoutine()
		time.Sleep(2 * time.Second)

		conn, err := net.Dial("tcp", ":8001")
		if err != nil {
			t.Error("Cannot connect to server")
			return
		}

		conn.Close()
		server.Close()

	})
	t.Run("Success", func(t *testing.T) {
		_, err := makeConnection()
		if err != nil {
			log.Println("Cannot Make connection")
			return
		}
		defaultListener, err := net.Listen("tcp", ":8002")
		if err != nil {
			log.Println("Listen error in test, free port", defaultTCPPort)
			return
		}
		mockNetworkUtil.EXPECT().ListenIP(gomock.Any(), gomock.Any()).Return(defaultListener, nil)
		mockTun.EXPECT().CreateTUN().Return(tunIntf, nil)
		mockTun.EXPECT().SetTUNIP(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockTun.EXPECT().SetTUNStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		serverInstance := GetInstance()
		server, err := serverInstance.CreateServer("", "", false)
		if err != nil {
			t.Error("Server not started")
			return
		}

		go server.AcceptRoutine()
		time.Sleep(2 * time.Second)

		conn, err := net.Dial("tcp", ":8002")
		if err != nil {
			t.Error("Cannot connect to server")
			return
		}
		conn.Write([]byte(defaultID))
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			t.Error("Couldn't register with correct password")
			conn.Close()
			return
		}
		_, _, err = net.ParseCIDR(string(buf[0:n]) + "/24")
		if err != nil {
			t.Error("Invalid params received")
			conn.Close()
			return
		}
		server.SetClientIP(defaultID, defaultIP, defaultVirtualIP)
		conn.Close()
		time.Sleep(2 * time.Second)
		_ = server.Close()
	})
}

func TestDispatchAndRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	serverIns = &Server{
		isAlive:     true,
		clientCount: 1,

		clients:                 map[string]*clientConnection{},
		clientIDByAddress:       map[string]string{},
		clientAddressByDeviceID: map[string]string{},
		clientIPInfoByDeviceID:  map[string]IPTypes{},

		incomingChannel:      make(chan *NetPacket, 200),
		outgoingChannel:      make(chan *NetPacket, 200),
		incomingIPPacketChan: make(chan *NetPacketIP, 200), //Packet channel related to clients
	}

	t.Run("SendPacketWithNoDestination", func(t *testing.T) {

		packet := make([]byte, 1024)
		packet[16] = 1
		packet[17] = 1
		packet[18] = 1
		packet[19] = 1

		noDestPacket := &NetPacket{
			Packet: packet,
		}

		serverIns.Route(noDestPacket)
	})

	t.Run("SendPacketWithDestination", func(t *testing.T) {

		packet := make([]byte, 1024)
		packet[16] = 127
		packet[17] = 0
		packet[18] = 0
		packet[19] = 1
		serverIns.clientIDByAddress[localhost] = defaultID

		pkt := &NetPacket{
			Packet: packet,
		}

		serverIns.Route(pkt)

		delete(serverIns.clientIDByAddress, localhost)
	})

	t.Run("DispatchRoutineIncomingChannel", func(t *testing.T) {
		packet := make([]byte, 1024)
		packet[16] = 1
		packet[17] = 1
		packet[18] = 1
		packet[19] = 1

		pkt := &NetPacket{
			Packet: packet,
		}

		go serverIns.DispatchRoutine()
		time.Sleep(2 * time.Second)
		serverIns.incomingChannel <- pkt
		time.Sleep(1 * time.Second)
	})

	t.Run("DispatchRoutineIncomingIPChannel", func(t *testing.T) {
		packet := make([]byte, 1024)
		packet[16] = 1
		packet[17] = 1
		packet[18] = 1
		packet[19] = 1

		pkt := &NetPacket{
			Packet: packet,
		}

		pktIP := &NetPacketIP{
			Packet:   pkt,
			ClientID: defaultID,
		}

		go serverIns.DispatchRoutine()
		time.Sleep(2 * time.Second)
		serverIns.incomingIPPacketChan <- pktIP
		time.Sleep(1 * time.Second)
	})
}

func TestDispatchRoutine(t *testing.T) {

}
func TestTunReadRoutine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn, err := makeConnection()
	if err != nil {
		log.Println("Cannot Make connection")
		return
	}

	createMockIns(ctrl)

	server := &Server{
		incomingChannel: make(chan *NetPacket, 200),
		isAlive:         true,
		intf:            tunIntf,
	}
	go server.TunReadRoutine()
	time.Sleep(time.Second * 2)
	conn.Write([]byte(defaultMsg))
	time.Sleep(1 * time.Second)

	tunIntf.Close()
}

func TestTunWriteRoutine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := makeConnection()
	if err != nil {
		log.Println("Cannot Make connection")
		return
	}

	createMockIns(ctrl)

	server := &Server{
		outgoingChannel: make(chan *NetPacket, 200),
		isAlive:         true,
		intf:            tunIntf,
	}
	packet := []byte(defaultMsg)
	pkt := &NetPacket{
		Packet: packet,
	}
	go server.TunWriteRoutine()
	time.Sleep(time.Second * 2)

	server.outgoingChannel <- pkt
	tunIntf.Close()
	closeConnection()
}

func TestRunAndQueueIP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)
	server := &Server{
		isAlive: false,
	}

	server.Run()

	clientConn := &clientConnection{
		outgoingChannel: make(chan *NetPacket, 10),
	}
	packet := []byte(defaultMsg)
	pkt := &NetPacket{
		Packet: packet,
	}
	clientConn.queueIP(pkt)
}
func createMockIns(ctrl *gomock.Controller) {
	mockTun = tunMocks.NewMockTun(ctrl)
	tunIns = mockTun

	mockNetworkUtil = networkUtilMocks.NewMockNetworkUtil(ctrl)
	networkUtilIns = mockNetworkUtil
}
