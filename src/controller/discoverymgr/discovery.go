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
	"errors"
	"io/ioutil"
	"log"
	"net"
	"time"

	errormsg "common/errormsg"
	networkhelper "common/networkhelper"
	wrapper "controller/discoverymgr/wrapper"

	uuid "github.com/satori/go.uuid"
)

// Discovery is the interface implementedy by all discovery functions
type Discovery interface {
	StartDiscovery(UUIDpath string, platform string, executionType string) error
	StopDiscovery()
	GetDeviceList() (ExportDeviceMap, error)
	GetDeviceIPListWithService(targetService string) ([]string, error)
	GetDeviceListWithService(targetService string) (ExportDeviceMap, error)
	GetDeviceWithID(ID string) (ExportDeviceMap, error)
	DeleteDeviceWithIP(targetIP string)
	DeleteDeviceWithID(ID string)
	AddNewServiceName(serviceName string) error
	RemoveServiceName(serviceName string) error
	ResetServiceName()
}

type discoveryImpl struct{}

var (
	discoverymgrInfo discoverymgrInformation

	discoveryIns discoveryImpl
	networkIns   networkhelper.Network
)

func init() {
	discoverymgrInfo.wrapperIns = wrapper.GetZeroconfImpl()
	discoverymgrInfo.orchestrationMap = make(OrchestrationMap)
	discoverymgrInfo.shutdownChan = make(chan struct{})

	networkIns = networkhelper.GetInstance()
}

// GetInstance returns discovery instaance
func GetInstance() Discovery {
	return discoveryIns
}

// InitDiscovery starts server for network registration and do orchestration discovery activity
func (discoveryImpl) StartDiscovery(UUIDpath string, platform string, executionType string) (err error) {
	networkIns.StartNetwork()

	setDeviceInfo(platform, executionType)

	UUIDStr, err := getDeviceID(UUIDpath)
	if err != nil {
		log.Print(logPrefix, "[StartDiscovery]", "UUID ", UUIDStr, " is Temporary")
	}

	// NOTE : startServer blocks until server is registered
	startServer(UUIDStr)

	go detectNetworkChgRoutine()

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
	discoverymgrInfo.wrapperIns.Shutdown()
}

// GetDeviceList returns orchestration device info list
func (discoveryImpl) GetDeviceList() (ExportDeviceMap, error) {
	discoverymgrInfo.mapMTX.Lock()
	defer discoverymgrInfo.mapMTX.Unlock()

	ret := make(ExportDeviceMap)
	for key, value := range discoverymgrInfo.orchestrationMap {
		ret[key] = *value
	}
	if len(ret) == 0 {
		err := errormsg.ToError(errormsg.ErrorNoDeviceReturn)
		return nil, err
	}

	return ret, nil
}

// GetDeviceListWithService returns orchestration deviceIP list using service application name
func (discoveryImpl) GetDeviceIPListWithService(targetService string) ([]string, error) {
	discoverymgrInfo.mapMTX.Lock()
	defer discoverymgrInfo.mapMTX.Unlock()

	var ret []string
	var err error
	for _, value := range discoverymgrInfo.orchestrationMap {
		for _, val := range value.ServiceList {
			if val == targetService {
				ret = append(ret, value.IPv4...)
			}
		}
	}
	if len(ret) == 0 {
		err = errormsg.ToError(errormsg.ErrorNoDeviceReturn)
		return nil, err
	}

	return ret, nil
}

// GetDeviceListWithService returns device info list using service application name
func (discoveryImpl) GetDeviceListWithService(targetService string) (ExportDeviceMap, error) {
	discoverymgrInfo.mapMTX.Lock()
	defer discoverymgrInfo.mapMTX.Unlock()

	ret := make(ExportDeviceMap)
	var err error
	for key, value := range discoverymgrInfo.orchestrationMap {
		for _, val := range value.ServiceList {
			if val == targetService {
				ret[key] = *value
			}
		}
	}

	if len(ret) == 0 {
		err = errormsg.ToError(errormsg.ErrorNoDeviceReturn)
		return nil, err
	}

	return ret, nil
}

// GetDeviceWithID returns device info using deviceID
func (discoveryImpl) GetDeviceWithID(ID string) (ExportDeviceMap, error) {
	discoverymgrInfo.mapMTX.Lock()
	defer discoverymgrInfo.mapMTX.Unlock()

	ret := make(ExportDeviceMap)
	var err error

	if value, ok := discoverymgrInfo.orchestrationMap[ID]; ok {
		ret[ID] = *value
		return ret, nil
	}
	err = errormsg.ToError(errormsg.ErrorNoDeviceReturn)
	return nil, err
}

// DeleteDevice deletes device info using deviceIP
func (discoveryImpl) DeleteDeviceWithIP(targetIP string) {
	discoverymgrInfo.mapMTX.Lock()
	defer discoverymgrInfo.mapMTX.Unlock()

	for key, val := range discoverymgrInfo.orchestrationMap {
		for _, ipv4 := range val.IPv4 {
			if discoverymgrInfo.deviceID == key {
				continue
			}
			if targetIP == ipv4 {
				delete(discoverymgrInfo.orchestrationMap, key)
				return
			}
		}
	}
	log.Println(logPrefix, "[DeleteDeviceWithIP]", "Cannot Delete Self")
}

// DeleteDevice delete device using deviceID
func (discoveryImpl) DeleteDeviceWithID(ID string) {
	discoverymgrInfo.mapMTX.Lock()
	defer discoverymgrInfo.mapMTX.Unlock()

	if discoverymgrInfo.deviceID == ID {
		log.Println(logPrefix, "[DeleteDeviceWithID]", "Cannot Delete Self")
		return
	}

	delete(discoverymgrInfo.orchestrationMap, ID)
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

	serverTXT := discoverymgrInfo.wrapperIns.GetText()
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
	discoverymgrInfo.orchestrationMap[discoverymgrInfo.deviceID].ServiceList = nil
	var serverTXT []string
	serverTXT = append(serverTXT, discoverymgrInfo.orchestrationMap[discoverymgrInfo.deviceID].ExecutionType)
	serverTXT = append(serverTXT, discoverymgrInfo.orchestrationMap[discoverymgrInfo.deviceID].Platform)
	setNewServiceList(serverTXT)
}

func detectNetworkChgRoutine() {
	ip := networkIns.AppendSubscriber()

	for {
		select {
		case <-discoverymgrInfo.shutdownChan:
			return
		case newIP := <-ip:
			var ips []net.IP
			ips = append(ips, newIP)
			err := serverPresenceChecker()
			if err != nil {
				continue
			}
			discoverymgrInfo.wrapperIns.ResetServer(ips)
		}
	}
}

func getDeviceID(UUIDPath string) (UUIDstr string, err error) {

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

func setDeviceInfo(platform string, executionType string) {
	log.Println(logPrefix, "Platform::", platform, " OnboardType::", executionType)
	discoverymgrInfo.platform = platform
	discoverymgrInfo.executionType = executionType
}

func startServer(deviceUUID string) {

	deviceDetectionRoutine()

	deviceID, hostName, Text := setDeviceArgument(deviceUUID)
	hostIPAddr, netIface := setNetwotkArgument()
	var myDeviceEntity wrapper.Entity

	for {
		var err error
		myDeviceEntity, err = discoverymgrInfo.wrapperIns.RegisterProxy(
			deviceID, serviceType, domain, servicePort, hostName, hostIPAddr, Text, netIface)
		if err != nil {
			log.Println(logPrefix, "[startServer]", "Register Server Failed : ", err)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	deviceID, deviceInfo := convertwrappertoDB(myDeviceEntity)
	newDeviceHandler(deviceID, deviceInfo)
	discoverymgrInfo.deviceID = deviceID

	return
}

func setDeviceArgument(deviceUUID string) (deviceID string, hostName string, Text []string) {
	deviceID = "edge-orchestration-" + deviceUUID
	hostName = "edge-" + deviceUUID

	Text = append(Text, discoverymgrInfo.platform)
	Text = append(Text, discoverymgrInfo.executionType)
	return
}

func setNetwotkArgument() (hostIPAddr []string, netIface []net.Interface) {
	var ip string
	var err error
	// TODO : change to channel
	for {
		ip, err = networkIns.GetOutboundIP()
		if len(ip) != 0 {
			break
		}
		log.Println(logPrefix, errormsg.ToString(err))
		time.Sleep(1 * time.Second)
	}
	log.Println(logPrefix + " ip : " + ip)

	hostIPAddr = append(hostIPAddr, ip)

	netIface, _ = networkIns.GetNetInterface()

	return
}

func deviceDetectionRoutine() {
	go func() {
		subchan, err := discoverymgrInfo.wrapperIns.GetSubscriberChan()
		if err != nil {
			log.Println(logPrefix, err)
			return
		}
		for {
			select {
			case <-discoverymgrInfo.shutdownChan:
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
				deviceID, deviceInfo := convertwrappertoDB(*data)

				// @TODO : check locking logic
				discoverymgrInfo.mapMTX.Lock()
				_, isPresent := discoverymgrInfo.orchestrationMap[deviceID]
				discoverymgrInfo.mapMTX.Unlock()

				if isPresent {
					updateInfoHandler(deviceID, deviceInfo)
					continue
				}

				newDeviceHandler(deviceID, deviceInfo)

				// case default:
				//resource return
			}
		}
	}()
}

func serverPresenceChecker() error {
	if len(discoverymgrInfo.deviceID) == 0 {
		return errors.New("no server initiated yet")
	}
	return nil
}

func shutdownDiscoverymgr() {
	discoverymgrInfo.deviceID = ""
	discoverymgrInfo.orchestrationMap = make(OrchestrationMap)
	close(discoverymgrInfo.shutdownChan)
}

func serviceNameChecker(serviceName string) error {
	if serviceName == "" {
		return errors.New("no argument")
	}
	if serviceName == discoverymgrInfo.platform || serviceName == discoverymgrInfo.executionType {
		return errors.New("cannot change fixed field")
	}
	return nil
}

func appendServiceToTXT(serviceName string) ([]string, error) {
	serverTXT := discoverymgrInfo.wrapperIns.GetText()
	for _, str := range serverTXT {
		if str == serviceName {
			return nil, errors.New("service name duplicated")
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
		return errors.New("TXT Size is Too much for mDNS TXT - 400B")
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
		err = errors.New("no service found")
	}
	return
}

func setNewServiceList(serverTXT []string) {
	if len(serverTXT) > 2 {
		newServiceList := serverTXT[2:]
		discoverymgrInfo.orchestrationMap[discoverymgrInfo.deviceID].ServiceList = newServiceList
	}
	discoverymgrInfo.wrapperIns.SetText(serverTXT)
}

func convertwrappertoDB(entity wrapper.Entity) (string, *OrchestrationInformation) {
	data := entity.OrchestrationInfo
	newdata := OrchestrationInformation{
		IPv4:          data.IPv4,
		Platform:      data.Platform,
		ExecutionType: data.ExecutionType,
		ServiceList:   data.ServiceList}

	return entity.DeviceID, &newdata
}

// updateInfoHandler applies ip/text changes
func updateInfoHandler(key string, data *OrchestrationInformation) {
	// log.Println(logPrefix, "[updateInfoHandler]", key, data)
	discoverymgrInfo.mapMTX.Lock()
	defer discoverymgrInfo.mapMTX.Unlock()

	if data.IPv4 != nil {
		discoverymgrInfo.orchestrationMap[key].IPv4 = data.IPv4
	}

	if data.ServiceList != nil {
		discoverymgrInfo.orchestrationMap[key].ServiceList = data.ServiceList
	}

	return
}

// newDeviceHandler updates key or adds new db
func newDeviceHandler(key string, data *OrchestrationInformation) {
	// log.Println(logPrefix, "[newDeviceHandler]", key, data)
	discoverymgrInfo.mapMTX.Lock()
	defer discoverymgrInfo.mapMTX.Unlock()

	if data.IPv4 == nil {
		return
	}
	discoverymgrInfo.orchestrationMap[key] = data
}

// DeleteDevice deletes device info by key
func deleteDevice(key string) {
	log.Println(logPrefix, "[deleteDevice]", key)
	discoverymgrInfo.mapMTX.Lock()
	defer discoverymgrInfo.mapMTX.Unlock()

	delete(discoverymgrInfo.orchestrationMap, key)
}

// ClearMap makes map empty and only leaves my device info
func clearMap() {
	log.Println(logPrefix, "[clearMap]")
	discoverymgrInfo.mapMTX.Lock()
	defer discoverymgrInfo.mapMTX.Unlock()
	myDeviceData := discoverymgrInfo.orchestrationMap[discoverymgrInfo.deviceID]
	discoverymgrInfo.orchestrationMap = make(OrchestrationMap)
	discoverymgrInfo.orchestrationMap[discoverymgrInfo.deviceID] = myDeviceData
}

// activeDiscovery calls advertise function of Zeroconf
// Does Not Clear Map
func activeDiscovery() {
	discoverymgrInfo.wrapperIns.Advertise()
}

// resetServer calls advertise function of Zeroconf
// It Clears Map
func resetServer(ips []net.IP) {
	discoverymgrInfo.wrapperIns.ResetServer(ips)
}
