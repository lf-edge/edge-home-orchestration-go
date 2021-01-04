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

package nativeexecutor

import (
	"testing"

	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/executor"
	notificationMock "github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/notification/mocks"
	clientApiMock "github.com/lf-edge/edge-home-orchestration-go/src/restinterface/client/mocks"

	"github.com/golang/mock/gomock"
)

func initializeMock(t *testing.T) (*notificationMock.MockNotification, *clientApiMock.MockClienter) {
	t.Helper()

	ctrl := gomock.NewController(t)
	noti := notificationMock.NewMockNotification(ctrl)
	client := clientApiMock.NewMockClienter(ctrl)

	return noti, client
}

func TestClient(t *testing.T) {
	tExecutor := GetInstance()

	noti, client := initializeMock(t)

	noti.EXPECT().SetClient(gomock.Any()).DoAndReturn(
		func(clientParam *clientApiMock.MockClienter) {
			if clientParam != client {
				t.Fail()
			}
		},
	)

	tExecutor.SetNotiImpl(noti)
	tExecutor.SetClient(client)
}

func TestExecute(t *testing.T) {
	tExecutor := GetInstance()
	noti, _ := initializeMock(t)

	gomock.InOrder(
		noti.EXPECT().InvokeNotification(gomock.Any(), gomock.Any(), gomock.Any()),
	)

	s := executor.ServiceExecutionInfo{ServiceID: uint64(1), ServiceName: "ls_service", ParamStr: []string{"ls", "-ail"}, NotificationTargetURL: ""}

	tExecutor.SetNotiImpl(noti)
	err := tExecutor.Execute(s)

	if err != nil {
		t.Error()
	}
}

func TestExecuteFailWithEmptyServiceCmd(t *testing.T) {
	tExecutor := GetInstance()

	noti, _ := initializeMock(t)
	gomock.InOrder(
		noti.EXPECT().InvokeNotification(gomock.Any(), gomock.Any(), gomock.Any()),
	)

	s := executor.ServiceExecutionInfo{ServiceID: uint64(1), ServiceName: "ls_service", NotificationTargetURL: ""}

	tExecutor.SetNotiImpl(noti)
	err := tExecutor.Execute(s)

	if err == nil {
		t.Error()
	}

}

func TestExecuteFailWithInvalidServiceName(t *testing.T) {
	tExecutor := GetInstance()

	ctrl := gomock.NewController(t)
	noti := notificationMock.NewMockNotification(ctrl)

	gomock.InOrder(
		noti.EXPECT().InvokeNotification(gomock.Any(), gomock.Any(), gomock.Any()),
	)

	s := executor.ServiceExecutionInfo{ServiceID: uint64(1), ServiceName: "InvalidService", ParamStr: []string{"invalid", "-ail"}, NotificationTargetURL: ""}

	tExecutor.SetNotiImpl(noti)
	err := tExecutor.Execute(s)

	if err == nil {
		t.Error()
	}
}

func TestExecuteFailWithInvalidServiceArgs(t *testing.T) {
	tExecutor := GetInstance()

	ctrl := gomock.NewController(t)
	noti := notificationMock.NewMockNotification(ctrl)

	gomock.InOrder(
		noti.EXPECT().InvokeNotification(gomock.Any(), gomock.Any(), gomock.Any()),
	)

	s := executor.ServiceExecutionInfo{ServiceID: uint64(1), ServiceName: "ls", ParamStr: []string{"ls", "InvalidArgs"}, NotificationTargetURL: ""}

	tExecutor.SetNotiImpl(noti)
	err := tExecutor.Execute(s)

	if err == nil {
		t.Error()
	}
}
