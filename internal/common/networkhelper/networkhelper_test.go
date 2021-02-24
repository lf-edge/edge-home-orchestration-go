/*******************************************************************************
 * Copyright 2019-2020 Samsung Electronics All Rights Reserved.
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

package networkhelper

import (
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/networkhelper/detector/mocks"
	"errors"
	"net"
	"reflect"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
)

var (
	TESTIPV4       = "11.111.111.11"
	TESTIPV6       = "fe80::f112:19a8:eca4:724f"
	TESTMAC        = "AB:CD:EF:GH:IJ:KL"
	TESTNETIFLIST  []net.Interface
	TESTERR        error
	TESTNEWIP      net.IP
	TESTNEWWIREDIP net.IP
	TESTNEWIPS     []net.IP
	TESTIFINDEX    = 1

	isCalledGetNetInfo bool
	wait               sync.WaitGroup
	network            Network
)

type addrstr struct {
	name string
	addr string
}

func (str addrstr) Network() string {
	return str.name
}

func init() {
	TESTERR = errors.New("ERROR")
	TESTNETIF := net.Interface{Index: TESTIFINDEX}
	TESTNETIFLIST = append(TESTNETIFLIST, TESTNETIF)
	TESTNEWIP = net.ParseIP("22.222.222.22")
	TESTNEWWIREDIP = net.ParseIP("11.111.111.11")
	TESTNEWIPS = []net.IP{TESTNEWIP}

	network = GetInstance()
}

func setPassCondOfNetInfo() {
	netInfo.addrInfos = make([]addrInformation, 1)
	netInfo.addrInfos[0].ipv4 = TESTNEWIP
	netInfo.addrInfos[0].macAddr = TESTMAC
	netInfo.netInterface = TESTNETIFLIST
	netInfo.netError = nil
}

func setFailCondOfNetInfo() {
	netInfo.addrInfos = make([]addrInformation, 1)
	netInfo.addrInfos[0].ipv4 = TESTNEWIP
	netInfo.addrInfos[0].macAddr = TESTMAC
	netInfo.netInterface = nil
	netInfo.netError = TESTERR
}

func TestGetNetInterface(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setPassCondOfNetInfo()
		netif, err := network.GetNetInterface()
		if err != nil {
			t.Error()
		}
		if len(netif) == 0 || netif[0].Index != TESTIFINDEX {
			t.Error()
		}
	})
	t.Run("Fail", func(t *testing.T) {
		setFailCondOfNetInfo()
		_, err := network.GetNetInterface()
		if err == nil {
			t.Error()
		}
	})
	t.Log()
}

func TestCheckConnectivity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setPassCondOfNetInfo()
		err := network.CheckConnectivity()
		if err != nil {
			t.Error()
		}
	})
	t.Run("Fail", func(t *testing.T) {
		setFailCondOfNetInfo()
		err := network.CheckConnectivity()
		if err == nil {
			t.Error()
		}
	})
}
func TestGetOutboundIP(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setPassCondOfNetInfo()
		ip, err := network.GetOutboundIP()
		if err != nil {
			t.Error()
		}
		if ip != TESTNEWIP.String() {
			t.Error()
		}

	})
	t.Run("Fail", func(t *testing.T) {
		setFailCondOfNetInfo()
		_, err := network.GetOutboundIP()
		if err == nil {
			t.Error()
		}
	})
}

func TestGetMACAddress(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setPassCondOfNetInfo()
		mac, err := network.GetMACAddress()
		if err != nil {
			t.Error()
		}
		if mac != TESTMAC {
			t.Error()
		}
	})
	t.Run("Fail", func(t *testing.T) {
		setFailCondOfNetInfo()
		_, err := network.GetMACAddress()
		if err == nil {
			t.Error()
		}
	})
}
func TestAppendSubscriber(t *testing.T) {
	netInfo.ipChans = nil

	network.AppendSubscriber()

	if len(netInfo.ipChans) != 1 {
		t.Error()
	}
}

func MockGetNetworkInfo() {
	isCalledGetNetInfo = true
	log.Println("MockGetNetworkInfo")

	wait.Done()
}

func TestSuccessSubAddrChange(t *testing.T) {
	ctrl := gomock.NewController(t)
	det := mocks.NewMockDetector(ctrl)

	isCalledGetNetInfo = false
	wait.Add(1)
	getNetworkInformationFP = MockGetNetworkInfo

	isNewConnection := make(chan bool, 1)

	det.EXPECT().AddrSubscribe(isNewConnection)

	go subAddrChange(isNewConnection)

	// @Note : Invoke case of detecting network change
	isNewConnection <- true

	wait.Wait()
}

// @Note : From this line, TC is related with networkInformation struct
func TestNotify(t *testing.T) {
	netInfo.addrInfos = make([]addrInformation, 1)
	netInfo.addrInfos[0].ipv4 = TESTNEWIP
	netInfo.addrInfos[0].isWired = true
	netInfo.ipChans = nil

	ConnectivityChan := make(chan []net.IP, 1)
	netInfo.ipChans = append(netInfo.ipChans, ConnectivityChan)

	var newIP []net.IP
	wait.Add(1)

	go func() {
		newIP = <-ConnectivityChan
		wait.Done()
	}()

	netInfo.Notify(TESTNEWIPS)
	wait.Wait()

	if reflect.DeepEqual(newIP, TESTNEWIPS) != true {
		t.Error()
	}
}

func TestSuccessGetIP(t *testing.T) {
	netInfo.addrInfos = make([]addrInformation, 2)
	for _, addInfo := range netInfo.addrInfos {
		addInfo.macAddr = TESTMAC
	}

	netInfo.addrInfos[0].isWired = true
	netInfo.addrInfos[1].isWired = false

	netInfo.addrInfos[0].ipv4 = TESTNEWWIREDIP
	netInfo.addrInfos[1].ipv4 = TESTNEWIP

	// @Note : Get Wired IP
	if reflect.DeepEqual(netInfo.GetIP(), TESTNEWWIREDIP) != true {
		t.Error()
	}
}

func TestSuccessGetIPs(t *testing.T) {
	netInfo.addrInfos = make([]addrInformation, 2)
	for _, addInfo := range netInfo.addrInfos {
		addInfo.macAddr = TESTMAC
	}

	netInfo.addrInfos[0].isWired = true
	netInfo.addrInfos[1].isWired = false

	netInfo.addrInfos[0].ipv4 = TESTNEWWIREDIP
	netInfo.addrInfos[1].ipv4 = TESTNEWIP

	if reflect.DeepEqual(
		netInfo.GetIPs(), []net.IP{TESTNEWWIREDIP, TESTNEWIP}) != true {
		t.Error()
	}
}

func TestGetVirtualIP(t *testing.T) {

	t.Run("FailNetInfo", func(t *testing.T) {
		setFailCondOfNetInfo()
		_, err := network.GetVirtualIP()
		if err == nil {
			t.Error("Expected error")
			return
		}
	})
	t.Run("AddrNotFound", func(t *testing.T) {
		setPassCondOfNetInfo()
		_, err := network.GetVirtualIP()
		if err == nil {
			t.Error("Expected error")
			return
		}
	})
	t.Run("Success", func(t *testing.T) {
		setPassCondOfNetInfo()
		netInfo.addrInfos = make([]addrInformation, 1)
		netInfo.addrInfos[0].isVirtual = true
		netInfo.addrInfos[0].ipv4 = TESTNEWIP
		_, err := network.GetVirtualIP()
		if err != nil {
			t.Error("error not expected")
			return
		}
	})
}
