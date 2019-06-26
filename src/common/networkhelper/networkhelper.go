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

// Package networkhelper checks the status of network interfaces and let subscribers know if it is updated
package networkhelper

import (
	//	"errors"
	"log"
	"net"

	"common/errormsg"
	"common/networkhelper/detector"
)

type networkImpl struct{}

var networkIns networkImpl
var detectorIns detector.Detector
var netInfo networkInformation

// Network gets the informations of network interfaces of local device
type Network interface {
	StartNetwork()
	CheckConnectivity() error
	GetOutboundIP() (string, error)
	GetMACAddress() (string, error)
	GetNetInterface() ([]net.Interface, error)
	AppendSubscriber() chan net.IP
}

func init() {
	detectorIns = detector.GetInstance()
}

// GetInstance returns the networkImpl struct
func GetInstance() Network {
	return networkIns
}

// StartNetwork finds the IPs and network interfaces of local device
func (networkImpl) StartNetwork() {
	getNetworkInformation()

	go subscribeAddrChg()

}

func subscribeAddrChg() {
	isNewConnection := make(chan bool, 1)
	go detectorIns.AddrSubscribe(isNewConnection)
	for {
		select {
		case ConnectionDetected := <-isNewConnection:
			if ConnectionDetected {
				getNetworkInformation()
			}
		}
	}
	//apply detectorIns.Done when normal termination
}

// CheckConnectivity returns nil when connected or error when disconnected
func (networkImpl) CheckConnectivity() error {
	return netInfo.netError
}

// GetOutboundIP returns IPv4 address
func (networkImpl) GetOutboundIP() (string, error) {
	if netInfo.netError == nil {
		return netInfo.ipv4, nil
	}
	return "", netInfo.netError
}

//GetMACAddress returns MAC address
func (networkImpl) GetMACAddress() (string, error) {
	if netInfo.netError == nil {
		return netInfo.macAddress, nil
	}
	return "", netInfo.netError
}

//GetNetInterface returns wl network interface
func (networkImpl) GetNetInterface() ([]net.Interface, error) {
	if netInfo.netError == nil {
		return netInfo.netInterface, nil
	}
	return nil, netInfo.netError
}

// AppendSubscriber appends subscriber
func (networkImpl) AppendSubscriber() chan net.IP {
	ipChan := make(chan net.IP, 1)
	netInfo.ipChans = append(netInfo.ipChans, ipChan)
	return ipChan
}

func getNetworkInformation() {
	ifaces, _ := net.Interfaces()
	err := getWifiInterfaceInfo(ifaces)
	if err != nil {
		return
	}
	netInfo.notify()
}

func getWifiInterfaceInfo(ifaces []net.Interface) (err error) {
	netInfo.netInterface = nil

	for _, iface := range ifaces {
		if len(iface.Name) == 0 || iface.Name[0:2] != "wl" {
			continue
		}
		addrs, errAddrs := iface.Addrs()
		if len(addrs) == 0 || errAddrs != nil {
			continue
		}
		netInfo.netInterface = []net.Interface{iface}
		netInfo.macAddress = iface.HardwareAddr.String()
		netInfo.ipv4, err = netInfo.getIPv4(addrs)
		if err != nil {
			continue
		}
		break
	}

	if len(netInfo.netInterface) == 0 {
		err = errormsg.ToError(errormsg.ErrorTurnOffWifi)
	}

	netInfo.netError = err
	return
}

func (netInfo *networkInformation) getIPv4(addrs []net.Addr) (string, error) {
	for _, address := range addrs {
		strAddr := address.String()
		strAddr = strAddr[0 : len(strAddr)-3]
		log.Println(">> ", strAddr)
		ip := net.ParseIP(strAddr).To4()
		if ip == nil || netInfo.ipv4 == strAddr {
			continue
		}
		return strAddr, nil
		// TODO : Two Wlan Connection
	}
	return "", errormsg.ToError(errormsg.ErrorDisconnectWifi)
}

func (netInfo *networkInformation) notify() {
	ipv4 := net.ParseIP(netInfo.ipv4)
	for _, sub := range netInfo.ipChans {
		select {
		case sub <- ipv4:
		default:
			log.Println(logPrefix, "[notify] ", "subchan is not receiving")
		}
	}
}
