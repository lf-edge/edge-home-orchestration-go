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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	networkmocks "common/networkhelper/mocks"
	resourceutilmocks "common/resourceutil/mocks"
	contextmgrmocks "controller/configuremgr/mocks"
	discoverymocks "controller/discoverymgr/mocks"
	scoringmocks "controller/scoringmgr/mocks"
	executormocks "controller/servicemgr/executor/mocks"
	servicemocks "controller/servicemgr/mocks"
	dbsystemMocks "db/bolt/system/mocks"
	dbhelpermocks "db/helper/mocks"
	clientmocks "restinterface/client/mocks"
)

const (
	defaultServiceName = "default_service"
)

var (
	mockWatcher      *contextmgrmocks.MockWatcher
	mockDiscovery    *discoverymocks.MockDiscovery
	mockScoring      *scoringmocks.MockScoring
	mockService      *servicemocks.MockServiceMgr
	mockExecutor     *executormocks.MockServiceExecutor
	mockDBHelper     *dbhelpermocks.MockMultipleBucketQuery
	mockClient       *clientmocks.MockClienter
	mockNetwork      *networkmocks.MockNetwork
	mockResourceutil *resourceutilmocks.MockMonitor

	mockSystemDBExecutor *dbsystemMocks.MockDBInterface
)

func createMockIns(ctrl *gomock.Controller) {
	mockWatcher = contextmgrmocks.NewMockWatcher(ctrl)
	mockDiscovery = discoverymocks.NewMockDiscovery(ctrl)
	mockScoring = scoringmocks.NewMockScoring(ctrl)
	mockService = servicemocks.NewMockServiceMgr(ctrl)
	mockExecutor = executormocks.NewMockServiceExecutor(ctrl)
	mockDBHelper = dbhelpermocks.NewMockMultipleBucketQuery(ctrl)
	mockClient = clientmocks.NewMockClienter(ctrl)
	mockNetwork = networkmocks.NewMockNetwork(ctrl)
	mockResourceutil = resourceutilmocks.NewMockMonitor(ctrl)
	mockSystemDBExecutor = dbsystemMocks.NewMockDBInterface(ctrl)
}

func getOcheIns(ctrl *gomock.Controller) Orche {
	var builder OrchestrationBuilder

	builder.SetDiscovery(mockDiscovery)
	builder.SetExecutor(mockExecutor)
	builder.SetScoring(mockScoring)
	builder.SetService(mockService)
	builder.SetWatcher(mockWatcher)
	builder.SetClient(mockClient)

	helper = mockDBHelper
	sysDBExecutor = mockSystemDBExecutor

	orche := builder.Build()
	resourceMonitorImpl = mockResourceutil
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
			mockResourceutil.EXPECT().StartMonitoringResource(),
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
			mockDiscovery.EXPECT().AddNewServiceName(gomock.Any()).Return(nil),
		)

		getOcheIns(ctrl)
		getOrcheImple().Ready = true
		api, err := GetInternalAPI()
		if err != nil {
			t.Error("unexpected error " + err.Error())
		}
		api.Notify(defaultServiceName)
	})
	t.Run("Error", func(t *testing.T) {
		t.Run("AddNewServiceName", func(t *testing.T) {
			gomock.InOrder(
				mockService.EXPECT().SetLocalServiceExecutor(mockExecutor),
				mockDiscovery.EXPECT().AddNewServiceName(gomock.Any()).Return(errors.New("error test")),
			)
			getOcheIns(ctrl)
			getOrcheImple().Ready = true
			api, err := GetInternalAPI()
			if err != nil {
				t.Error("unexpected error " + err.Error())
			}
			api.Notify(defaultServiceName)
		})
	})
}
