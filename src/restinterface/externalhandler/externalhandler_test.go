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

package externalhandler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	networkhelper "github.com/lf-edge/edge-home-orchestration-go/src/common/networkhelper/mocks"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/requestervalidator"
	orchestrationapi "github.com/lf-edge/edge-home-orchestration-go/src/orchestrationapi"
	orchemock "github.com/lf-edge/edge-home-orchestration-go/src/orchestrationapi/mocks"
	ciphermock "github.com/lf-edge/edge-home-orchestration-go/src/restinterface/cipher/mocks"
	helpermock "github.com/lf-edge/edge-home-orchestration-go/src/restinterface/resthelper/mocks"

	"github.com/golang/mock/gomock"
)

func TestGetHandler(t *testing.T) {
	handler := GetHandler()
	if handler == nil {
		t.Error("unexpected return value")
	}
}

func TestSetOrchestrationAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := GetHandler()
	if handler == nil {
		t.Error("unexpected return value")
	}

	t.Run("Error", func(t *testing.T) {
		t.Run("WrongDefaultFlag", func(t *testing.T) {
			errorHandler := GetHandler()
			if errorHandler.isSetAPI == true {
				t.Error("unexpected flag value")
			}
		})
	})
	t.Run("Success", func(t *testing.T) {
		orcheExternalAPIMock := orchemock.NewMockOrcheExternalAPI(ctrl)

		handler.SetOrchestrationAPI(orcheExternalAPIMock)
		if handler.api != orcheExternalAPIMock {
			t.Error("unexpected difference of copied value")
		}
		if handler.isSetAPI == false {
			t.Error("unexpected flag value")
		}
	})
}

func getReqeustArgs() (orchestrationapi.ReqeustService, map[string]interface{}) {
	args := []string{"-a", "-b"}

	serviceName := "test"
	serviceInfo := make([]orchestrationapi.RequestServiceInfo, 1)
	serviceInfo[0].ExecutionType = "native"
	serviceInfo[0].ExeCmd = args
	requestService := orchestrationapi.ReqeustService{
		ServiceName:      serviceName,
		SelfSelection:    true,
		ServiceRequester: "test",
		ServiceInfo:      serviceInfo,
	}

	requestervalidator.RequesterValidator{}.StoreRequesterInfo(serviceName, []string{"test"})

	execCmd := make([]interface{}, len(args))
	for idx, arg := range args {
		execCmd[idx] = arg
	}

	sInfo := make(map[string]interface{})
	sInfo["ExecutionType"] = "native"
	sInfo["ExecCmd"] = execCmd

	sInfos := make([]interface{}, 1)
	sInfos[0] = sInfo

	appCommand := make(map[string]interface{})
	appCommand["ServiceName"] = serviceName
	appCommand["ServiceInfo"] = sInfos
	appCommand["ServiceRequester"] = "test"

	return requestService, appCommand
}

func getInvalidParamResponse() map[string]interface{} {
	response := make(map[string]interface{})
	response["Message"] = "INVALID_PARAMETER"
	response["ServiceName"] = "test"

	targetInfo := make(map[string]interface{})
	targetInfo["ExecutionType"] = "native"
	targetInfo["Target"] = "0.0.0.0"
	response["RemoteTargetInfo"] = targetInfo

	return response
}

func TestAPIV1RequestServicePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := GetHandler()
	if handler == nil {
		t.Error("unexpected return value")
	}

	mockOrchestration := orchemock.NewMockOrcheExternalAPI(ctrl)
	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)
	mockNetHelper := networkhelper.NewMockNetwork(ctrl)

	r := httptest.NewRequest("POST", "http://localhost:1234", nil)
	w := httptest.NewRecorder()

	addr := strings.Split(r.RemoteAddr, ":")[0]

	t.Run("Error", func(t *testing.T) {
		t.Run("IsNotSetApi", func(t *testing.T) {
			handler.setHelper(mockHelper)
			mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable))

			handler.isSetAPI = false
			handler.APIV1RequestServicePost(w, r)
		})
		t.Run("IsNotSetKey", func(t *testing.T) {
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)
			mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable))

			handler.IsSetKey = false
			handler.APIV1RequestServicePost(w, r)
		})
		t.Run("DecryptionFail", func(t *testing.T) {
			handler.SetCipher(mockCipher)
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)
			handler.netHelper = mockNetHelper
			gomock.InOrder(
				mockNetHelper.EXPECT().GetIPs().Return([]string{addr}, nil),
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(nil, errors.New("")),
				mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable)),
			)

			handler.APIV1RequestServicePost(w, r)
		})
		t.Run("EncryptionFail", func(t *testing.T) {
			handler.SetCipher(mockCipher)
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)
			handler.netHelper = mockNetHelper

			requestService, appCommand := getReqeustArgs()

			gomock.InOrder(
				mockNetHelper.EXPECT().GetIPs().Return([]string{addr}, nil),
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(appCommand, nil),
				mockOrchestration.EXPECT().RequestService(gomock.Eq(requestService)),
				mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, errors.New("")),
				mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable)),
			)

			handler.APIV1RequestServicePost(w, r)
		})
		t.Run("InvalidParam", func(t *testing.T) {
			t.Run("ServiceName", func(t *testing.T) {
				handler.SetCipher(mockCipher)
				handler.SetOrchestrationAPI(mockOrchestration)
				handler.setHelper(mockHelper)
				handler.netHelper = mockNetHelper

				_, appCommand := getReqeustArgs()
				delete(appCommand, "ServiceName")

				gomock.InOrder(
					mockNetHelper.EXPECT().GetIPs().Return([]string{addr}, nil),
					mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(appCommand, nil),
					mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Do(func(resp map[string]interface{}) {
						if resp["Message"] != orchestrationapi.INVALID_PARAMETER {
							t.Error("unexpected response")
						}
					}).Return(nil, nil),
					mockHelper.EXPECT().ResponseJSON(gomock.Any(), gomock.Any(), gomock.Eq(http.StatusOK)),
				)

				handler.APIV1RequestServicePost(w, r)
			})
			t.Run("ServiceInfo", func(t *testing.T) {
				handler.SetCipher(mockCipher)
				handler.SetOrchestrationAPI(mockOrchestration)
				handler.setHelper(mockHelper)
				handler.netHelper = mockNetHelper

				_, appCommand := getReqeustArgs()
				delete(appCommand, "ServiceInfo")

				gomock.InOrder(
					mockNetHelper.EXPECT().GetIPs().Return([]string{addr}, nil),
					mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(appCommand, nil),
					mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Do(func(resp map[string]interface{}) {
						if resp["Message"] != orchestrationapi.INVALID_PARAMETER {
							t.Error("unexpected response")
						}
					}).Return(nil, nil),
					mockHelper.EXPECT().ResponseJSON(gomock.Any(), gomock.Any(), gomock.Eq(http.StatusOK)),
				)

				handler.APIV1RequestServicePost(w, r)
			})
			t.Run("ExecutionType", func(t *testing.T) {
				handler.SetCipher(mockCipher)
				handler.SetOrchestrationAPI(mockOrchestration)
				handler.setHelper(mockHelper)
				handler.netHelper = mockNetHelper

				_, appCommand := getReqeustArgs()
				tmp := appCommand["ServiceInfo"].([]interface{})
				serviceInfo := tmp[0].(map[string]interface{})
				delete(serviceInfo, "ExecutionType")

				gomock.InOrder(
					mockNetHelper.EXPECT().GetIPs().Return([]string{addr}, nil),
					mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(appCommand, nil),
					mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Do(func(resp map[string]interface{}) {
						if resp["Message"] != orchestrationapi.INVALID_PARAMETER {
							t.Error("unexpected response")
						}
					}).Return(nil, nil),
					mockHelper.EXPECT().ResponseJSON(gomock.Any(), gomock.Any(), gomock.Eq(http.StatusOK)),
				)

				handler.APIV1RequestServicePost(w, r)
			})
			t.Run("ExecCmd", func(t *testing.T) {
				handler.SetCipher(mockCipher)
				handler.SetOrchestrationAPI(mockOrchestration)
				handler.setHelper(mockHelper)
				handler.netHelper = mockNetHelper

				_, appCommand := getReqeustArgs()
				tmp := appCommand["ServiceInfo"].([]interface{})
				serviceInfo := tmp[0].(map[string]interface{})
				delete(serviceInfo, "ExecCmd")

				gomock.InOrder(
					mockNetHelper.EXPECT().GetIPs().Return([]string{addr}, nil),
					mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(appCommand, nil),
					mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Do(func(resp map[string]interface{}) {
						if resp["Message"] != orchestrationapi.INVALID_PARAMETER {
							t.Error("unexpected response")
						}
					}).Return(nil, nil),
					mockHelper.EXPECT().ResponseJSON(gomock.Any(), gomock.Any(), gomock.Eq(http.StatusOK)),
				)

				handler.APIV1RequestServicePost(w, r)
			})
		})
	})

	t.Run("Success", func(t *testing.T) {
		handler.SetCipher(mockCipher)
		handler.SetOrchestrationAPI(mockOrchestration)
		handler.setHelper(mockHelper)
		handler.netHelper = mockNetHelper

		requestService, appCommand := getReqeustArgs()
		respByte := []byte{'1'}

		gomock.InOrder(
			mockNetHelper.EXPECT().GetIPs().Return([]string{addr}, nil),
			mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(appCommand, nil),
			mockOrchestration.EXPECT().RequestService(gomock.Eq(requestService)),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(respByte, nil),
			mockHelper.EXPECT().ResponseJSON(gomock.Any(), gomock.Eq(respByte), gomock.Eq(http.StatusOK)),
		)

		handler.APIV1RequestServicePost(w, r)
	})
}
