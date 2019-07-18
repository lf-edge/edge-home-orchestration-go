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
	"strings"
	"time"

	errormsg "common/errormsg"
	errors "common/errors"
	networkhelper "common/networkhelper"
	wrapper "controller/discoverymgr/wrapper"

	configurationdb "db/bolt/configuration"
	networkdb "db/bolt/network"
	servicedb "db/bolt/service"
	systemdb "db/bolt/system"

	uuid "github.com/satori/go.uuid"
)

// Discovery is the interface implementedy by all discovery functions
type Discovery interface {
	StartDiscovery(UUIDpath string, platform string, executionType string) error
	StopDiscovery()
	GetDeviceList() (ExportDeviceMap, error)
	GetDeviceIPListWithService(targetService string) ([]string, error)
	// GetDeviceListWithService(targetService string) (ExportDeviceMap, error)
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


var (
	discoveryIns discoveryImpl
	networkIns   networkhelper.Network

	sysQuery     systemdb.Query
	confQuery    configurationdb.Query
	netQuery     networkdb.Query
	serviceQuery servicedb.Query
)

func init() {
	discoverymgrInfo.wrapperIns = wrapper.GetZeroconfImpl()
	discoverymgrInfo.orchestrationMap = make(OrchestrationMap)
	discoverymgrInfo.shutdownChan = make(chan struct{})

	networkIns = networkhelper.GetInstance()

	sysQuery = systemdb.Query{}
	confQuery = configurationdb.Query{}
	netQuery = networkdb.Query{}
	serviceQuery = servicedb.Query{}
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
	items, err := confQuery.GetList()
	if err != nil {
		return nil, err
	}

	ret := make(ExportDeviceMap)
	for _, item := range items {
		netItems, err := netQuery.Get(item.ID)
		if err != nil {
			continue
		}

		serviceItems, err := serviceQuery.Get(item.ID)
		if err != nil {
			continue
		}

		deviceMap := OrchestrationInformation{
			Platform: item.Platform, ExecutionType: item.ExecType,
			IPv4: netItems.IPv4, ServiceList: serviceItems.Services}

		ret[item.ID] = deviceMap

	}

	if len(ret) == 0 {
		err := errors.NotFound{Message: errormsg.ToString(errormsg.ErrorNoDeviceReturn)}
		return nil, err
	}

	return ret, nil
}

// GetDeviceIPListWithService returns orchestration deviceIP list using service application name
func (discoveryImpl) GetDeviceIPListWithService(targetService string) ([]string, error) {
	var ret []string

	serviceItems, err := serviceQuery.GetList()
	if err != nil {
		return nil, err
	}

	for _, item := range serviceItems {
		for _, service := range item.Services {
			if strings.Compare(service, targetService) == 0 {
				netItems, err := netQuery.Get(item.ID)
				if err != nil {
					continue
				}

				ret = append(ret, netItems.IPv4...)
			}
		}
	}

	if len(ret) == 0 {
		err = errors.NotFound{Message: errormsg.ToString(errormsg.ErrorNoDeviceReturn)}
		return nil, err
	}

	return ret, nil
}

// GetDeviceListWithService returns device info list using service application name
// func (discoveryImpl) GetDeviceListWithService(targetService string) (ExportDeviceMap, error) {
// 	discoverymgrInfo.mapMTX.Lock()
// 	defer discoverymgrInfo.mapMTX.Unlock()

// 	ret := make(ExportDeviceMap)
// 	var err error
// 	for key, value := range discoverymgrInfo.orchestrationMap {
// 		for _, val := range value.ServiceList {
// 			if val == targetService {
// 				ret[key] = *value
// 			}
// 		}
// 	}

// 	if len(ret) == 0 {
// 		err = errormsg.ToError(errormsg.ErrorNoDeviceReturn)
// 		return nil, err
// 	}

// 	return ret, nil
// }

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
	err = errors.NotFound{Message: errormsg.ToString(errormsg.ErrorNoDeviceReturn)}
	return nil, err
}

// DeleteDevice deletes device info using deviceIP
func (discoveryImpl) DeleteDeviceWithIP(targetIP string) {
	// @TODO Delete device with ip in DB
}

// DeleteDevice delete device using deviceID
func (discoveryImpl) DeleteDeviceWithID(ID string) {
	// @Note Delete device with id in DB
	err := confQuery.Delete(ID)
	if err != nil {
		log.Println(err.Error())
	}

	err = netQuery.Delete(ID)
	if err != nil {
		log.Println(err.Error())
	}

	err = serviceQuery.Delete(ID)
	if err != nil {
		log.Println(err.Error())
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

	deviceID, err := getDeviceID()
	if err != nil {
		return
	}

	serviceInfo := servicedb.ServiceInfo{ID: deviceID, Services: nil}
	updateServiceDB(serviceInfo)

	confItem, err := confQuery.Get(deviceID)
	if err != nil {
		log.Println(err.Error())
		return
	}

	var serverTXT []string
	serverTXT = append(serverTXT, confItem.ExecType)
	serverTXT = append(serverTXT, confItem.Platform)

	setNewServiceList(serverTXT)
}

func detectNetworkChgRoutine() {
	ips := networkIns.AppendSubscriber()

	for {
		select {
		case <-discoverymgrInfo.shutdownChan:
			return
		// @TODO set network db will be implemented in next commits,
		// because change of networkhelper is not applied yet.
		case latestIPs := <-ips:

			err := serverPresenceChecker()
			if err != nil {
				continue
			}
			discoverymgrInfo.wrapperIns.ResetServer(latestIPs)
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

func getDeviceID() (string, error) {
	sysInfo, err := getSystemDB()
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	return sysInfo.ID, nil
}

func getPlatform() (string, error) {
	sysInfo, err := getSystemDB()
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	return sysInfo.Platform, nil
}

func getExecType() (string, error) {
	sysInfo, err := getSystemDB()
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	return sysInfo.ExecType, nil
}

func setDeviceInfo(platform string, executionType string) {
	log.Println(logPrefix, "Platform::", platform, " OnboardType::", executionType)

	discoverymgrInfo.platform = platform
	discoverymgrInfo.executionType = executionType
}

func startServer(deviceUUID string, platform string, executionType string) {
	deviceDetectionRoutine()

	deviceID, hostName, Text := setDeviceArgument(deviceUUID, platform, executionType)

	// @Note store system information(id, platform and execution type) to system db
	sysInfo := systemdb.SystemInfo{ID: deviceID, Platform: platform, ExecType: executionType}
	setSystemDB(sysInfo)

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

	// Set Configuration Information to configuration DB
	_, confInfo, netInfo, serviceInfo := convertToDBInfo(myDeviceEntity)

	setConfigurationDB(confInfo)
	setNetworkDB(netInfo)
	setServiceDB(serviceInfo)

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
	var ip []string
	var err error
	// TODO : change to channel
	for {
		ip, err = networkIns.GetIPs()
		if len(ip) != 0 {
			break
		}
		log.Println(logPrefix, errormsg.ToString(err))
		time.Sleep(1 * time.Second)
	}
	log.Println(logPrefix, ip)

	hostIPAddr = ip

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

				_, confInfo, netInfo, serviceInfo := convertToDBInfo(*data)

				// @Note Is it need to call Update API?
				setConfigurationDB(confInfo)
				setNetworkDB(netInfo)
				setServiceDB(serviceInfo)
			}
		}
	}()
}

func serverPresenceChecker() error {

	list, _ := confQuery.GetList()
	if len(list) == 0 {
		return errors.SystemError{Message: "no server initiated yet"}
	}
  
	return nil
}

func shutdownDiscoverymgr() {
	// discoverymgrInfo.deviceID = ""
	// discoverymgrInfo.orchestrationMap = make(OrchestrationMap)
	close(discoverymgrInfo.shutdownChan)
}

func serviceNameChecker(serviceName string) error {
	if serviceName == "" {
		return errors.InvalidParam{Message: "no argument"}
	}
	if serviceName == discoverymgrInfo.platform || serviceName == discoverymgrInfo.executionType {
		return errors.InvalidParam{Message: "cannot change fixed field"}
	}
	return nil
}

func appendServiceToTXT(serviceName string) ([]string, error) {
	serverTXT := discoverymgrInfo.wrapperIns.GetText()
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
	if len(serverTXT) > 2 {
		newServiceList := serverTXT[2:]

		deviceID, err := getDeviceID()
		if err != nil {
			return
		}

		serviceInfo := servicedb.ServiceInfo{ID: deviceID, Services: newServiceList}

		updateServiceDB(serviceInfo)
	}

	discoverymgrInfo.wrapperIns.SetText(serverTXT)
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

func convertToDBInfo(entity wrapper.Entity) (string, configurationdb.Configuration, networkdb.NetworkInfo, servicedb.ServiceInfo) {
	data := entity.OrchestrationInfo

	confInfo := configurationdb.Configuration{}
	netInfo := networkdb.NetworkInfo{}
	serviceInfo := servicedb.ServiceInfo{}

	confInfo.ID = entity.DeviceID
	confInfo.ExecType = data.ExecutionType
	confInfo.Platform = data.Platform

	netInfo.ID = entity.DeviceID
	netInfo.IPv4 = data.IPv4

	serviceInfo.ID = entity.DeviceID
	serviceInfo.Services = data.ServiceList

	return entity.DeviceID, confInfo, netInfo, serviceInfo
}

func setSystemDB(sysInfo systemdb.SystemInfo) {
	err := sysQuery.Set(sysInfo)
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

func setServiceDB(serviceInfo servicedb.ServiceInfo) {
	err := serviceQuery.Set(serviceInfo)
	if err != nil {
		log.Println(logPrefix, err.Error())
	}
}

func getSystemDB() (sysInfo systemdb.SystemInfo, err error) {
	sysInfos, err := sysQuery.GetList()
	if err != nil {
		log.Println(logPrefix, err.Error())
	}

	if len(sysInfos) != 0 {
		sysInfo.ID = sysInfos[0].ID
		sysInfo.Platform = sysInfos[0].Platform
		sysInfo.ExecType = sysInfos[0].ExecType
	}

	return
}

func updateServiceDB(serviceInfo servicedb.ServiceInfo) {
	err := serviceQuery.Update(serviceInfo)
	if err != nil {
		log.Println(logPrefix, err.Error())
	}
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

	err = serviceQuery.Delete(deviceID)
	if err != nil {
		log.Println(err.Error())
	}
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
