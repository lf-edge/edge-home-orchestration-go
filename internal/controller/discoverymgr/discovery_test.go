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
	"net"
	"reflect"
	"testing"
	"time"

	errormsg "github.com/lf-edge/edge-home-orchestration-go/internal/common/errormsg"
	errors "github.com/lf-edge/edge-home-orchestration-go/internal/common/errors"
	networkmocks "github.com/lf-edge/edge-home-orchestration-go/internal/common/networkhelper/mocks"
	wrapper "github.com/lf-edge/edge-home-orchestration-go/internal/controller/discoverymgr/wrapper"
	wrappermocks "github.com/lf-edge/edge-home-orchestration-go/internal/controller/discoverymgr/wrapper/mocks"
	systemdb "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/system"
	goerror "errors"
	clientMocks "github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/client/mocks"

	dbwrapper "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/wrapper"

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
	defaultConfAlternate = "testdata/client.config"

	anotherService     = "docker"
	anotherIPv4        = "2.2.2.2"
	anotherDeviceID    = "edge-orchestration-test-device-id2"
	anotherIPv4List    = []string{anotherIPv4}
	anotherServiceList = []string{anotherService}

	defaultVirtualIP = "10.10.10.10"
	anotherVirtualIP = "10.10.10.11"

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

	log.Println(logPrefix, "[addDevice]", deviceID)
	setSystemDB(deviceID, defaultPlatform, defaultExecutionType)
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
	_, err := sysQuery.Get(systemdb.ID)
	if err != nil {
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
	systemNames := []string{systemdb.ID, systemdb.Platform, systemdb.ExecType}

	for _, info := range systemNames {
		err := sysQuery.Delete(info)
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

func init() {
	//should remove bolt db file when finished unittest
	dbwrapper.SetBoltDBPath("./testDB")
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

		discoveryInstance := GetInstance()
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

			deviceID, err := discoveryInstance.GetDeviceID()
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

			deviceID, err := discoveryInstance.GetDeviceID()
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

			deviceID, err := discoveryInstance.GetDeviceID()
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

			deviceID, err := discoveryInstance.GetDeviceID()
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

	discoveryInstance := GetInstance()

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
		deviceID, err := discoveryInstance.GetDeviceID()
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

func TestAddDeviceInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	discoveryInstance := GetInstance()

	t.Run("ClienterNil", func(t *testing.T) {
		discoveryInstance.AddDeviceInfo(defaultMyDeviceID, defaultVirtualIP, defaultIPv4)
	})

	closeTest()

}

func TestAddDeviceInfoRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	mockClient := clientMocks.NewMockClienter(ctrl)
	discoveryInstance := GetInstance()

	t.Run("CallRequestdeviceInfo", func(t *testing.T) {
		discoveryInstance.SetClient(mockClient)
		discoveryInstance.SetRestResource()
		mockClient.EXPECT().DoGetOrchestrationInfo(gomock.Any()).Return(defaultPlatform, defaultExecutionType, defaultServiceList, nil).AnyTimes()
		discoveryInstance.AddDeviceInfo(defaultMyDeviceID, defaultVirtualIP, defaultIPv4)
		time.Sleep(2 * time.Second)
	})

	closeTest()

}

func TestAddDeviceInfoRestError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	mockClient := clientMocks.NewMockClienter(ctrl)
	discoveryInstance := GetInstance()

	t.Run("RestError", func(t *testing.T) {
		discoveryInstance.SetClient(mockClient)
		discoveryInstance.SetRestResource()
		mockClient.EXPECT().DoGetOrchestrationInfo(gomock.Any()).Return(defaultPlatform, defaultExecutionType, defaultServiceList, goerror.New("Dummy Error")).AnyTimes()

		discoveryInstance.AddDeviceInfo(defaultMyDeviceID, defaultVirtualIP, defaultIPv4)
		time.Sleep(7 * time.Second)
	})

	closeTest()

}

func TestMNEDCReconciledCallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("ReestablishedTest", func(t *testing.T) {
		discoveryInstance := GetInstance()
		mockNetwork.EXPECT().GetVirtualIP().Return("", goerror.New("No Virtual IP"))
		discoveryInstance.MNEDCReconciledCallback()
	})
	closeTest()
}

func TestGetOrchestrationInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	addDevice(true)

	discoveryInstance := GetInstance()
	platform, executionType, serviceList, err := discoveryInstance.GetOrchestrationInfo()

	if err != nil {
		t.Error("Error should not be thrown")
		return
	}
	if platform != defaultPlatform {
		t.Error("Platform incorrect")
		return
	}
	if executionType != defaultExecutionType {
		t.Error("Execution Type incorrect")
		return
	}
	if !reflect.DeepEqual(serviceList, defaultServiceList) {
		t.Error("Service List incorrect")
		return
	}
	closeTest()
}

func TestNotifyMNEDCBroadcastServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	addDevice(false)
	discoveryInstance := GetInstance()
	mockClient := clientMocks.NewMockClienter(ctrl)

	t.Run("VirtualIPError", func(t *testing.T) {
		mockNetwork.EXPECT().GetVirtualIP().Return("", goerror.New("No Virtual IP"))

		err := discoveryInstance.NotifyMNEDCBroadcastServer()
		if err == nil {
			t.Error("Error should not be nil")
			return
		}

	})
	t.Run("OutboundIPError", func(t *testing.T) {
		mockNetwork.EXPECT().GetVirtualIP().Return(defaultIPv4, nil)
		mockNetwork.EXPECT().GetOutboundIP().Return("", goerror.New("No outbound IP"))

		err := discoveryInstance.NotifyMNEDCBroadcastServer()
		if err == nil {
			t.Error("Error should not be nil")
			return
		}

	})
	t.Run("Success", func(t *testing.T) {
		configAlternate = defaultConfAlternate
		discoveryInstance.SetClient(mockClient)
		discoveryInstance.SetRestResource()
		mockNetwork.EXPECT().GetVirtualIP().Return(defaultVirtualIP, nil)
		mockNetwork.EXPECT().GetOutboundIP().Return(defaultIPv4, nil)
		mockClient.EXPECT().DoNotifyMNEDCBroadcastServer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		err := discoveryInstance.NotifyMNEDCBroadcastServer()
		if err != nil {
			t.Error("Error should be nil")
			return
		}

	})
	closeTest()
}

func TestGetDeviceID(t *testing.T) {
	t.Run("SuccessNewV4", func(t *testing.T) {
		uuid, _ := setDeviceID("/x/y/z/NoFileIsThisName")
		if len(uuid) == 0 {
			t.Error()
		}
	})
}
