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

package servicemgr

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/networkhelper"
	executorMock "github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/executor/mocks"
	clientApiMock "github.com/lf-edge/edge-home-orchestration-go/src/restinterface/client/mocks"

	"github.com/golang/mock/gomock"
	//	"bytes"
	//	"encoding/json"
	//	"io/ioutil"
	//	"controller/servicemgr/notification"
)

const (
	serviceName  = "ls"
	serviceName2 = "main2"
	requester    = "test_request"
)

var (
	paramStr         = []interface{}{"ls"}
	paramStrWithArgs = []interface{}{"ls", "-ail"}
)

var (
	targetLocalAddr, _ = networkhelper.GetInstance().GetOutboundIP()
	targetRemoteAddr   = "127.0.0.1"
	targetCustomURL    = fmt.Sprintf("127.0.0.1:%d", 56001)
)

func TestExecuteAppOnLocal(t *testing.T) {
	serviceIns := GetInstance()
	ctrl := gomock.NewController(t)
	exec := executorMock.NewMockServiceExecutor(ctrl)

	gomock.InOrder(
		exec.EXPECT().SetClient(gomock.Any()),
		exec.EXPECT().Execute(gomock.Any()).Return(nil),
	)

	serviceIns.SetLocalServiceExecutor(exec)
	notiChan := make(chan string)

	ifArgs := make([]interface{}, len(paramStrWithArgs))
	for i, v := range paramStrWithArgs {
		ifArgs[i] = v
	}

	err := serviceIns.Execute(targetLocalAddr, serviceName, requester, ifArgs, notiChan)
	checkError(t, err)

	time.Sleep(time.Millisecond * 10)
}

func TestExecuteAppOnRemote(t *testing.T) {
	serviceIns := GetInstance()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	exec := executorMock.NewMockServiceExecutor(ctrl)
	client := clientApiMock.NewMockClienter(ctrl)

	exec.EXPECT().SetClient(gomock.Any()).DoAndReturn(
		func(clientMock *clientApiMock.MockClienter) {
			if clientMock != client {
				t.Fail()
			}
		},
	)
	client.EXPECT().DoExecuteRemoteDevice(gomock.Any(), gomock.Any()).Return(nil)

	serviceIns.Clienter = client
	serviceIns.SetLocalServiceExecutor(exec)
	notiChan := make(chan string)

	err := serviceIns.Execute(targetRemoteAddr, serviceName, requester, paramStrWithArgs, notiChan)
	checkError(t, err)
}

/**************** SERVICEMGR REST INIT TEST ***********************/
//func TestRestInit(t *testing.T) {
//	//for coverage
//	RestInit()
//}

///**************** SERVICEMGR REST CLIENT TEST ***********************/
//func TestServer(t *testing.T) {
//	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		fmt.Fprintln(w, "Hello, client")
//	}))
//	defer ts.Close()
//
//	res, err := http.Get(ts.URL)
//	if err != nil {
//		log.Fatal(err)
//	}
//	greeting, err := ioutil.ReadAll(res.Body)
//	res.Body.Close()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("%s", greeting)
//	// Output: Hello, client
//}
//
//func TestClientDoExecuteRemoteDevice(t *testing.T) {
//
//	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		log.Println("TestClientDoExecuteRemoteDevice Handler")
//		log.Println(r.RequestURI)
//
//		respJSONMsg := make(map[string]interface{})
//		respJSONMsg["Status"] = ConstServiceStatusStarted
//		respBytes := resthelper.MakeBinaryStream(respJSONMsg)
//		respEncryptBytes := resthelper.EncryptBytes(Key, respBytes) //NOTE : Key is in servicemgr_resttypes.go
//		resthelper.ResponseJSON(w, respEncryptBytes, http.StatusOK)
//
//	}))
//	// server := httptest.NewUnstartedServer(http.HandlerFunc(restserver.APIV1ServicemgrServicesPost))
//	l, _ := net.Listen("tcp", targetCustomURL)
//
//	server.Listener = l
//	server.Start()
//	defer server.Close()
//
//	appInfo := make(map[string]interface{})
//	appInfo["ServiceName"] = "ls"
//	appInfo["NotificationTargetURL"] = "127.0.0.1"
//
//	err := DoExecuteRemoteDevice(appInfo, "127.0.0.1")
//
//	if err != nil {
//		t.Fatal("err is not nil")
//	}
//}
//
//func TestClientDoNotifyAppStatusRemoteDevice(t *testing.T) {
//	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		log.Println("TestClientDoNotifyAppStatusRemoteDevice Handler")
//		log.Println(r.RequestURI)
//
//	}))
//	l, _ := net.Listen("tcp", targetCustomURL)
//
//	server.Listener = l
//	server.Start()
//	defer server.Close()
//
//	statusNotificationInfo := make(map[string]interface{})
//	statusNotificationInfo["ServiceID"] = 1
//	statusNotificationInfo["Status"] = "Failed"
//
//	err := DoNotifyAppStatusRemoteDevice(statusNotificationInfo, uint64(1), "127.0.0.1")
//
//	if err != nil {
//		t.Fatal("err is not nil")
//	}
//}
//
//// /**************** SERVICEMGR REST SERVER TEST ***********************/
//func TestServerRoutesInnerAPIV1ServicemgrServicesPost(t *testing.T) {
//
//	appInfo := make(map[string]interface{})
//	appInfo["ServiceID"] = 1
//	appInfo["ServiceName"] = "ls"
//	appInfo["NotificationTargetURL"] = "URL"
//	appInfo["UserArgs"] = []string{"-al"}
//	bodybytes, err := json.Marshal(appInfo)
//	buff := bytes.NewBuffer(bodybytes)
//
//	req, err := http.NewRequest("POST", "/api/v1/servicemgr/services", buff)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	resp := httptest.NewRecorder()
//	handler := http.HandlerFunc(APIV1ServicemgrServicesPost)
//
//	handler.ServeHTTP(resp, req)
//
//	if resp.Result().StatusCode != http.StatusOK {
//		t.Fatal("expected : StatusOK , ret : ", resp.Result().StatusCode)
//	}
//}
//
//func TestServerRoutesInnerAPIV1ServicemgrServicesNotificationServiceIDPost(t *testing.T) {
//
//	//init
//	id := uint64(1)
//	notiChan := make(chan string, 1)
//	notification.GetInstance().AddNotificationChan(id, notiChan)
//
//	message := make(map[string]interface{})
//	message["ServiceID"] = id
//	message["Status"] = "Failed"
//	bodybytes, err := json.Marshal(message)
//	buff := bytes.NewBuffer(bodybytes)
//
//	req, err := http.NewRequest("POST", "/api/v1/servicemgr/services/notification/1", buff)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	resp := httptest.NewRecorder()
//	handler := http.HandlerFunc(APIV1ServicemgrServicesNotificationServiceIDPost)
//
//	handler.ServeHTTP(resp, req)
//
//	if resp.Result().StatusCode != http.StatusOK {
//		t.Fatal("expected : StatusOK , ret : ", resp.Result().StatusCode)
//	}
//}

func assertEqualStr(t *testing.T, a, b string) {
	t.Helper()
	if strings.Compare(a, b) != 0 {
		t.Errorf("%s != %s", a, b)
	}
}

func checkError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Error(err.Error())
	}
}
