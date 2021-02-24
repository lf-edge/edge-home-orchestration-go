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

package discoverymgr

import (
	"bufio"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"sync"
	"time"

	errors "github.com/lf-edge/edge-home-orchestration-go/internal/common/errors"
	networkhelper "github.com/lf-edge/edge-home-orchestration-go/internal/common/networkhelper"
	mnedc "github.com/lf-edge/edge-home-orchestration-go/internal/controller/discoverymgr/mnedc"
	wrapper "github.com/lf-edge/edge-home-orchestration-go/internal/controller/discoverymgr/wrapper"

	configurationdb "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/configuration"
	networkdb "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/network"
	servicedb "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/service"
	systemdb "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/system"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/cipher"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/client"

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
	AddDeviceInfo(deviceID string, virtualAddr string, privateAddr string)
	GetOrchestrationInfo() (platfrom string, executionType string, serviceList []string, err error)
	SetRestResource()
	MNEDCClosedCallback()
	NotifyMNEDCBroadcastServer() error
	MNEDCReconciledCallback()
	GetDeviceID() (id string, err error)
	StartMNEDCClient(string, string)
	StartMNEDCServer(string)
	client.Setter
	cipher.Setter
}

//DiscoveryImpl Structure
type DiscoveryImpl struct {
	client.HasClient
	cipher.HasCipher
}

func (d *DiscoveryImpl) GetDeviceID() (id string, err error) {
	id, err = getSystemDB(systemdb.ID)
	if err != nil {
		log.Println(err.Error())
	}

	return
}

var (
	discoveryIns *DiscoveryImpl
	networkIns   networkhelper.Network
	mutexLock    sync.Mutex
	log          = logmgr.GetInstance()
)

func init() {
	discoveryIns = &DiscoveryImpl{}
	wrapperIns = wrapper.GetZeroconfImpl()
	shutdownChan = make(chan struct{})

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

// SetRestResource sets clienter
func (d *DiscoveryImpl) SetRestResource() {
	d.SetClient(d.Clienter)
}

// StartDiscovery starts server for network registration and do orchestration discovery activity
func (DiscoveryImpl) StartDiscovery(UUIDpath string, platform string, executionType string) (err error) {
	mutexLock.Lock()
	clearAllDeviceInfo()
	mutexLock.Unlock()
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
func (DiscoveryImpl) StopDiscovery() {
	err := serverPresenceChecker()
	if err != nil {
		log.Println(logPrefix, "[StopDiscovery]", err)
		return
	}
	shutdownDiscoverymgr()
	wrapperIns.Shutdown()
}

// DeleteDeviceWithIP deletes device info using deviceIP
func (DiscoveryImpl) DeleteDeviceWithIP(targetIP string) {
	// @TODO Delete device with ip in DB
}

// DeleteDeviceWithID delete device using deviceID
func (d DiscoveryImpl) DeleteDeviceWithID(ID string) {
	// @Note Delete device with id in DB
	deviceID, err := d.GetDeviceID()
	if err != nil {
		log.Println(err.Error())
		return
	}

	if deviceID != ID {
		deleteDevice(ID)
	}
}

// AddNewServiceName sets text field of mdns message with service application name
func (d *DiscoveryImpl) AddNewServiceName(serviceName string) error {
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

	d.setNewServiceList(serverTXT)

	return nil
}

// RemoveServiceName removes text field of mdns message with service application name
func (d *DiscoveryImpl) RemoveServiceName(serviceName string) error {
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
	d.setNewServiceList(serverTXT)

	return nil
}

// ResetServiceName resets text field of mdns message
func (d *DiscoveryImpl) ResetServiceName() {
	err := serverPresenceChecker()
	if err != nil {
		log.Println(logPrefix, "[ResetServiceName]", err)
		return
	}

	deviceID, err := d.GetDeviceID()
	if err != nil {
		return
	}

	serviceInfo := servicedb.ServiceInfo{ID: deviceID, Services: nil}
	setServiceDB(serviceInfo)

	confItem, err := confQuery.Get(deviceID)
	if err != nil {
		log.Println(err.Error())
		return
	}

	var serverTXT []string
	serverTXT = append(serverTXT, confItem.ExecType)
	serverTXT = append(serverTXT, confItem.Platform)

	d.setNewServiceList(serverTXT)
}

// AddDeviceInfo takes public and private IP from relay and requests for orchestration info from this device
func (d *DiscoveryImpl) AddDeviceInfo(deviceID string, virtualAddr string, privateAddr string) {

	log.Println(logPrefix, "[AddDeviceInfo]", "private Addr", privateAddr)
	log.Println(logPrefix, "[AddDeviceInfo]", "Virtual Addr", virtualAddr)

	//Check if the private addr already exists in OrchestrationMap. If exists, dont call requestDeviceInfo()
	isPresent, err := isIPPresent(deviceID, privateAddr)
	if err != nil || !isPresent {
		go d.requestDeviceInfo(deviceID, virtualAddr)
	} else {
		log.Println(logPrefix, "[Add New Device]", "New device Info already present")
	}
}

//GetOrchestrationInfo returns the orchestration info of the device
func (DiscoveryImpl) GetOrchestrationInfo() (platfrom string, executionType string, serviceList []string, err error) {

	log.Println(logPrefix, "Orch info requested")
	serviceList, err = getServiceList()
	if err != nil {
		return
	}
	platfrom, err = getPlatform()
	if err != nil {
		return
	}
	executionType, err = getExecType()
	return
}

func isIPPresent(deviceID string, privateIP string) (isPresent bool, err error) {
	networkInfo, err := netQuery.Get(deviceID)
	if err != nil {
		log.Println(logPrefix, "Error in getting network info of", deviceID)
		return
	}
	ipList := networkInfo.IPv4
	for _, ip := range ipList {
		if ip == privateIP {
			return true, nil
		}
	}
	return false, nil
}

func (d *DiscoveryImpl) requestDeviceInfo(deviceID string, address string) {
	if d.Clienter == nil {
		log.Println(logPrefix, "Client is nil, returning")
		return
	}
	limit := 1
	for {
		platform, executionType, serviceList, err := d.Clienter.DoGetOrchestrationInfo(address)
		if err != nil {
			if limit == 5 {
				log.Println(logPrefix, "Limit reached", "error getting device info", err.Error())
				break
			}
			log.Println(logPrefix, "error getting device info", err.Error())
			limit = limit + 1
			time.Sleep(1 * time.Second)
			continue
		}
		//save the info in db
		log.Println(logPrefix, "Got The Info")
		log.Println(logPrefix, deviceID, platform, executionType, serviceList)
		var ip []string
		ip = append(ip, address)
		data := wrapper.Entity{
			DeviceID: deviceID,
			TTL:      1,
			OrchestrationInfo: wrapper.OrchestrationInformation{
				IPv4:          ip,
				Platform:      platform,
				ExecutionType: executionType,
				ServiceList:   serviceList,
			},
		}
		_, confInfo, netInfo, serviceInfo := convertToDBInfo(data)

		log.Println(logPrefix, "netInfoIP:", netInfo.IPv4)
		log.Println(logPrefix, "netInfoID:", netInfo.ID)
		log.Println(logPrefix, "confInfoID:", confInfo.ID)
		log.Println(logPrefix, "confInfoExec:", confInfo.ExecType)
		log.Println(logPrefix, "confInfoPlatf:", confInfo.Platform)
		log.Println(logPrefix, "serviceInfoID:", serviceInfo.ID)
		log.Println(logPrefix, "serviceInfoServices:", serviceInfo.Services)

		if len(netInfo.IPv4) != 0 {
			setNetworkDB(netInfo)
		}
		// @Note Is it need to call Update API?
		setConfigurationDB(confInfo)
		setServiceDB(serviceInfo)
		break
	}
}

func detectNetworkChgRoutine() {
	ips := networkIns.AppendSubscriber()

	for {
		select {
		case <-shutdownChan:
			return
		case latestIPs := <-ips:
			id, err := discoveryIns.GetDeviceID()
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

func getServiceList() (serviceList []string, err error) {
	id, err := discoveryIns.GetDeviceID()
	if err != nil {
		return
	}
	serviceInfo, err := serviceQuery.Get(id)
	if err != nil {
		return
	}
	serviceList = serviceInfo.Services
	return
}

func startServer(deviceUUID string, platform string, executionType string) {
	deviceDetectionRoutine()

	deviceID, hostName, Text := setDeviceArgument(deviceUUID, platform, executionType)

	// @Note store system information(id, platform and execution type) to system db
	setSystemDB(deviceID, platform, executionType)

	hostIPAddr, netIface := SetNetwotkArgument()
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
	_, confInfo, netInfo, serviceInfo := convertToDBInfo(myDeviceEntity)

	setConfigurationDB(confInfo)
	setNetworkDB(netInfo)
	setServiceDB(serviceInfo)

	return
}

//NotifyMNEDCBroadcastServer registers to MNEDC
func (d *DiscoveryImpl) NotifyMNEDCBroadcastServer() error {
	log.Println(logPrefix, "Registering to Broadcast server")
	isMNEDCConnected = true
	virtualIP, err := networkIns.GetVirtualIP()
	if err != nil {
		log.Println(logPrefix, "Cant register to Broadcast server, virtual IP error", err.Error())
		return err
	}

	privateIP, err := networkIns.GetOutboundIP()
	if err != nil {
		log.Println(logPrefix, "Cant register to Broadcast server, outbound IP error", err.Error())
		return err
	}

	deviceID, err := d.GetDeviceID()
	if err != nil {
		log.Println(logPrefix, "Error getting device ID while registering to Broadcast server", err.Error())
		return err
	}

	file, err := os.Open(configPath)

	if err != nil {
		log.Println(logPrefix, "cant read config file from", configPath, err.Error(), "trying config alternate")

		file, err = os.Open(configAlternate)
		if err != nil {
			log.Println(logPrefix, "cant register to server", "failed for config alternate too", err.Error())
			return err
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	scanner.Scan()
	serverIP := scanner.Text()

	go func() {

		if d.Clienter == nil {
			log.Println(logPrefix, "Client is nil, returning")
			err = errors.InvalidParam{Message: "Client Nil"}
		}
		err = d.Clienter.DoNotifyMNEDCBroadcastServer(serverIP, mnedcBroadcastServerPort, deviceID, privateIP, virtualIP)
		if err != nil {
			log.Println(logPrefix, "Cannot register to Broadcast server", err.Error())
		}
	}()

	time.Sleep(5 * time.Second)
	if err != nil {
		return err
	}

	return nil

}

func setDeviceArgument(deviceUUID string, platform string, executionType string) (deviceID string, hostName string, Text []string) {
	deviceID = "edge-orchestration-" + deviceUUID
	hostName = "edge-" + deviceUUID

	Text = append(Text, platform)
	Text = append(Text, executionType)
	return
}

func SetNetwotkArgument() (hostIPAddr []string, netIface []net.Interface) {
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

				_, confInfo, netInfo, serviceInfo := convertToDBInfo(*data)

				log.Printf("[deviceDetectionRoutine] %s", data.DeviceID)
				log.Printf("[deviceDetectionRoutine] confInfo    : ExecType(%s), Platform(%s)", confInfo.ExecType, confInfo.Platform)
				log.Printf("[deviceDetectionRoutine] netInfo     : IPv4(%s), RTT(%v)", netInfo.IPv4, netInfo.RTT)
				log.Printf("[deviceDetectionRoutine] serviceInfo : Services(%v)", serviceInfo.Services)
				log.Printf("")

				var info networkdb.NetworkInfo
				info, err = getNetworkDB(netInfo.ID)

				if err != nil || !reflect.DeepEqual(netInfo.IPv4, info.IPv4) {
					setNetworkDB(netInfo)
				}

				// @Note Is it need to call Update API?
				setConfigurationDB(confInfo)
				setServiceDB(serviceInfo)
			}
		}
	}()
}

func serverPresenceChecker() error {
	_, err := discoveryIns.GetDeviceID()
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

func (d *DiscoveryImpl) setNewServiceList(serverTXT []string) {
	// if len(serverTXT) > 2 {
	newServiceList := serverTXT[2:]

	deviceID, err := discoveryIns.GetDeviceID()
	if err != nil {
		return
	}

	serviceInfo := servicedb.ServiceInfo{ID: deviceID, Services: newServiceList}

	setServiceDB(serviceInfo)

	wrapperIns.SetText(serverTXT)

	//Again Register to Broadcast server to let other devices know the updated service list

	if isMNEDCConnected {
		err = d.NotifyMNEDCBroadcastServer()
		if err != nil {
			log.Println(logPrefix, "Service updation failed through Broadcast server")
		}
	}
}

//MNEDCClosedCallback handles discovery behaviour when MNEDC connection is closed
func (d *DiscoveryImpl) MNEDCClosedCallback() {
	isMNEDCConnected = false
	//delete devices with virtual IPs
}

//MNEDCReconciledCallback handles discovery behaviour when MNEDC connection is closed
func (d *DiscoveryImpl) MNEDCReconciledCallback() {
	isMNEDCConnected = true
	err := d.NotifyMNEDCBroadcastServer()
	if err != nil {
		log.Println(logPrefix, "Could not reconect to Broadcast server")
	}
}

//StartMNEDCClient Starts MNEDC client
func (d *DiscoveryImpl) StartMNEDCClient(deviceIDFilePath, mnedcServerConfig string) {
	mnedc.GetClientInstance().StartMNEDCClient(deviceIDFilePath, mnedcServerConfig)
}

//StartMNEDCServer Starts MNEDC server
func (d *DiscoveryImpl) StartMNEDCServer(deviceIDFilePath string) {
	mnedc.GetServerInstance().StartMNEDCServer(deviceIDFilePath)
}

// ClearMap makes map empty and only leaves my device info
func clearMap() {
	log.Println(logPrefix, "[clearMap]")

	confItems, err := confQuery.GetList()
	if err != nil {
		log.Println(logPrefix, err.Error())
		return
	}

	deviceID, err := discoveryIns.GetDeviceID()
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

// clearAllDeviceInfo deletes all information from DB
func clearAllDeviceInfo() {
	log.Println(logPrefix, "Delete All Device Info")

        confItems, err := confQuery.GetList()
        if err != nil {
                log.Println(logPrefix, err.Error())
                return
        }

        for _, confItem := range confItems {
                deleteDevice(confItem.ID)
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

func setServiceDB(serviceInfo servicedb.ServiceInfo) {
	err := serviceQuery.Set(serviceInfo)
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

	err = serviceQuery.Delete(deviceID)
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
