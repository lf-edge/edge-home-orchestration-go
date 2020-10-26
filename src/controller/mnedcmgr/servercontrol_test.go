/*******************************************************************************
 * Copyright 2020 Samsung Electronics All Rights Reserved.
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

package mnedcmgr

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	networkmocks "common/networkhelper/mocks"
	"controller/mnedcmgr/server"
	serverMocks "controller/mnedcmgr/server/mocks"
	ciphermock "restinterface/cipher/mocks"
	helpermock "restinterface/resthelper/mocks"

	"github.com/golang/mock/gomock"
)

var (
	defaultOutboundIP      = "2.2.2.2"
	defaultClientDeviceID  = "clientdummyID"
	clientDefaultVirtualIP = "3.3.3.3"
	clientDefaultPrivateIP = "4.4.4.4"
	anotherClientDeviceID  = "clientAnotherdummyID"
	clientAnotherVirtualIP = "5.5.5.5"
	clientAnotherPrivateIP = "6.6.6.6"
	defaultMessage         = "dummy"
	mockMnedcServer        *serverMocks.MockMNEDCServer
	mockNetwork            *networkmocks.MockNetwork
)

func init() {

}

func TestStartMNEDCServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createServerMockIns(ctrl)
	err := createDeviceIDFileServer()
	if err != nil {
		log.Println("Could not create id file")
		return
	}

	t.Run("ServerError", func(t *testing.T) {
		s := GetServerInstance()
		mockMnedcServer.EXPECT().CreateServer(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New(""))
		s.StartMNEDCServer(defaultDeviceIDFilePath)
	})
	t.Run("TestWaitInterrupt", func(t *testing.T) {
		fatalErrChan := make(chan error)
		mockMnedcServer.EXPECT().Close().Return(nil).AnyTimes()
		go serverWaitInterrupt(fatalErrChan)
		time.Sleep(2 * time.Second)

		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	})
	t.Run("GetOutboundIPError", func(t *testing.T) {
		s := GetServerInstance()
		mockMnedcServer.EXPECT().CreateServer(gomock.Any(), gomock.Any(), gomock.Any()).Return(&server.Server{}, nil)
		mockMnedcServer.EXPECT().Run()
		mockNetwork.EXPECT().GetOutboundIP().Return("", errors.New(""))
		s.StartMNEDCServer(defaultDeviceIDFilePath)
	})
	t.Run("Success", func(t *testing.T) {
		s := GetServerInstance()
		mockMnedcServer.EXPECT().CreateServer(gomock.Any(), gomock.Any(), gomock.Any()).Return(&server.Server{}, nil)
		mockMnedcServer.EXPECT().Run()
		mockNetwork.EXPECT().GetOutboundIP().Return(defaultOutboundIP, nil)
		mockMnedcServer.EXPECT().SetClientIP(gomock.Any(), gomock.Any(), gomock.Any())
		s.StartMNEDCServer(defaultDeviceIDFilePath)
	})

	deleteDeviceIDFileServer()
}

func TestRequestHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createServerMockIns(ctrl)

	ipInfo := server.IPTypes{
		PrivateIP: clientDefaultPrivateIP,
		VirtualIP: clientDefaultVirtualIP,
	}

	ipInfoTwo := server.IPTypes{
		PrivateIP: clientAnotherPrivateIP,
		VirtualIP: clientAnotherVirtualIP,
	}

	clientInfoMap := map[string]server.IPTypes{}
	clientInfoMap[defaultClientDeviceID] = ipInfo

	requestJSON := make(map[string]interface{})
	requestJSON["DeviceID"] = anotherClientDeviceID
	requestJSON["VirtualIP"] = clientAnotherVirtualIP
	requestJSON["PrivateIP"] = clientAnotherPrivateIP

	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)
	helper = mockHelper

	mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(requestJSON, nil)
	mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusOK))
	mockMnedcServer.EXPECT().SetClientIP(gomock.Any(), gomock.Any(), gomock.Any()).Return()
	mockMnedcServer.EXPECT().GetClientIPMap().Return(clientInfoMap)
	mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return([]byte(defaultMessage), nil).AnyTimes()
	mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return("").AnyTimes()
	mockHelper.EXPECT().DoPost(gomock.Any(), gomock.Any()).Return([]byte(defaultMessage), 200, nil).AnyTimes()

	data, err := json.Marshal(ipInfoTwo)
	if err != nil {
		log.Println(logPrefix, "Cannot Marshal request Data", err.Error())
		return
	}

	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleClientInfo)
	GetServerInstance().SetCipher(mockCipher)
	handler.ServeHTTP(rr, req)
	deleteDeviceIDFileServer()
}

func createDeviceIDFileServer() error {
	f, err := os.Create(defaultDeviceIDFilePath)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(defaultDeviceID)
	if err != nil {
		return err
	}

	f.Sync()
	return nil
}

func deleteDeviceIDFileServer() {
	err := os.Remove(defaultDeviceIDFilePath)
	if err != nil {
		log.Println("Could not delete file")
	}
}

func createServerMockIns(ctrl *gomock.Controller) {
	mockNetwork = networkmocks.NewMockNetwork(ctrl)
	mockMnedcServer = serverMocks.NewMockMNEDCServer(ctrl)
	mnedcServerIns = mockMnedcServer
	networkIns = mockNetwork
}
