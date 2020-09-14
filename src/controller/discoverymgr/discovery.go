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

package discoverymgr

import (
	"io/ioutil"
	"log"
	"net"
	"time"
	"reflect"

	errors "common/errors"
	networkhelper "common/networkhelper"
	wrapper "controller/discoverymgr/wrapper"

	configurationdb "db/bolt/configuration"
	networkdb "db/bolt/network"
	systemdb "db/bolt/system"

	uuid "github.com/satori/go.uuid"
)

// Discovery is the interface implementedy by all discovery functions
type Discovery interface {
	StartDiscovery(UUIDpath string, platform string, executionType string) error
	StopDiscovery()
	DeleteDeviceWithIP(targetIP string)
	DeleteDeviceWithID(ID string)
	AddNewServiceName(serviceName string) error
	RemoveServiceName(serviceName string) error
	ResetServiceName()
}

type discoveryImpl struct{}

var (
	discoveryIns discoveryImpl
	networkIns   networkhelper.Network
)

func init() {
	wrapperIns = wrapper.GetZeroconfImpl()
	shutdownChan = make(chan struct{})

	networkIns = networkhelper.GetInstance()

	sysQuery = systemdb.Query{}
	confQuery = configurationdb.Query{}
	netQuery = networkdb.Query{}
}

// GetInstance returns discovery instaance
func GetInstance() Discovery {
	return discoveryIns
}

// InitDiscovery starts server for network registration and do orchestration discovery activity
func (discoveryImpl) StartDiscovery(UUIDpath string, platform string, executionType string) (err error) {
	networkIns.StartNetwork()

	UUIDStr, err := setDeviceID(UUIDpath)
	if err != nil {
		log.Print(logPrefix, "[StartDiscovery]", "UUID ", UUIDStr, " is Temporary")
	}

	// NOTE : startServer blocks until server is registered
	startServer(UUIDStr, platform, executionType)

	go detectNetworkChgRoutine()

	go func() {
		for {
			time.Sleep(time.Minute)
			activeDiscovery()
		}
	}()

	return
}

// StopDiscovery shutdowns server
// Todo : Set Normal Termination Function For Each Platform / Execution Type
func (discoveryImpl) StopDiscovery() {
	err := serverPresenceChecker()
	if err != nil {
		log.Println(logPrefix, "[StopDiscovery]", err)
		return
	}
	shutdownDiscoverymgr()
	wrapperIns.Shutdown()
}

// DeleteDevice deletes device info using deviceIP
func (discoveryImpl) DeleteDeviceWithIP(targetIP string) {
	// @TODO Delete device with ip in DB
}

// DeleteDevice delete device using deviceID
func (discoveryImpl) DeleteDeviceWithID(ID string) {
	// @Note Delete device with id in DB
	deviceID, err := getDeviceID()
	if err != nil {
		log.Println(err.Error())
		return
	}

	if deviceID != ID {
		deleteDevice(ID)
	}
}

// AddNewServiceName sets text field of mdns message with service application name
func (discoveryImpl) AddNewServiceName(serviceName string) error {
	err := serviceNameChecker(serviceName)
	if err != nil {
		return err
	}

	err = serverPresenceChecker()
	if err != nil {
		return err
	}

	serverTXT, err := appendServiceToTXT(serviceName)
	if err != nil {
		return err
	}

	err = mdnsTXTSizeChecker(serverTXT)
	if err != nil {
		return err
	}

	setNewServiceList(serverTXT)

	return nil
}

// RemoveServiceName removes text field of mdns message with service application name
func (discoveryImpl) RemoveServiceName(serviceName string) error {
	err := serviceNameChecker(serviceName)
	if err != nil {
		return err
	}

	err = serverPresenceChecker()
	if err != nil {
		return err
	}

	serverTXT := wrapperIns.GetText()

	idxToDel, err := getIndexToDelete(serverTXT, serviceName)
	if err != nil {
		return err
	}
	serverTXT = append(serverTXT[:idxToDel], serverTXT[idxToDel+1:]...)
	setNewServiceList(serverTXT)

	return nil
}

// ResetServiceName resets text field of mdns message
func (discoveryImpl) ResetServiceName() {
	err := serverPresenceChecker()
	if err != nil {
		log.Println(logPrefix, "[ResetServiceName]", err)
		return
	}

	deviceID, err := getDeviceID()
	if err != nil {
		return
	}

	confItem, err := confQuery.Get(deviceID)
	if err != nil {
		log.Println(err.Error())
		return
	}
	confItem.Services = nil
	setConfigurationDB(confItem)

	var serverTXT []string
	serverTXT = append(serverTXT, confItem.ExecType)
	serverTXT = append(serverTXT, confItem.Platform)

	setNewServiceList(serverTXT)
}

func detectNetworkChgRoutine() {
	ips := networkIns.AppendSubscriber()

	for {
		select {
		case <-shutdownChan:
			return
		case latestIPs := <-ips:
			id, err := getDeviceID()
			if err != nil {
				continue
			}

			var strIPs []string
			for _, ip := range latestIPs {
				strIPs = append(strIPs, ip.To4().String())
			}

			netInfo := networkdb.NetworkInfo{ID: id, IPv4: strIPs}
			setNetworkDB(netInfo)

			err = serverPresenceChecker()
			if err != nil {
				continue
			}

			wrapperIns.ResetServer(latestIPs)
		}
	}
}

func setDeviceID(UUIDPath string) (UUIDstr string, err error) {

	UUIDv4, err := ioutil.ReadFile(UUIDPath)

	if err != nil {
		log.Println(logPrefix, "No saved UUID : ", err)
		UUIDv4New := uuid.NewV4()

		UUIDstr = UUIDv4New.String()

		err = ioutil.WriteFile(UUIDPath, []byte(UUIDstr), 0644)
		if err != nil {
			log.Println(logPrefix, "Error Write UUID : ", err)
		}
	} else {
		UUIDstr = string(UUIDv4)
	}
	log.Println(logPrefix, "UUID : ", UUIDstr)
	return UUIDstr, err
}

func getDeviceID() (id string, err error) {
	id, err = getSystemDB(systemdb.ID)
	if err != nil {
		log.Println(err.Error())
	}

	return
}

func getPlatform() (platform string, err error) {
	platform, err = getSystemDB(systemdb.Platform)
	if err != nil {
		log.Println(err.Error())
	}

	return
}

func getExecType() (execType string, err error) {
	execType, err = getSystemDB(systemdb.ExecType)
	if err != nil {
		log.Println(err.Error())
	}

	return
}

func startServer(deviceUUID string, platform string, executionType string) {
	deviceDetectionRoutine()

	deviceID, hostName, Text := setDeviceArgument(deviceUUID, platform, executionType)

	// @Note store system information(id, platform and execution type) to system db
	setSystemDB(deviceID, platform, executionType)

	hostIPAddr, netIface := setNetwotkArgument()
	var myDeviceEntity wrapper.Entity

	for {
		var err error
		myDeviceEntity, err = wrapperIns.RegisterProxy(
			deviceID, serviceType, domain, servicePort, hostName, hostIPAddr, Text, netIface)
		if err != nil {
			log.Println(logPrefix, "[startServer]", "Register Server Failed : ", err)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	// Set Configuration Information to configuration DB
	_, confInfo, netInfo := convertToDBInfo(myDeviceEntity)

	setConfigurationDB(confInfo)
	setNetworkDB(netInfo)

	return
}

func setDeviceArgument(deviceUUID string, platform string, executionType string) (deviceID string, hostName string, Text []string) {
	deviceID = "edge-orchestration-" + deviceUUID
	hostName = "edge-" + deviceUUID

	Text = append(Text, platform)
	Text = append(Text, executionType)
	return
}

func setNetwotkArgument() (hostIPAddr []string, netIface []net.Interface) {
	for {
		hostIPAddr, _ = networkIns.GetIPs()
		if len(hostIPAddr) != 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	log.Println(logPrefix, hostIPAddr)
	netIface, _ = networkIns.GetNetInterface()

	return
}

func deviceDetectionRoutine() {
	go func() {
		subchan, err := wrapperIns.GetSubscriberChan()
		if err != nil {
			log.Println(logPrefix, err)
			return
		}
		for {
			select {
			case <-shutdownChan:
				log.Println(logPrefix, "[deviceDetectionRoutine]", "Shutdown")
				return
			case data := <-subchan:
				// log.Println(logPrefix, "[detectDevice] ", data)
				if data == nil {
					clearMap()
					continue
				}
				if data.TTL == 0 {
					deleteDevice(data.DeviceID)
					continue
				}

				_, confInfo, netInfo := convertToDBInfo(*data)

				log.Printf("[deviceDetectionRoutine] %s", data.DeviceID)
				log.Printf("[deviceDetectionRoutine] confInfo    : ExecType(%s), Platform(%s)", confInfo.ExecType, confInfo.Platform)
				log.Printf("[deviceDetectionRoutine] netInfo     : IPv4(%s), RTT(%v)", netInfo.IPv4, netInfo.RTT)
				log.Printf("[deviceDetectionRoutine] serviceInfo : Services(%v)", confInfo.Services)
				log.Printf("")

				var info networkdb.NetworkInfo
				info, err = getNetworkDB(netInfo.ID)

				if err != nil || !reflect.DeepEqual(netInfo.IPv4, info.IPv4) {
					setNetworkDB(netInfo)
				}

				// @Note Is it need to call Update API?
				setConfigurationDB(confInfo)
			}
		}
	}()
}

func serverPresenceChecker() error {
	_, err := getDeviceID()
	if err != nil {
		return errors.SystemError{Message: "no server initiated yet"}
	}

	return nil
}

func shutdownDiscoverymgr() {
	close(shutdownChan)
}

func serviceNameChecker(serviceName string) error {
	if serviceName == "" {
		return errors.InvalidParam{Message: "no argument"}
	}

	platform, _ := getPlatform()
	executionType, _ := getExecType()

	if serviceName == platform || serviceName == executionType {
		return errors.InvalidParam{Message: "cannot change fixed field"}
	}

	return nil
}

func appendServiceToTXT(serviceName string) ([]string, error) {
	serverTXT := wrapperIns.GetText()
	for _, str := range serverTXT {
		if str == serviceName {
			return nil, errors.InvalidParam{Message: "service name duplicated"}
		}
	}
	serverTXT = append(serverTXT, serviceName)
	return serverTXT, nil
}

func mdnsTXTSizeChecker(serverTXT []string) error {
	var TXTSize int
	for _, str := range serverTXT {
		TXTSize += len(str)
	}
	log.Println(logPrefix, "[mdnsTXTSizeChecker] size :: ", TXTSize, " Bytes")
	if TXTSize > maxTXTSize {
		return errors.InvalidParam{Message: "TXT Size is Too much for mDNS TXT - 400B"}
	}
	return nil
}

func getIndexToDelete(serverTXT []string, serviceName string) (idxToDel int, err error) {
	idxToDel = -1
	for i, str := range serverTXT {
		if str == serviceName {
			idxToDel = i
			break
		}
	}

	if idxToDel == -1 {
		err = errors.SystemError{Message: "no service found"}
	}
	return
}

func setNewServiceList(serverTXT []string) {
	// if len(serverTXT) > 2 {
	newServiceList := serverTXT[2:]

	deviceID, err := getDeviceID()
	if err != nil {
		return
	}

	confItem, err := confQuery.Get(deviceID)
	if err != nil {
		log.Println(err.Error())
		return
	}
	confItem.Services = newServiceList
	setConfigurationDB(confItem)

	wrapperIns.SetText(serverTXT)
}

// ClearMap makes map empty and only leaves my device info
func clearMap() {
	log.Println(logPrefix, "[clearMap]")

	confItems, err := confQuery.GetList()
	if err != nil {
		log.Println(logPrefix, err.Error())
		return
	}

	deviceID, err := getDeviceID()
	if err != nil {
		return
	}
	for _, confItem := range confItems {
		id := confItem.ID

		if id != deviceID {
			deleteDevice(id)
		}
	}
}

func convertToDBInfo(entity wrapper.Entity) (string, configurationdb.Configuration, networkdb.NetworkInfo) {
	data := entity.OrchestrationInfo

	confInfo := configurationdb.Configuration{}
	netInfo := networkdb.NetworkInfo{}

	confInfo.ID = entity.DeviceID
	confInfo.ExecType = data.ExecutionType
	confInfo.Platform = data.Platform
	confInfo.Services = data.ServiceList

	netInfo.ID = entity.DeviceID
	netInfo.IPv4 = data.IPv4

	return entity.DeviceID, confInfo, netInfo
}

func setSystemDB(id string, platform string, execType string) {
	sysInfo := systemdb.SystemInfo{Name: systemdb.ID, Value: id}
	err := sysQuery.Set(sysInfo)
	if err != nil {
		log.Println(logPrefix, err.Error())
	}

	sysInfo = systemdb.SystemInfo{Name: systemdb.Platform, Value: platform}
	err = sysQuery.Set(sysInfo)
	if err != nil {
		log.Println(logPrefix, err.Error())
	}

	sysInfo = systemdb.SystemInfo{Name: systemdb.ExecType, Value: execType}
	err = sysQuery.Set(sysInfo)
	if err != nil {
		log.Println(logPrefix, err.Error())
	}
}

func setConfigurationDB(confInfo configurationdb.Configuration) {
	err := confQuery.Set(confInfo)
	if err != nil {
		log.Println(logPrefix, err.Error())
	}
}

func setNetworkDB(netInfo networkdb.NetworkInfo) {
	err := netQuery.Set(netInfo)
	if err != nil {
		log.Println(logPrefix, err.Error())
	}
}

func getSystemDB(name string) (string, error) {
	sysInfo, err := sysQuery.Get(name)
	if err != nil {
		log.Println(logPrefix, err.Error())
		return "", err
	}

	return sysInfo.Value, err
}

func getNetworkDB(id string) (networkdb.NetworkInfo, error) {
	netInfo, err := netQuery.Get(id)
	if err != nil {
		log.Println(logPrefix, err.Error())
	}

	return netInfo, err
}

// DeleteDevice deletes device info by key
func deleteDevice(deviceID string) {
	log.Println(logPrefix, "[deleteDevice]", deviceID)
	err := confQuery.Delete(deviceID)
	if err != nil {
		log.Println(err.Error())
	}

	err = netQuery.Delete(deviceID)
	if err != nil {
		log.Println(err.Error())
	}
}

// activeDiscovery calls advertise function of Zeroconf
// Does Not Clear Map
func activeDiscovery() {
	log.Printf("[discoverymgr] activeDiscovery!!!")
	wrapperIns.Advertise()
}

// resetServer calls advertise function of Zeroconf
// It Clears Map
func resetServer(ips []net.IP) {
	wrapperIns.ResetServer(ips)
}
