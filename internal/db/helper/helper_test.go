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

package helper

import (
	"github.com/golang/mock/gomock"
	"testing"

	"errors"

	configuration "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/configuration"
	dbConfigurationMocks "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/configuration/mocks"
	network "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/network"
	dbNetworkMocks "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/network/mocks"
	service "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/service"
	dbServiceMocks "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/service/mocks"
	system "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/system"
	dbSysMocks "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/system/mocks"
)

var (
	mockConf    *dbConfigurationMocks.MockDBInterface
	mockNet     *dbNetworkMocks.MockDBInterface
	mockService *dbServiceMocks.MockDBInterface
	mockSys     *dbSysMocks.MockDBInterface
)

func testInit(ctrl *gomock.Controller) func() {
	originConfQuery := confQuery
	originNetQuery := netQuery
	originServiceQuery := serviceQuery
	originSysQuery := sysQuery

	mockConf = dbConfigurationMocks.NewMockDBInterface(ctrl)
	mockNet = dbNetworkMocks.NewMockDBInterface(ctrl)
	mockService = dbServiceMocks.NewMockDBInterface(ctrl)
	mockSys = dbSysMocks.NewMockDBInterface(ctrl)

	confQuery = mockConf
	netQuery = mockNet
	serviceQuery = mockService
	sysQuery = mockSys

	return func() {
		confQuery = originConfQuery
		netQuery = originNetQuery
		serviceQuery = originServiceQuery
		sysQuery = originSysQuery
	}
}

func TestGetDeviceInfoWithService(t *testing.T) {
	ctrl := gomock.NewController(t)
	f := testInit(ctrl)
	defer ctrl.Finish()
	defer f()

	t.Run("Success", func(t *testing.T) {
		gomock.InOrder(
			mockConf.EXPECT().GetList().Return([]configuration.Configuration{
				{
					ID:       "test",
					Platform: "test",
					ExecType: "native",
				},
				{
					ID:       "test",
					Platform: "test",
					ExecType: "container",
				},
			}, nil),
			mockService.EXPECT().Get(gomock.Eq("test")).Return(service.Info{
				ID: "test",
				Services: []string{
					"testService1",
					"testService2",
				},
			}, nil),
			mockNet.EXPECT().Get(gomock.Eq("test")).Return(network.Info{
				ID:   "test",
				IPv4: []string{"1.1.2.1", "1.1.2.2"},
				RTT:  0.0,
			}, nil),
			mockNet.EXPECT().Get(gomock.Eq("test")).Return(network.Info{
				ID:   "test",
				IPv4: []string{"1.1.1.1", "1.1.1.2"},
				RTT:  0.0,
			}, nil),
		)

		ret, err := GetInstance().GetDeviceInfoWithService("testService1", []string{"native", "container"}, false)
		if err != nil {
			t.Error("unexpected error")
		} else if len(ret) != 2 {
			t.Error("unexpected return length")
		}

		for _, candidate := range ret {
			if candidate.ID != "test" {
				t.Error("unexpected service id")
			}
			if candidate.ExecType == "container" {
				if candidate.Endpoint[0] != "1.1.1.1" || candidate.Endpoint[1] != "1.1.1.2" {
					t.Error("unexpected endpoint of container")
				}
			} else if candidate.ExecType == "native" {
				if candidate.Endpoint[0] != "1.1.2.1" || candidate.Endpoint[1] != "1.1.2.2" {
					t.Error("unexpected endpoint of native")
				}
			}
		}
	})
	t.Run("Error", func(t *testing.T) {
		t.Run("confQueryGetList", func(t *testing.T) {
			mockConf.EXPECT().GetList().Return(nil, errors.New(""))

			ret, err := GetInstance().GetDeviceInfoWithService("", nil, false)
			if err == nil {
				t.Error("unexpected success")
			} else if ret != nil {
				t.Error("unexpected success")
			}
		})
		t.Run("noHasExecType", func(t *testing.T) {
			mockConf.EXPECT().GetList().Return([]configuration.Configuration{
				{
					ID:       "test",
					Platform: "test",
					ExecType: "native",
				},
			}, nil)

			ret, err := GetInstance().GetDeviceInfoWithService("", []string{"container"}, false)
			if err == nil {
				t.Error("unexpected success")
			} else if ret != nil {
				t.Error("unexpected success")
			}
		})
		t.Run("netQueryGet", func(t *testing.T) {
			gomock.InOrder(
				mockConf.EXPECT().GetList().Return([]configuration.Configuration{
					{
						ID:       "test",
						Platform: "test",
						ExecType: "container",
					},
				}, nil),
				mockNet.EXPECT().Get(gomock.Eq("test")).Return(network.Info{}, errors.New("")),
			)

			ret, err := GetInstance().GetDeviceInfoWithService("", []string{"container"}, false)
			if err == nil {
				t.Error("unexpected success")
			} else if ret != nil {
				t.Error("unexpected success")
			}
		})
		t.Run("serviceQueryGet", func(t *testing.T) {
			gomock.InOrder(
				mockConf.EXPECT().GetList().Return([]configuration.Configuration{
					{
						ID:       "test",
						Platform: "test",
						ExecType: "native",
					},
				}, nil),
				mockService.EXPECT().Get(gomock.Eq("test")).Return(service.Info{}, errors.New("")),
			)

			ret, err := GetInstance().GetDeviceInfoWithService("", []string{"native"}, false)
			if err == nil {
				t.Error("unexpected success")
			} else if ret != nil {
				t.Error("unexpected success")
			}
		})
		t.Run("noMatchServiceName", func(t *testing.T) {
			gomock.InOrder(
				mockConf.EXPECT().GetList().Return([]configuration.Configuration{
					{
						ID:       "test",
						Platform: "test",
						ExecType: "native",
					},
				}, nil),
				mockService.EXPECT().Get(gomock.Eq("test")).Return(service.Info{
					ID: "test1",
					Services: []string{
						"testService1",
						"testService2",
					},
				}, nil),
			)

			ret, err := GetInstance().GetDeviceInfoWithService("", []string{"native"}, false)
			if err == nil {
				t.Error("unexpected success")
			} else if ret != nil {
				t.Error("unexpected success")
			}
		})
		t.Run("NonContainerNetQueryGet", func(t *testing.T) {
			gomock.InOrder(
				mockConf.EXPECT().GetList().Return([]configuration.Configuration{
					{
						ID:       "test",
						Platform: "test",
						ExecType: "native",
					},
				}, nil),
				mockService.EXPECT().Get(gomock.Eq("test")).Return(service.Info{
					ID: "test",
					Services: []string{
						"testService1",
						"testService2",
					},
				}, nil),
				mockNet.EXPECT().Get(gomock.Eq("test")).Return(network.Info{}, errors.New("")),
			)

			ret, err := GetInstance().GetDeviceInfoWithService("testService1", []string{"native"}, false)
			if err == nil {
				t.Error("unexpected success")
			} else if ret != nil {
				t.Error("unexpected success")
			}
		})
	})
}

func TestGetDeviceID(t *testing.T) {
	ctrl := gomock.NewController(t)
	f := testInit(ctrl)
	defer ctrl.Finish()
	defer f()

	t.Run("Success", func(t *testing.T) {
		mockSys.EXPECT().Get(gomock.Eq("id")).Return(system.Info{
			Name:  "id",
			Value: "testID",
		}, nil)
		_, err := GetInstance().GetDeviceID()
		if err != nil {
			t.Error("unexpected error")
		}
	})
	t.Run("Fail", func(t *testing.T) {
		mockSys.EXPECT().Get(gomock.Eq("id")).Return(system.Info{
			Name:  "id",
			Value: "testID",
		}, errors.New(""))
		_, err := GetInstance().GetDeviceID()
		if err == nil {
			t.Error("unexpected success")
		}
	})
}
