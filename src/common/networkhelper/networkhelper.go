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
	"os"
	"path/filepath"
	"strings"

	"common/networkhelper/detector"
)

type networkImpl struct{}

var networkIns networkImpl
var detectorIns detector.Detector
var netInfo networkInformation
var getNetworkInformationFP func()

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
	getNetworkInformationFP = getNetworkInformation
}

// GetInstance returns the networkImpl struct
func GetInstance() Network {
	return networkIns
}

// StartNetwork finds the IPs and network interfaces of local device
func (networkImpl) StartNetwork() {
	getNetworkInformationFP()

	isNewConnection := make(chan bool, 1)
	go subAddrChange(isNewConnection)
}

// CheckConnectivity returns nil when connected or error when disconnected
func (networkImpl) CheckConnectivity() error {
	return netInfo.netError
}

// GetOutboundIP returns IPv4 address
func (networkImpl) GetOutboundIP() (string, error) {
	if netInfo.netError == nil {
		ip := netInfo.GetIP()
		return ip.String(), nil
	}
	return "", netInfo.netError
}

//GetMACAddress returns MAC address
func (networkImpl) GetMACAddress() (string, error) {
	if netInfo.netError == nil {
		return netInfo.addrInfos[0].macAddr, nil
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

	err := setAddrInfo(ifaces)
	if err != nil {
		return
	}

	netInfo.Notify()
}

func setAddrInfo(ifaces []net.Interface) (err error) {
	netDirPathPrefix := "/sys/class/net/"

	addrInfos := make([]addrInformation, 1)
	for _, i := range ifaces {
		path, _ := filepath.EvalSymlinks(netDirPathPrefix + i.Name)
		if checkVirtualNet(path) {
			continue
		}

		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			ipnet, isPresent := addr.(*net.IPNet)
			if isPresent == false {
				continue
			}

			if ipnet.IP.To4() != nil {
				var addrInfo addrInformation

				addrInfo.ipv4 = ipnet.IP.To4()
				addrInfo.macAddr = i.HardwareAddr.String()
				addrInfo.isWired = checkWiredNet(netDirPathPrefix + i.Name)

				addrInfos = append(addrInfos, addrInfo)
			}
		}
	}

	netInfo.netInterface = ifaces
	netInfo.addrInfos = addrInfos

	return
}

func subAddrChange(isNewConnection chan bool) {
	go detectorIns.AddrSubscribe(isNewConnection)
	for {
		select {
		case ConnectionDetected := <-isNewConnection:
			if ConnectionDetected {
				getNetworkInformationFP()
			}
		}
	}
	//apply detectorIns.Done when normal termination
}

func checkWiredNet(path string) (isWired bool) {
	if _, err := os.Stat(path + "/wireless"); os.IsNotExist(err) {
		isWired = true
	}
	log.Println(path, isWired)

	return
}

func checkVirtualNet(path string) bool {
	return strings.Contains(path, "virtual")
}

func (netInfo *networkInformation) Notify() {
	if len(netInfo.addrInfos) == 0 {
		return
	}

	ipv4 := netInfo.GetIP()

	for _, sub := range netInfo.ipChans {
		select {
		case sub <- ipv4:
		default:
			log.Println(logPrefix, "[notify] ", "subchan is not receiving")
		}
	}
}

func (netInfo *networkInformation) GetIP() (ipv4 net.IP) {
	for _, addrInfo := range netInfo.addrInfos {
		// @Note : ethernet network have a prior
		if addrInfo.isWired {
			return addrInfo.ipv4
		}

		ipv4 = addrInfo.ipv4
	}

	return ipv4
}
