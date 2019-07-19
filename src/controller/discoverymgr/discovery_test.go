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
	systemdb "db/bolt/system"
	"log"
	"net"
	"reflect"
	"testing"
	"time"

	errormsg "common/errormsg"
	errors "common/errors"
	networkmocks "common/networkhelper/mocks"
	wrapper "controller/discoverymgr/wrapper"
	wrappermocks "controller/discoverymgr/wrapper/mocks"

	"github.com/golang/mock/gomock"
)

var (
	mockWrapper          *wrappermocks.MockZeroconfInterface
	mockNetwork          *networkmocks.MockNetwork
	defaultUUIDPath      = "/etc/orchestration_deviceID.txt"
	defaultPlatform      = "LINUX"
	defaultExecutionType = "Executable"
	defaultService       = "ls"
	defaultIPv4          = "1.1.1.1"
	defaultMyDeviceID    = "edge-orchestration-test-device-id1"
	defaultIPv4List      = []string{defaultIPv4}
	defaultServiceList   = []string{defaultService}

	anotherService     = "docker"
	anotherIPv4        = "2.2.2.2"
	anotherDeviceID    = "edge-orchestration-test-device-id2"
	anotherIPv4List    = []string{anotherIPv4}
	anotherServiceList = []string{anotherService}

	defaultMyDeviceEntity = wrapper.Entity{
		DeviceID: defaultMyDeviceID,
		TTL:      120,
		OrchestrationInfo: wrapper.OrchestrationInformation{
			IPv4:          defaultIPv4List,
			Platform:      defaultPlatform,
			ExecutionType: defaultExecutionType,
			ServiceList:   defaultServiceList,
		},
	}
	anotherEntity = wrapper.Entity{
		DeviceID: anotherDeviceID,
		TTL:      120,
		OrchestrationInfo: wrapper.OrchestrationInformation{
			IPv4:          anotherIPv4List,
			Platform:      defaultPlatform,
			ExecutionType: defaultExecutionType,
			ServiceList:   anotherServiceList,
		},
	}

	noDeviceReturnErr = errors.NotFound{Message: errormsg.ToString(errormsg.ErrorNoDeviceReturn)}
)

func createMockIns(ctrl *gomock.Controller) {
	mockWrapper = wrappermocks.NewMockZeroconfInterface(ctrl)
	mockNetwork = networkmocks.NewMockNetwork(ctrl)

	wrapperIns = mockWrapper
	networkIns = mockNetwork
}

func addDevice(Another bool) {
	deviceID, confInfo, netInfo, serviceInfo := convertToDBInfo(defaultMyDeviceEntity)
	sysInfo := systemdb.SystemInfo{
		ID: deviceID, Platform: defaultPlatform, ExecType: defaultExecutionType}

	log.Println(logPrefix, "[addDevice]", deviceID)
	setSystemDB(sysInfo)
	setConfigurationDB(confInfo)
	setNetworkDB(netInfo)
	setServiceDB(serviceInfo)

	if !Another {
		return
	}

	deviceID, confInfo, netInfo, serviceInfo = convertToDBInfo(anotherEntity)

	setConfigurationDB(confInfo)
	setNetworkDB(netInfo)
	setServiceDB(serviceInfo)
}

func checkPresence(t *testing.T, deviceID string) {
	t.Helper()

	_, err := confQuery.Get(deviceID)
	if err != nil {
		t.Error(err.Error())
	}

	_, err = netQuery.Get(deviceID)
	if err != nil {
		t.Error(err.Error())
	}

	_, err = serviceQuery.Get(deviceID)
	if err != nil {
		t.Error(err.Error())
	}

	return
}

func checkNotPresence(t *testing.T, deviceID string) {
	t.Helper()

	_, err := confQuery.Get(deviceID)
	if err == nil {
		t.Error()
	}

	_, err = netQuery.Get(deviceID)
	if err == nil {
		t.Error()
	}

	_, err = serviceQuery.Get(deviceID)
	if err == nil {
		t.Error()
	}

	return
}

func checkClearMap(t *testing.T) {
	t.Helper()
	sysInfos, err := sysQuery.GetList()
	if len(sysInfos) != 1 || err != nil {
		t.Error("Delete MySelf")
	}

	confInfos, err := confQuery.GetList()
	if len(confInfos) != 1 || err != nil {
		t.Error("Not Cleared : Configuration Map")
	}

	netInfos, err := netQuery.GetList()
	if len(netInfos) != 1 || err != nil {
		t.Error("Not Cleared : Network Map")
	}

	serviceInfos, err := serviceQuery.GetList()
	if len(serviceInfos) != 1 || err != nil {
		t.Error("Not Cleared : Service Map")
	}
}

func closeTest() {
	sysInfos, err := sysQuery.GetList()
	if err != nil {
		return
	}

	for _, info := range sysInfos {
		err = sysQuery.Delete(info.ID)
		if err != nil {
			log.Println(logPrefix, err.Error())
		}
	}

	confItems, err := confQuery.GetList()
	if err != nil {
		log.Println(logPrefix, err.Error())
		return
	}

	for _, confItem := range confItems {
		id := confItem.ID
		deleteDevice(id)
	}
}

func TestStartDiscovery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	//success case
	t.Run("Success", func(t *testing.T) {
		//declare mock argument
		devicesubchan := make(chan *wrapper.Entity, 1)
		ipsub := make(chan []net.IP, 1)

		mockNetwork.EXPECT().AppendSubscriber().Return(ipsub)
		mockNetwork.EXPECT().StartNetwork().Return()
		mockNetwork.EXPECT().GetIPs().Return(defaultIPv4List, nil)
		mockNetwork.EXPECT().GetNetInterface().Return(nil, nil)
		mockWrapper.EXPECT().RegisterProxy(gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return(defaultMyDeviceEntity, nil)
		mockWrapper.EXPECT().GetSubscriberChan().Return(devicesubchan, nil)
		mockWrapper.EXPECT().Shutdown().Return()
		//let the test start
		discoveryInstance := GetInstance()
		discoveryInstance.StartDiscovery(defaultUUIDPath,
			defaultPlatform, defaultExecutionType)

		err := serverPresenceChecker()
		if err != nil {
			t.Error("Server is not registered")
		}
		time.Sleep(2 * time.Second)
		discoveryInstance.StopDiscovery()
	})

	closeTest()
}

func TestDeviceDetectionRoutine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("Success", func(t *testing.T) {

		//declare mock argument
		devicesubchan := make(chan *wrapper.Entity, 20)

		mockWrapper.EXPECT().GetSubscriberChan().Return(devicesubchan, nil)

		//set my device info
		addDevice(false)
		shutdownChan = make(chan struct{})
		defer close(shutdownChan)
		//let the test start
		go deviceDetectionRoutine()
		time.Sleep(1 * time.Second)

		t.Run("SuccessClearMap", func(t *testing.T) {
			_, confInfo, netInfo, serviceInfo := convertToDBInfo(anotherEntity)

			setConfigurationDB(confInfo)
			setNetworkDB(netInfo)
			setServiceDB(serviceInfo)

			devicesubchan <- nil
			time.Sleep(1 * time.Second)

			checkNotPresence(t, anotherDeviceID)
			checkPresence(t, defaultMyDeviceID)
		})
		t.Run("SuccessDeleteDevice", func(t *testing.T) {
			_, confInfo, netInfo, serviceInfo := convertToDBInfo(anotherEntity)

			setConfigurationDB(confInfo)
			setNetworkDB(netInfo)
			setServiceDB(serviceInfo)

			tmpEntity := anotherEntity
			tmpEntity.TTL = 0
			devicesubchan <- &tmpEntity
			time.Sleep(1 * time.Second)

			checkNotPresence(t, anotherDeviceID)
		})
		t.Run("SuccessUpdateInfo", func(t *testing.T) {
			tmpEntity := defaultMyDeviceEntity
			tmpEntity.OrchestrationInfo.IPv4 = anotherIPv4List
			devicesubchan <- &tmpEntity
			time.Sleep(1 * time.Second)

			presence := false

			deviceID, err := getDeviceID()
			log.Println(deviceID)
			if err != nil {
				t.Fatal(err.Error())
			}

			netInfo, err := netQuery.Get(deviceID)
			if err != nil {
				t.Fatal(err.Error())
			}

			for _, ip := range netInfo.IPv4 {
				if ip == anotherIPv4 {
					presence = true
				}
			}
			if presence == false {
				t.Error("Info Not Updated ::", netInfo.IPv4)
			}
		})
		t.Run("SuccessNewDevice", func(t *testing.T) {
			devicesubchan <- &anotherEntity
			time.Sleep(1 * time.Second)

			checkPresence(t, anotherDeviceID)
		})
	})

	closeTest()
}

func TestGetDeviceList(t *testing.T) {

	discoveryInstance := GetInstance()

	addDevice(true)

	t.Run("Success", func(t *testing.T) {
		t.Run("GetDeviceList", func(t *testing.T) {
			deviceMap, _ := discoveryInstance.GetDeviceList()
			if len(deviceMap) != 2 {
				t.Error("More or less Device Registered : ", len(deviceMap))
			}
		})
	})
	closeTest()
}

func TestGetDeviceIPListWithService(t *testing.T) {

	discoveryInstance := GetInstance()

	addDevice(true)

	t.Run("Success", func(t *testing.T) {
		t.Run("GetDeviceIPListWithService", func(t *testing.T) {
			deviceStringList, _ := discoveryInstance.GetDeviceIPListWithService(defaultService)
			if deviceStringList[0] != defaultIPv4 {
				t.Error("Error return []string W/ Service : ", deviceStringList)
			}
		})
	})
	t.Run("Fail", func(t *testing.T) {
		failService := "NoServiceIsThisName"
		t.Run("GetDeviceIPListWithService", func(t *testing.T) {
			_, err := discoveryInstance.GetDeviceIPListWithService(failService)
			log.Println(err)
			if err != noDeviceReturnErr {
				t.Error("Error is not generated : ", err)
			}
		})
	})
	closeTest()
}

// func TestGetDeviceListWithService(t *testing.T) {

// 	discoveryInstance := GetInstance()

// 	addDevice(true)

// 	t.Run("Success", func(t *testing.T) {
// 		t.Run("GetDeviceListWithService", func(t *testing.T) {
// 			deviceMap, _ := discoveryInstance.GetDeviceListWithService(defaultService)
// 			presence := false
// 			for k, v := range deviceMap {
// 				if k == defaultMyDeviceID {
// 					if v.IPv4[0] == defaultIPv4 {
// 						presence = true
// 					}
// 				}
// 			}
// 			if presence == false {
// 				t.Error("Cannot Find my device w/ service")
// 			}
// 		})
// 	})
// 	t.Run("Fail", func(t *testing.T) {
// 		failService := "NoServiceIsThisName"
// 		t.Run("GetDeviceListWithService", func(t *testing.T) {
// 			_, err := discoveryInstance.GetDeviceListWithService(failService)
// 			if err != noDeviceReturnErr {
// 				t.Error("Error is not generated : ", err)
// 			}
// 		})
// 	})
// 	closeTest()
// }
func TestGetDeviceWithID(t *testing.T) {

	discoveryInstance := GetInstance()

	addDevice(true)

	t.Run("Success", func(t *testing.T) {
		t.Run("GetDeviceWithID", func(t *testing.T) {
			deviceMap, _ := discoveryInstance.GetDeviceWithID(defaultMyDeviceID)
			presence := false
			for k, v := range deviceMap {
				if k == defaultMyDeviceID {
					if v.IPv4[0] == defaultIPv4 {
						presence = true
					}
				}
			}
			if presence == false {
				t.Error("Cannot Find my device w/ service")
			}
		})
	})
	t.Run("Fail", func(t *testing.T) {
		failID := "NoDeviceIsThisID"
		t.Run("GetDeviceWithID", func(t *testing.T) {
			_, err := discoveryInstance.GetDeviceWithID(failID)
			if err != noDeviceReturnErr {
				t.Error("Error is not generated : ", err)
			}
		})
	})
	closeTest()
}

func TestDeleteDeviceWithID(t *testing.T) {

	discoveryInstance := GetInstance()

	addDevice(true)

	t.Run("Success", func(t *testing.T) {
		t.Run("DeleteDeviceWithID", func(t *testing.T) {
			discoveryInstance.DeleteDeviceWithID(anotherDeviceID)
			checkNotPresence(t, anotherDeviceID)
		})
		addDevice(true)
	})

	t.Run("Fail", func(t *testing.T) {
		t.Run("DeleteDeviceWithID", func(t *testing.T) {
			discoveryInstance.DeleteDeviceWithID(defaultMyDeviceID)
			checkPresence(t, defaultMyDeviceID)
		})
	})
	closeTest()
}
func TestAddNewServiceName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	discoveryInstance := GetInstance()

	addDevice(false)

	serverTXT := []string{defaultPlatform, defaultExecutionType, defaultService}

	t.Run("Success", func(t *testing.T) {
		newServiceName := "NewService"

		t.Run("AddNewServiceName", func(t *testing.T) {
			mockWrapper.EXPECT().GetText().Return(serverTXT)
			mockWrapper.EXPECT().SetText(gomock.Any()).Return()

			err := discoveryInstance.AddNewServiceName(newServiceName)
			if err != nil {
				t.Error("add new service error : ", err)
			}
			presence := false

			deviceID, err := getDeviceID()
			if err != nil {
				t.Fatal(err.Error())
			}

			serviceInfo, err := serviceQuery.Get(deviceID)
			if err != nil {
				t.Fatal(err.Error())
			}

			for _, service := range serviceInfo.Services {
				if service == newServiceName {
					presence = true
				}
			}

			if presence == false {
				t.Error("add new service fail without error")
			}
		})
	})
	closeTest()
}
func TestRemoveServiceName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	discoveryInstance := GetInstance()

	addDevice(false)

	serverTXT := []string{defaultPlatform, defaultExecutionType, defaultService}

	t.Run("Success", func(t *testing.T) {
		t.Run("RemoveServiceName", func(t *testing.T) {
			mockWrapper.EXPECT().GetText().Return(serverTXT)
			mockWrapper.EXPECT().SetText(gomock.Any()).Return()
			err := discoveryInstance.RemoveServiceName(defaultService)
			if err != nil {
				t.Error("delete service error : ", err)
			}

			isPresence := false

			deviceID, err := getDeviceID()
			if err != nil {
				t.Fatal(err.Error())
			}

			serviceInfo, err := serviceQuery.Get(deviceID)
			if err != nil {
				t.Fatal(err.Error())
			}

			for _, service := range serviceInfo.Services {
				if service == defaultService {
					isPresence = true
				}
			}

			if isPresence == true {
				t.Error("remove service fail")
			}
		})
	})

	closeTest()
}
func TestResetServiceName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	discoveryInstance := GetInstance()

	addDevice(false)

	t.Run("Success", func(t *testing.T) {
		t.Run("ResetServiceName", func(t *testing.T) {
			mockWrapper.EXPECT().SetText(gomock.Any()).Return()
			discoveryInstance.ResetServiceName()

			deviceID, err := getDeviceID()
			if err != nil {
				t.Fatal(err.Error())
			}

			serviceInfo, err := serviceQuery.Get(deviceID)
			if err != nil {
				t.Fatal(err.Error())
			}

			if len(serviceInfo.Services) != 0 {
				t.Error("Reset failed")
			}
		})
	})

	closeTest()
}
func TestServiceNameChecker(t *testing.T) {
	addDevice(false)

	t.Run("Fail", func(t *testing.T) {
		t.Run("serviceNameChecker", func(t *testing.T) {
			err := serviceNameChecker("")
			if err == nil {
				t.Error()
			}
			err = serviceNameChecker(defaultPlatform)
			if err == nil {
				t.Error()
			}
		})
	})
	closeTest()
}
func TestAppendServiceToTXT(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	serverTXT := []string{defaultPlatform, defaultExecutionType, defaultService}

	t.Run("Fail", func(t *testing.T) {
		t.Run("appendServiceToTXT", func(t *testing.T) {
			mockWrapper.EXPECT().GetText().Return(serverTXT)
			_, err := appendServiceToTXT(defaultService)
			if err == nil {
				t.Error()
			}
		})
	})
	closeTest()
}
func TestMdnsTXTSizeChecker(t *testing.T) {
	t.Run("Fail", func(t *testing.T) {
		t.Run("mdnsTXTSizeChecker", func(t *testing.T) {
			ServiceListMoreThan400B := []string{
				"TXT Size is Too much for mDNS TXT - 400B0",
				"TXT Size is Too much for mDNS TXT - 400B1",
				"TXT Size is Too much for mDNS TXT - 400B2",
				"TXT Size is Too much for mDNS TXT - 400B3",
				"TXT Size is Too much for mDNS TXT - 400B4",
				"TXT Size is Too much for mDNS TXT - 400B5",
				"TXT Size is Too much for mDNS TXT - 400B6",
				"TXT Size is Too much for mDNS TXT - 400B7",
				"TXT Size is Too much for mDNS TXT - 400B8",
				"TXT Size is Too much for mDNS TXT - 400B9",
			}
			err := mdnsTXTSizeChecker(ServiceListMoreThan400B)
			if err == nil {
				t.Error()
			}
		})
	})
	closeTest()
}
func TestServiceName(t *testing.T) {
	serverTXT := []string{defaultPlatform, defaultExecutionType, defaultService}
	t.Run("Fail", func(t *testing.T) {
		t.Run("getIndexToDelete", func(t *testing.T) {
			failService := "NoServiceIsThisName"
			_, err := getIndexToDelete(serverTXT, failService)
			if err == nil {
				t.Error()
			}
		})
	})
	closeTest()
}
func TestDetectNetworkChgRoutine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)
	addDevice(false)

	shutdownChan = make(chan struct{})
	defer close(shutdownChan)
	//set mock in order
	ipsub := make(chan []net.IP, 2)
	mockNetwork.EXPECT().AppendSubscriber().Return(ipsub)
	mockWrapper.EXPECT().ResetServer(gomock.Any()).Return()
	go detectNetworkChgRoutine()

	t.Run("Success", func(t *testing.T) {
		expectedIP := []string{"192.0.2.1"}
		ipsub <- []net.IP{net.ParseIP("192.0.2.1")}

		time.Sleep(time.Millisecond * time.Duration(10))
		deviceID, err := getDeviceID()
		if err != nil {
			t.Error(err.Error())
		}

		netInfo, err := netQuery.Get(deviceID)
		if err != nil {
			t.Error(err.Error())
		}

		log.Println(netInfo.IPv4, expectedIP)
		if reflect.DeepEqual(netInfo.IPv4, expectedIP) != true {
			t.Error()
		}
	})
	// time.Sleep(1 * time.Second)
	// t.Run("Fail", func(t *testing.T) {
	// 	// discoverymgrInfo.deviceID = ""
	// 	ipsub <- []net.IP{net.ParseIP("192.0.2.1")}

	// })

	shutdownChan <- struct{}{}
}

func TestGetDeviceID(t *testing.T) {
	t.Run("SuccessNewV4", func(t *testing.T) {
		uuid, _ := setDeviceID("/x/y/z/NoFileIsThisName")
		if len(uuid) == 0 {
			t.Error()
		}
	})
}
