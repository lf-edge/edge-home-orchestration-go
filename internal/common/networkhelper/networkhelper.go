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

// Package networkhelper checks the status of network interfaces and let subscribers know if it is updated
package networkhelper

import (
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/errormsg"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/errors"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/networkhelper/detector"
)

type networkImpl struct{}

var (
	networkIns              networkImpl
	detectorIns             detector.Detector
	netInfo                 networkInformation
	getNetworkInformationFP func()
	log                     = logmgr.GetInstance()
)

const (
	virtualInterfaceName = "tun"
	virtualNetworkString = "virtual"
)

// Network gets the informations of network interfaces of local device
type Network interface {
	StartNetwork()
	CheckConnectivity() error
	GetOutboundIP() (string, error)
	GetIPs() ([]string, error)
	GetMACAddress() (string, error)
	GetNetInterface() ([]net.Interface, error)
	AppendSubscriber() chan []net.IP
	GetVirtualIP() (string, error)
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

	netInfo.Notify(netInfo.GetIPs())

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

// GetOutboundIP returns IPv4 addresses
func (networkImpl) GetIPs() ([]string, error) {
	ipsStr := make([]string, 0)
	if netInfo.netError == nil {
		ips := netInfo.GetIPs()
		for _, ip := range ips {
			ipsStr = append(ipsStr, ip.String())
		}
	}
	return ipsStr, netInfo.netError
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

//GetVirtualIP returns Virtual IP of the device
func (networkImpl) GetVirtualIP() (string, error) {
	log.Println(logPrefix, "Virtual IP asked")
	if netInfo.netError == nil {
		for _, addrInfo := range netInfo.addrInfos {
			if addrInfo.isVirtual {
				log.Println(logPrefix, "returning", addrInfo.ipv4.String())
				return addrInfo.ipv4.String(), nil
			}
		}
	} else {
		return "", netInfo.netError
	}
	err := errors.NotFound{
		Message: "Virtual Network Not Found",
	}
	return "", err
}

// AppendSubscriber appends subscriber
func (networkImpl) AppendSubscriber() chan []net.IP {
	ipChan := make(chan []net.IP, 1)

	netInfo.ipChans = append(netInfo.ipChans, ipChan)

	return ipChan
}

func getNetworkInformation() {
	ifaces, _ := net.Interfaces()

	err := setAddrInfo(ifaces)
	if err != nil {
		return
	}
}

func setAddrInfo(ifaces []net.Interface) (err error) {
	netDirPathPrefix := "/sys/class/net/"

	if len(ifaces) == 0 {
		err = errors.NetworkError{
			Message: errormsg.ToString(errormsg.ErrorNoNetworkInterface)}
		netInfo.netError = err

		return
	}

	var filterIfaces []net.Interface
	var addrInfos []addrInformation
	for _, i := range ifaces {
		path, _ := filepath.EvalSymlinks(netDirPathPrefix + i.Name)
		if checkVirtualNet(path) && checkWiredNet(path) && !strings.Contains(path, virtualInterfaceName) {
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
				addrInfo.isVirtual = checkTunNet(path)

				log.Println(logPrefix, "addr", addr)
				addrInfo.isWired = !addrInfo.isVirtual

				addrInfos = append(addrInfos, addrInfo)
				filterIfaces = append(filterIfaces, i)
			}
		}
	}

	netInfo.netInterface = filterIfaces
	netInfo.addrInfos = addrInfos
	netInfo.netError = nil

	return
}

func subAddrChange(isNewConnection chan bool) {
	go detectorIns.AddrSubscribe(isNewConnection)
	for {
		select {
		// @Note : If network status is changed, need to update network information
		case connectionDetected := <-isNewConnection:
			if connectionDetected == true {
				getNetworkInformationFP()
				netInfo.Notify(netInfo.GetIPs())
			}
		}
	}
	//apply detectorIns.Done when normal termination
}

func checkWiredNet(path string) bool {
	if _, err := os.Stat(path + "/wireless"); os.IsNotExist(err) {
		return true
	}

	return false
}

func checkVirtualNet(path string) bool {
	return strings.Contains(path, virtualNetworkString)
}

func checkTunNet(path string) bool {
	return strings.Contains(path, virtualInterfaceName)
}

func (netInfo *networkInformation) Notify(ips []net.IP) {
	if len(netInfo.addrInfos) == 0 {
		return
	}

	for _, sub := range netInfo.ipChans {
		select {
		case sub <- ips:
		default:
			log.Println(logPrefix, "[notify] ", "subchan is not receiving")
		}
	}
}

func (netInfo *networkInformation) GetIP() (ipv4 net.IP) {
	for _, addrInfo := range netInfo.addrInfos {
		// @Note : ethernet network have a priority
		if addrInfo.isVirtual {
			continue
		}
		if addrInfo.isWired {
			return addrInfo.ipv4
		}

		ipv4 = addrInfo.ipv4
	}

	return ipv4
}

func (netInfo *networkInformation) GetIPs() []net.IP {
	ips := make([]net.IP, 0)
	for _, addrInfo := range netInfo.addrInfos {
		ips = append(ips, addrInfo.ipv4)
	}

	return ips
}
