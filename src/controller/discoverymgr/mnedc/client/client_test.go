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
	"errors"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/songgao/water"

	networkUtilMocks "github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/connectionutil/mocks"
	discoveryMocks "github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mocks"
	tunMocks "github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/tunmgr/mocks"
)

const (
	defaultVirtualIP      = "10.0.0.2"
	defaultIP             = "2.2.2.2"
	defaultWrongIP        = "1234.5678.1234.5678"
	defaultID             = "edge-orchestration-dummy"
	defaultConnectionPort = "9999"
	defaultServerPort     = "9930"
	defaultServerIP       = "1.1.1.1"
	defaultConfigPath     = "client.Config"
	defaultMessage        = "dummy"
)

var (
	tunIntf         *water.Interface
	mockTun         *tunMocks.MockTun
	mockNetworkUtil *networkUtilMocks.MockNetworkUtil
	mockDiscovery   *discoveryMocks.MockDiscovery
	listener        net.Listener
	serverListener  net.Listener
	connection      net.Conn
)

func init() {
	tunIntf = &water.Interface{}
	startServerConnection()
	startConnection()
	createDeviceIDFile()
}

func TestParseVirtualIP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)
	t.Run("IPError", func(t *testing.T) {
		client := GetInstance()
		err := client.ParseVirtualIP(defaultWrongIP)
		if err == nil {
			t.Error("Error should be there for wrong IP")
			return
		}
	})
	t.Run("Success", func(t *testing.T) {
		client := GetInstance()
		err := client.ParseVirtualIP(defaultVirtualIP)
		if err != nil {
			t.Error("Error should not be there for proper IP")
			return
		}
	})
}
func TestMNEDCClosedAndReestablished(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)
	t.Run("NotifyClose", func(t *testing.T) {
		client := GetInstance()
		mockDiscovery.EXPECT().MNEDCClosedCallback().Return()
		client.NotifyClose()
	})
	t.Run("ConnectionEstablished", func(t *testing.T) {
		client := GetInstance()
		mockDiscovery.EXPECT().MNEDCReconciledCallback().Return()
		client.ConnectionReconciled()
	})
}

func TestStartSendRoutine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("SendRoutine", func(t *testing.T) {
		client := &Client{}
		p := &NetPacket{
			Packet: []byte(defaultMessage),
		}

		client.isAlive = true
		client.isConnected = true
		client.serverIP = defaultServerIP
		client.serverPort = defaultConnectionPort
		client.incomingChannel = make(chan *NetPacket, 200)

		mockNetworkUtil.EXPECT().WriteTo(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		go client.StartSendRoutine()
		time.Sleep(2 * time.Second)

		client.incomingChannel <- p
		client.isAlive = false
	})
}

func TestStartRecvRoutine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("RecvRoutine", func(t *testing.T) {
		client := &Client{}

		client.isAlive = true
		client.isConnected = true
		client.serverIP = defaultServerIP
		client.serverPort = defaultConnectionPort
		client.incomingChannel = make(chan *NetPacket, 200)
		client.outgoingChannel = make(chan *NetPacket, 200)

		mockNetworkUtil.EXPECT().ReadFrom(gomock.Any()).Return(5, []byte(defaultMessage), nil).AnyTimes()
		go client.StartRecvRoutine()
		time.Sleep(1 * time.Second)

		client.isAlive = false
	})
}

func TestStartRecvRoutineError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	go acceptServerConnection()
	time.Sleep(2 * time.Second)

	conn, err := makeConnection(defaultServerPort)
	if err != nil {
		log.Println("Cannot Make connection")
		return
	}

	go acceptConnection()
	time.Sleep(2 * time.Second)

	_, err = makeConnection(defaultConnectionPort)
	if err != nil {
		log.Println("Cannot Make connection")
		return
	}

	t.Run("RecvRoutineErr", func(t *testing.T) {
		client := &Client{}
		client.isAlive = true
		client.isConnected = true
		client.serverIP = defaultServerIP
		client.serverPort = defaultConnectionPort
		client.incomingChannel = make(chan *NetPacket, 200)
		client.outgoingChannel = make(chan *NetPacket, 200)
		client.intf = tunIntf
		client.conn = conn

		deleteDeviceIDFile()
		gomock.InOrder(
			mockNetworkUtil.EXPECT().ReadFrom(gomock.Any()).Return(5, []byte(defaultMessage), errors.New("")),
			mockDiscovery.EXPECT().MNEDCClosedCallback().Return(),
		)

		go client.StartRecvRoutine()
		time.Sleep(1 * time.Second)

		client.isAlive = false
	})
	t.Run("SendRoutineErr", func(t *testing.T) {
		client := &Client{}
		p := &NetPacket{
			Packet: []byte(defaultMessage),
		}

		client.mutexLock.Lock()
		client.isAlive = true
		client.isConnected = true
		client.mutexLock.Unlock()
		client.serverIP = defaultServerIP
		client.serverPort = defaultConnectionPort
		client.incomingChannel = make(chan *NetPacket, 200)
		client.intf = tunIntf
		client.conn = conn

		gomock.InOrder(
			mockNetworkUtil.EXPECT().WriteTo(gomock.Any(), gomock.Any()).Return(errors.New(defaultMessage)),
			mockDiscovery.EXPECT().MNEDCClosedCallback().Return(),
		)

		go client.StartSendRoutine()
		time.Sleep(2 * time.Second)

		client.incomingChannel <- p
		time.Sleep(1 * time.Second)
		client.isAlive = false
	})
}

func TestTunWriteRoutine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	go acceptConnection()
	time.Sleep(2 * time.Second)

	createMockIns(ctrl)

	t.Run("Success", func(t *testing.T) {
		_, err := makeConnection(defaultConnectionPort)
		if err != nil {
			log.Println("Cannot Make connection")
			return
		}
		client := &Client{
			intf:            tunIntf,
			isAlive:         true,
			incomingChannel: make(chan *NetPacket, 200),
			outgoingChannel: make(chan *NetPacket, 200),
			isConnected:     true,
		}
		p := &NetPacket{
			Packet: []byte(defaultMessage),
		}
		go client.TunWriteRoutine()
		time.Sleep(2 * time.Second)
		client.outgoingChannel <- p
		time.Sleep(1 * time.Second)
		client.intf.Close()
		client.outgoingChannel <- p
		client.isConnected = false
		client.isAlive = false
	})
}

func TestTunReadRoutine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	go acceptConnection()
	time.Sleep(2 * time.Second)

	createMockIns(ctrl)

	t.Run("Success", func(t *testing.T) {
		conn, err := makeConnection(defaultConnectionPort)
		if err != nil {
			log.Println("Cannot Make connection")
			return
		}
		client := &Client{
			intf:            tunIntf,
			incomingChannel: make(chan *NetPacket, 200),
			outgoingChannel: make(chan *NetPacket, 200),
			conn:            conn,
		}
		client.mutexLock.Lock()
		client.isAlive = true
		client.isConnected = true
		client.mutexLock.Unlock()
		go client.TunReadRoutine()
		time.Sleep(2 * time.Second)
		conn.Write([]byte(defaultMessage))
		time.Sleep(1 * time.Second)
		client.isConnected = false
		client.isAlive = false
	})
}
func TestHandleErrorNegativeRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)
	createDeviceIDFile()

	go acceptServerConnection()
	time.Sleep(2 * time.Second)

	conn, err := makeConnection(defaultServerPort)
	if err != nil {
		log.Println("Cannot Make connection")
		return
	}

	go acceptConnection()
	time.Sleep(2 * time.Second)

	_, err = makeConnection(defaultConnectionPort)
	if err != nil {
		log.Println("Cannot Make connection")
		return
	}

	t.Run("isAliveFalse", func(t *testing.T) {
		client := &Client{
			isAlive: false,
		}
		client.HandleError(errors.New(""))
	})

	t.Run("ReadFromError", func(t *testing.T) {
		gomock.InOrder(
			mockDiscovery.EXPECT().MNEDCClosedCallback(),
			mockNetworkUtil.EXPECT().ConnectToHost(gomock.Any(), gomock.Any(), gomock.Any()).Return(conn, nil).AnyTimes(),
			mockNetworkUtil.EXPECT().WriteTo(gomock.Any(), gomock.Any()).Return(nil).AnyTimes(),
			mockNetworkUtil.EXPECT().ReadFrom(gomock.Any()).Return(1, []byte(""), errors.New("")).Return(8, []byte(defaultVirtualIP), nil),
			mockTun.EXPECT().CreateTUN().Return(nil, errors.New("")),
		)

		client := &Client{
			serverIP:    defaultServerIP,
			serverPort:  defaultConnectionPort,
			conn:        conn,
			isConnected: true,
			deviceID:    defaultID,
			isAlive:     true,
			configPath:  defaultConfigPath,
		}
		client.HandleError(errors.New(""))
	})

}

func TestHadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	go acceptServerConnection()
	time.Sleep(2 * time.Second)

	conn, err := makeConnection(defaultServerPort)
	if err != nil {
		log.Println("Cannot Make connection")
		return
	}

	go acceptConnection()
	time.Sleep(2 * time.Second)

	_, err = makeConnection(defaultConnectionPort)
	if err != nil {
		log.Println("Cannot Make connection")
		return
	}

	c := &Client{
		serverIP:    defaultServerIP,
		serverPort:  defaultConnectionPort,
		conn:        conn,
		isConnected: true,
		deviceID:    defaultID,
		isAlive:     true,
		configPath:  defaultConfigPath,
	}
	gomock.InOrder(
		mockDiscovery.EXPECT().MNEDCClosedCallback(),
		mockNetworkUtil.EXPECT().ConnectToHost(gomock.Any(), gomock.Any(), gomock.Any()).Return(conn, nil),
		mockNetworkUtil.EXPECT().WriteTo(gomock.Any(), gomock.Any()).Return(nil),
		mockNetworkUtil.EXPECT().ReadFrom(gomock.Any()).Return(8, []byte(defaultVirtualIP), nil),
		mockTun.EXPECT().CreateTUN().Return(tunIntf, nil),
		mockTun.EXPECT().SetTUNIP(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		mockTun.EXPECT().SetTUNStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		mockDiscovery.EXPECT().MNEDCReconciledCallback().Return().AnyTimes(),
	)

	c.HandleError(errors.New("Read Error"))

	if c.isConnected == false {
		t.Error("Connection could not be re-established")
	}
	conn.Close()
	connection.Close()
}

func TestCreateClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	go acceptConnection()
	time.Sleep(2 * time.Second)

	createMockIns(ctrl)
	_, err := makeConnection(defaultConnectionPort)
	if err != nil {
		log.Println("Cannot Make connection")
		return
	}

	go acceptServerConnection()
	time.Sleep(2 * time.Second)

	conn, err := makeConnection(defaultServerPort)
	if err != nil {
		log.Println("Cannot Make connection")
		return
	}

	gomock.InOrder(
		mockNetworkUtil.EXPECT().ConnectToHost(gomock.Any(), gomock.Any(), gomock.Any()).Return(conn, nil),
		mockNetworkUtil.EXPECT().WriteTo(gomock.Any(), gomock.Any()).Return(nil),
		mockNetworkUtil.EXPECT().ReadFrom(gomock.Any()).Return(8, []byte(defaultVirtualIP), nil),
		mockTun.EXPECT().CreateTUN().Return(tunIntf, nil),
		mockTun.EXPECT().SetTUNIP(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		mockTun.EXPECT().SetTUNStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
	)

	client := GetInstance()
	c, err := client.CreateClient(defaultID, defaultConfigPath, false)

	if err != nil {
		t.Error("Err should be nil")
		return
	}

	log.Println("Connection Status: ", c.isConnected)
	if c.isConnected == false {
		t.Error("isConnected should be true")
		return
	}

	time.Sleep(1 * time.Second)
	deleteDeviceIDFile()
}

func TestRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	client := &Client{
		isAlive:     false,
		isConnected: false,
	}
	client.Run()
}

func TestClose(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	go acceptConnection()
	time.Sleep(2 * time.Second)

	createMockIns(ctrl)
	_, err := makeConnection(defaultConnectionPort)
	if err != nil {
		log.Println("Cannot Make connection")
		return
	}

	go acceptServerConnection()
	time.Sleep(2 * time.Second)

	conn, err := makeConnection(defaultServerPort)
	if err != nil {
		log.Println("Cannot Make connection")
		return
	}

	client := &Client{
		conn:    conn,
		intf:    tunIntf,
		isAlive: true,
	}
	mockDiscovery.EXPECT().MNEDCClosedCallback().Return()
	err = client.Close()

	if err != nil {
		t.Error("Error should be nil")
	}

	closeConnection()
}

func createMockIns(ctrl *gomock.Controller) {
	mockTun = tunMocks.NewMockTun(ctrl)
	mockNetworkUtil = networkUtilMocks.NewMockNetworkUtil(ctrl)
	tunIns = mockTun
	networkUtilIns = mockNetworkUtil
	mockDiscovery = discoveryMocks.NewMockDiscovery(ctrl)
	discoveryIns = mockDiscovery
}

func startConnection() {
	list, err := net.Listen("tcp", ":"+defaultConnectionPort)
	if err != nil {
		log.Println("Listen error in test, free port", defaultConnectionPort)
		return
	}
	listener = list
}

func makeConnection(port string) (net.Conn, error) {
	conn, err := net.Dial("tcp", ":"+port)
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
	if serverListener != nil {
		serverListener.Close()
	}
}

func startServerConnection() {
	list, err := net.Listen("tcp", ":"+defaultServerPort)
	if err != nil {
		log.Println("Listen error in test, free port", defaultServerPort)
		return
	}
	serverListener = list
}

func acceptServerConnection() {
	conn, err := serverListener.Accept()
	if err != nil {
		log.Println("Some connection error:", err.Error())
		return
	}
	connection = conn
}

func createDeviceIDFile() error {
	f, err := os.Create(defaultConfigPath)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(defaultServerIP + "\n" + defaultConnectionPort + "\n")
	if err != nil {
		return err
	}

	f.Sync()
	return nil
}

func deleteDeviceIDFile() {
	err := os.Remove(defaultConfigPath)
	if err != nil {
		log.Println("Could not delete file")
	}
}
