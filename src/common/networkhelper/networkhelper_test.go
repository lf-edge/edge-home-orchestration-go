/*******************************************************************************
 * Copyright 2019 Samsung Electronics All Rights Reserved.
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
	"common/errormsg"
	"errors"
	"log"
	"net"
	"testing"
	"time"
)

var (
	TESTIPV4      = "11.111.111.11"
	TESTIPV6      = "fe80::f112:19a8:eca4:724f"
	TESTMAC       = "AB:CD:EF:GH:IJ:KL"
	TESTNETIFLIST []net.Interface
	TESTERR       error
	TESTNEWIP     net.IP
	TESTIFINDEX   = 1
)

type addrstr struct {
	name string
	addr string
}

func (str addrstr) Network() string {
	return str.name
}
func (str addrstr) String() string {
	return str.addr
}

func init() {
	TESTERR = errors.New("ERROR")
	TESTNETIF := net.Interface{Index: TESTIFINDEX}
	TESTNETIFLIST = append(TESTNETIFLIST, TESTNETIF)
	TESTNEWIP = net.ParseIP("22.222.222.22")
}

func TestGetIPv4(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setPass()
		err := errormsg.ToError(errormsg.ErrorDisconnectWifi)
		addrs1 := addrstr{
			name: "tcp",
			addr: TESTIPV4 + "/64"}
		addrs2 := addrstr{
			name: "tcp",
			addr: "22.222.222.22" + "/64"}
		var addrlist []net.Addr
		addrlist = append(addrlist, addrs1)
		addrlist = append(addrlist, addrs2)
		_, err = netInfo.getIPv4(addrlist)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		setFail()
		err := errormsg.ToError(errormsg.ErrorDisconnectWifi)
		addrs := addrstr{
			name: "tcp",
			addr: TESTIPV6 + "/64"}
		var addrlist []net.Addr
		addrlist = append(addrlist, addrs)
		_, err = netInfo.getIPv4(addrlist)
		if err == nil {
			t.Error(err)
		}
	})
}

func TestGetWifiInterfaceInfo(t *testing.T) {
	ifaces := []net.Interface{
		net.Interface{},
		net.Interface{},
	}
	err := getWifiInterfaceInfo(ifaces)
	if err == nil {
		t.Error()
	}

	// TODO check interface name and hardware address
	//	pInterface := []net.Interface{
	//		net.Interface{
	//			Index:        3,
	//			MTU:          1500,
	//			Name:         "wlxd8fee35ec830",
	//			HardwareAddr: []byte("d8:fe:e3:5e:c8:30"),
	//			Flags:        (net.FlagUp | net.FlagBroadcast | net.FlagMulticast),
	//		},
	//		net.Interface{},
	//	}
	//	log.Println(pInterface)
	//	err = getWifiInterfaceInfo(pInterface)
	//	if err != nil {
	//		t.Error()
	//	}
}

func TestGetNetInterface(t *testing.T) {
	network := GetInstance()

	t.Run("Success", func(t *testing.T) {
		setPass()
		netif, err := network.GetNetInterface()
		if err != nil {
			t.Error()
		}
		if len(netif) == 0 || netif[0].Index != TESTIFINDEX {
			t.Error()
		}
	})
	t.Run("Fail", func(t *testing.T) {
		setFail()
		_, err := network.GetNetInterface()
		if err == nil {
			t.Error()
		}
	})
	t.Log()
}

func TestCheckConnectivity(t *testing.T) {
	network := GetInstance()

	t.Run("Success", func(t *testing.T) {
		setPass()
		err := network.CheckConnectivity()
		if err != nil {
			t.Error()
		}
	})
	t.Run("Fail", func(t *testing.T) {
		setFail()
		err := network.CheckConnectivity()
		if err == nil {
			t.Error()
		}
	})
}
func TestGetOutboundIP(t *testing.T) {
	network := GetInstance()

	t.Run("Success", func(t *testing.T) {
		setPass()
		ip, err := network.GetOutboundIP()
		if err != nil {
			t.Error()
		}
		if ip != TESTIPV4 {
			t.Error()
		}
	})
	t.Run("Fail", func(t *testing.T) {
		setFail()
		_, err := network.GetOutboundIP()
		if err == nil {
			t.Error()
		}
	})
}

func TestGetMACAddress(t *testing.T) {
	network := GetInstance()

	t.Run("Success", func(t *testing.T) {
		setPass()
		mac, err := network.GetMACAddress()
		if err != nil {
			t.Error()
		}
		if mac != TESTMAC {
			t.Error()
		}
	})
	t.Run("Fail", func(t *testing.T) {
		setFail()
		_, err := network.GetMACAddress()
		if err == nil {
			t.Error()
		}
	})
}
func TestAppendSubscriber(t *testing.T) {
	network := GetInstance()

	log.Println(logPrefix, "[test channel]")
	netInfo.ipChans = nil
	network.AppendSubscriber()
	if len(netInfo.ipChans) != 1 {
		t.Error()
	}
}

func TestNotify(t *testing.T) {
	netInfo.ipv4 = TESTNEWIP.String()
	netInfo.ipChans = nil

	ConnectivityChan := make(chan net.IP, 1)
	netInfo.ipChans = append(netInfo.ipChans, ConnectivityChan)

	var newip net.IP
	go func() {
		newip = <-ConnectivityChan
		return
	}()
	netInfo.notify()
	time.Sleep(2 * time.Second)

	if newip.String() != TESTNEWIP.String() {
		t.Error()
	}
}

func setPass() {
	netInfo.ipv4 = TESTIPV4
	netInfo.macAddress = TESTMAC
	netInfo.netInterface = TESTNETIFLIST
	netInfo.netError = nil
}

func setFail() {
	netInfo.ipv4 = ""
	netInfo.macAddress = ""
	netInfo.netInterface = nil
	netInfo.netError = TESTERR
}
