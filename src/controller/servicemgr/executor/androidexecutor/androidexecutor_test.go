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

package androidexecutor

import (
	"testing"

	"controller/servicemgr/executor"
	notificationMock "controller/servicemgr/notification/mocks"
	clientApiMock "restinterface/client/mocks"

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

	originAdbPath := adbPath
	adbPath = "echo"

	err := tExecutor.Execute(s)

	adbPath = originAdbPath

	if err != nil {
		t.Error()
	}
}
