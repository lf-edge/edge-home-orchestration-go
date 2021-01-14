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

package notification

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/networkhelper"
	clientMocks "github.com/lf-edge/edge-home-orchestration-go/src/restinterface/client/mocks"

	"github.com/golang/mock/gomock"
)

var (
	targetLocalAddr, _ = networkhelper.GetInstance().GetOutboundIP()
	targetRemoteAddr   = "127.0.0.1"
	targetCustomURL    = fmt.Sprintf("127.0.0.1:%d", 56001)
	status             = "Finished"
	id                 = uint64(123)
)

func TestInvokeNotificationOnLocal(t *testing.T) {
	notiChan := make(chan string, 1)

	GetInstance().AddNotificationChan(id, notiChan)
	err := GetInstance().InvokeNotification(targetLocalAddr, float64(id), status)

	if err != nil {
		t.Fail()
	}
}

func TestInvokeNotificationFailedWithInvalidChan(t *testing.T) {
	err := GetInstance().InvokeNotification(targetLocalAddr, float64(id), status)
	if err == nil {
		t.Fail()
	}
}

func TestInvokeNotificationOnRemote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := clientMocks.NewMockClienter(ctrl)

	notiChan := make(chan string, 1)

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("TestInvokeNotificationOnRemote Handler")
		log.Println(r.RequestURI)

		if strings.Contains(r.RequestURI, "/api/v1/servicemgr/services/notification/") == false {
			t.Error()
		}
	}))
	l, _ := net.Listen("tcp", targetCustomURL)

	server.Listener = l
	server.Start()

	GetInstance().AddNotificationChan(id, notiChan)
	GetInstance().Clienter = mockClient
	mockClient.EXPECT().DoNotifyAppStatusRemoteDevice(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	err := GetInstance().InvokeNotification(targetRemoteAddr, float64(id), status)

	if err != nil {
		t.Fail()
	}

	server.Close()
}
