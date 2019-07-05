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
	"log"
	"net"
	"testing"
	"time"

	errormsg "common/errormsg"
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
)

func createMockIns(ctrl *gomock.Controller) {
	mockWrapper = wrappermocks.NewMockZeroconfInterface(ctrl)
	mockNetwork = networkmocks.NewMockNetwork(ctrl)

	discoverymgrInfo.wrapperIns = mockWrapper
	networkIns = mockNetwork
}

func addDevice(Another bool) {
	deviceID, deviceInfo := convertwrappertoDB(defaultMyDeviceEntity)
	newDeviceHandler(deviceID, deviceInfo)
	discoverymgrInfo.deviceID = deviceID
	if !Another {
		return
	}
	deviceID, deviceInfo = convertwrappertoDB(anotherEntity)
	newDeviceHandler(deviceID, deviceInfo)
}

func closeTest() {
	discoverymgrInfo.deviceID = ""
	discoverymgrInfo.platform = defaultPlatform
	discoverymgrInfo.executionType = defaultExecutionType
	discoverymgrInfo.orchestrationMap = make(OrchestrationMap)
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
		discoverymgrInfo.shutdownChan = make(chan struct{})
		defer close(discoverymgrInfo.shutdownChan)
		//let the test start
		go deviceDetectionRoutine()

		t.Run("SuccessClearMap", func(t *testing.T) {
			deviceID, deviceInfo := convertwrappertoDB(anotherEntity)
			newDeviceHandler(deviceID, deviceInfo)

			devicesubchan <- nil
			time.Sleep(1 * time.Second)
			if len(discoverymgrInfo.orchestrationMap) != 1 {
				t.Error("Map Not Cleared : ", len(discoverymgrInfo.orchestrationMap))
			}
			_, presence := discoverymgrInfo.orchestrationMap[defaultMyDeviceID]
			if !presence {
				t.Error("Delete Myself")
			}
		})
		t.Run("SuccessDeleteDevice", func(t *testing.T) {
			deviceID, deviceInfo := convertwrappertoDB(anotherEntity)
			newDeviceHandler(deviceID, deviceInfo)
			tmpEntity := anotherEntity
			tmpEntity.TTL = 0
			devicesubchan <- &tmpEntity
			time.Sleep(1 * time.Second)
			_, presence := discoverymgrInfo.orchestrationMap[anotherDeviceID]
			if presence {
				t.Error("Device Not Deleted")
			}
		})
		t.Run("SuccessUpdateInfo", func(t *testing.T) {
			tmpEntity := defaultMyDeviceEntity
			tmpEntity.OrchestrationInfo.IPv4 = anotherIPv4List
			devicesubchan <- &tmpEntity
			time.Sleep(1 * time.Second)
			myDevice := discoverymgrInfo.orchestrationMap[defaultMyDeviceID]
			if myDevice.IPv4[0] != anotherIPv4 {
				t.Error("Info Not Updated ::", myDevice.IPv4)
			}
		})
		t.Run("SuccessNewDevice", func(t *testing.T) {
			devicesubchan <- &anotherEntity
			time.Sleep(1 * time.Second)
			_, presence := discoverymgrInfo.orchestrationMap[anotherDeviceID]
			if !presence {
				t.Error("Device Not Generated")
			}
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
			if errormsg.ToInt(err) != errormsg.ErrorNoDeviceReturn {
				t.Error("Error is not generated : ", err)
			}
		})
	})
	closeTest()
}

func TestGetDeviceListWithService(t *testing.T) {

	discoveryInstance := GetInstance()

	addDevice(true)

	t.Run("Success", func(t *testing.T) {
		t.Run("GetDeviceListWithService", func(t *testing.T) {
			deviceMap, _ := discoveryInstance.GetDeviceListWithService(defaultService)
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
		failService := "NoServiceIsThisName"
		t.Run("GetDeviceListWithService", func(t *testing.T) {
			_, err := discoveryInstance.GetDeviceListWithService(failService)
			if errormsg.ToInt(err) != errormsg.ErrorNoDeviceReturn {
				t.Error("Error is not generated : ", err)
			}
		})
	})
	closeTest()
}
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
			if errormsg.ToInt(err) != errormsg.ErrorNoDeviceReturn {
				t.Error("Error is not generated : ", err)
			}
		})
	})
	closeTest()
}

func TestDeleteDeviceWithIP(t *testing.T) {

	discoveryInstance := GetInstance()

	addDevice(true)

	t.Run("Success", func(t *testing.T) {
		t.Run("DeleteDeviceWithIP", func(t *testing.T) {
			discoveryInstance.DeleteDeviceWithIP(anotherIPv4)
			if _, presence := discoverymgrInfo.orchestrationMap[anotherDeviceID]; presence {
				t.Error("Delete Do Not Work")
			}
		})
		addDevice(true)
	})

	t.Run("Fail", func(t *testing.T) {
		t.Run("DeleteDeviceWithIP", func(t *testing.T) {
			discoveryInstance.DeleteDeviceWithIP(defaultIPv4)
			if _, presence := discoverymgrInfo.orchestrationMap[defaultMyDeviceID]; !presence {
				t.Error("Delete My Device")
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
			if _, presence := discoverymgrInfo.orchestrationMap[anotherDeviceID]; presence {
				t.Error("Delete Do Not Work")
			}
		})
		addDevice(true)
	})

	t.Run("Fail", func(t *testing.T) {
		t.Run("DeleteDeviceWithID", func(t *testing.T) {
			discoveryInstance.DeleteDeviceWithID(defaultMyDeviceID)
			if _, presence := discoverymgrInfo.orchestrationMap[defaultMyDeviceID]; !presence {
				t.Error("Delete My Device")
			}
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
			log.Println(discoverymgrInfo.orchestrationMap[discoverymgrInfo.deviceID].ServiceList)
			for _, val := range discoverymgrInfo.orchestrationMap[discoverymgrInfo.deviceID].ServiceList {
				if val == newServiceName {
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
			presence := true
			for _, val := range discoverymgrInfo.orchestrationMap[discoverymgrInfo.deviceID].ServiceList {
				if val == defaultService {
					presence = false
				}
			}
			if presence == true {
				t.Error("add new service fail without error")
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
			if len(discoverymgrInfo.orchestrationMap[discoverymgrInfo.deviceID].ServiceList) != 0 {
				t.Error("Reset failed")
			}
		})
	})

	closeTest()
}
func TestServiceNameChecker(t *testing.T) {

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

	discoverymgrInfo.shutdownChan = make(chan struct{})
	defer close(discoverymgrInfo.shutdownChan)
	//set mock in order
	ipsub := make(chan []net.IP, 1)
	mockNetwork.EXPECT().AppendSubscriber().Return(ipsub)
	mockWrapper.EXPECT().ResetServer(gomock.Any()).Return()
	go detectNetworkChgRoutine()
	t.Run("Success", func(t *testing.T) {
		discoverymgrInfo.deviceID = "foo"
		ipsub <- []net.IP{net.ParseIP("192.0.2.1")}
	})
	time.Sleep(1 * time.Second)
	t.Run("Fail", func(t *testing.T) {
		discoverymgrInfo.deviceID = ""
		ipsub <- []net.IP{net.ParseIP("192.0.2.1")}
	})
}

func TestGetDeviceID(t *testing.T) {
	t.Run("SuccessNewV4", func(t *testing.T) {
		uuid, _ := getDeviceID("/x/y/z/NoFileIsThisName")
		if len(uuid) == 0 {
			t.Error()
		}
	})
}
