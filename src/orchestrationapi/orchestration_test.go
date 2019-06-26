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

package orchestrationapi

import (
	"testing"

	"github.com/golang/mock/gomock"

	"errors"

	"common/types/configuremgrtypes"

	networkmocks "common/networkhelper/mocks"
	contextmgrmocks "controller/configuremgr/mocks"
	discoverymocks "controller/discoverymgr/mocks"
	scoringmocks "controller/scoringmgr/mocks"
	executormocks "controller/servicemgr/executor/mocks"
	servicemocks "controller/servicemgr/mocks"
	clientmocks "restinterface/client/mocks"
)

var (
	mockWatcher   *contextmgrmocks.MockWatcher
	mockDiscovery *discoverymocks.MockDiscovery
	mockScoring   *scoringmocks.MockScoring
	mockService   *servicemocks.MockServiceMgr
	mockExecutor  *executormocks.MockServiceExecutor
	mockClient    *clientmocks.MockClientApi
	mockNetwork   *networkmocks.MockNetwork
)

func createMockIns(ctrl *gomock.Controller) {
	mockWatcher = contextmgrmocks.NewMockWatcher(ctrl)
	mockDiscovery = discoverymocks.NewMockDiscovery(ctrl)
	mockScoring = scoringmocks.NewMockScoring(ctrl)
	mockService = servicemocks.NewMockServiceMgr(ctrl)
	mockExecutor = executormocks.NewMockServiceExecutor(ctrl)
	mockClient = clientmocks.NewMockClientApi(ctrl)
	mockNetwork = networkmocks.NewMockNetwork(ctrl)
}

func getOcheIns(ctrl *gomock.Controller) Orche {
	var builder OrchestrationBuilder

	builder.SetDiscovery(mockDiscovery)
	builder.SetExecutor(mockExecutor)
	builder.SetScoring(mockScoring)
	builder.SetService(mockService)
	builder.SetWatcher(mockWatcher)
	builder.SetClient(mockClient)

	orche := builder.Build()
	orche.(*orcheImpl).networkhelper = mockNetwork

	return orche
}

func TestBuild(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("Success", func(t *testing.T) {
		mockService.EXPECT().SetLocalServiceExecutor(mockExecutor)

		var builder OrchestrationBuilder

		builder.SetDiscovery(mockDiscovery)
		builder.SetExecutor(mockExecutor)
		builder.SetScoring(mockScoring)
		builder.SetService(mockService)
		builder.SetWatcher(mockWatcher)
		builder.SetClient(mockClient)

		if builder.Build() == nil {
			t.Error("ochestration object is nil, expected is not nil")
		}
	})
	t.Run("Error", func(t *testing.T) {
		t.Run("ExcludeDiscovery", func(t *testing.T) {
			var builder OrchestrationBuilder

			builder.SetExecutor(mockExecutor)
			builder.SetScoring(mockScoring)
			builder.SetService(mockService)
			builder.SetWatcher(mockWatcher)
			builder.SetClient(mockClient)

			if builder.Build() != nil {
				t.Error("ochestration object is not nil, expected is nil")
			}
		})
		t.Run("ExcludeExecutor", func(t *testing.T) {
			var builder OrchestrationBuilder

			builder.SetDiscovery(mockDiscovery)
			builder.SetScoring(mockScoring)
			builder.SetService(mockService)
			builder.SetWatcher(mockWatcher)
			builder.SetClient(mockClient)

			if builder.Build() != nil {
				t.Error("ochestration object is not nil, expected is nil")
			}
		})
		t.Run("ExcludeScoring", func(t *testing.T) {
			var builder OrchestrationBuilder

			builder.SetDiscovery(mockDiscovery)
			builder.SetExecutor(mockExecutor)
			builder.SetService(mockService)
			builder.SetWatcher(mockWatcher)
			builder.SetClient(mockClient)

			if builder.Build() != nil {
				t.Error("ochestration object is not nil, expected is nil")
			}
		})
		t.Run("ExcludeService", func(t *testing.T) {
			var builder OrchestrationBuilder

			builder.SetDiscovery(mockDiscovery)
			builder.SetExecutor(mockExecutor)
			builder.SetScoring(mockScoring)
			builder.SetWatcher(mockWatcher)
			builder.SetClient(mockClient)

			if builder.Build() != nil {
				t.Error("ochestration object is not nil, expected is nil")
			}
		})
		t.Run("ExcludeWatcher", func(t *testing.T) {
			var builder OrchestrationBuilder

			builder.SetDiscovery(mockDiscovery)
			builder.SetExecutor(mockExecutor)
			builder.SetScoring(mockScoring)
			builder.SetService(mockService)
			builder.SetClient(mockClient)

			if builder.Build() != nil {
				t.Error("ochestration object is not nil, expected is nil")
			}
		})
		t.Run("ExcludeClientApi", func(t *testing.T) {
			var builder OrchestrationBuilder

			builder.SetDiscovery(mockDiscovery)
			builder.SetExecutor(mockExecutor)
			builder.SetScoring(mockScoring)
			builder.SetService(mockService)
			builder.SetWatcher(mockWatcher)

			if builder.Build() != nil {
				t.Error("ochestration object is not nil, expected is nil")
			}
		})
	})
}

func TestStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("Success", func(t *testing.T) {
		deviceIDPath := "/etc/"
		platform := "linux"
		executionType := "container"

		gomock.InOrder(
			mockService.EXPECT().SetLocalServiceExecutor(mockExecutor),
			mockDiscovery.EXPECT().StartDiscovery(gomock.Eq(deviceIDPath), gomock.Eq(platform), gomock.Eq(executionType)),
			mockWatcher.EXPECT().Watch(gomock.Any()),
		)

		getOcheIns(ctrl).Start(deviceIDPath, platform, executionType)
	})
}
func TestNotify(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("Success", func(t *testing.T) {
		gomock.InOrder(
			mockService.EXPECT().SetLocalServiceExecutor(mockExecutor),
			mockScoring.EXPECT().AddScoring(gomock.Any()).Return(nil),
			mockDiscovery.EXPECT().AddNewServiceName(gomock.Any()).Return(nil),
		)

		getOcheIns(ctrl)
		getOrcheImple().Ready = true
		api, err := GetInternalAPI()
		if err != nil {
			t.Error("unexpected error " + err.Error())
		}
		api.Notify(configuremgrtypes.ServiceInfo{})
	})
	t.Run("Error", func(t *testing.T) {
		t.Run("AddScoring", func(t *testing.T) {
			gomock.InOrder(
				mockService.EXPECT().SetLocalServiceExecutor(mockExecutor),
				mockScoring.EXPECT().AddScoring(gomock.Any()).Return(errors.New("error test")),
			)
			getOcheIns(ctrl)
			getOrcheImple().Ready = true
			api, err := GetInternalAPI()
			if err != nil {
				t.Error("unexpected error " + err.Error())
			}
			api.Notify(configuremgrtypes.ServiceInfo{})
		})
		t.Run("AddNewServiceName", func(t *testing.T) {
			testName := "test"
			testInfo := configuremgrtypes.ServiceInfo{ServiceName: testName}
			gomock.InOrder(
				mockService.EXPECT().SetLocalServiceExecutor(mockExecutor),
				mockScoring.EXPECT().AddScoring(gomock.Eq(testInfo)).Return(nil),
				mockDiscovery.EXPECT().AddNewServiceName(gomock.Any()).Return(errors.New("error test")),
				mockScoring.EXPECT().RemoveScoring(gomock.Eq(testName)).Return(nil),
			)
			getOcheIns(ctrl)
			getOrcheImple().Ready = true
			api, err := GetInternalAPI()
			if err != nil {
				t.Error("unexpected error " + err.Error())
			}
			api.Notify(testInfo)
		})
	})
}
