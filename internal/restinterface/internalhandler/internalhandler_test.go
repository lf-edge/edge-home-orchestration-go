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

package internalhandler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"testing"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/commandvalidator"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/requestervalidator"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/types/configuremgrtypes"
	orchemock "github.com/lf-edge/edge-home-orchestration-go/internal/orchestrationapi/mocks"
	ciphermock "github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/cipher/mocks"
	helpermock "github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/resthelper/mocks"

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
		orcheExternalAPIMock := orchemock.NewMockOrcheInternalAPI(ctrl)

		handler.SetOrchestrationAPI(orcheExternalAPIMock)
		if handler.api != orcheExternalAPIMock {
			t.Error("unexpected difference of copied value")
		}
		if handler.isSetAPI == false {
			t.Error("unexpected flag value")
		}
	})
}

func TestAPIV1ServicemgrServicesPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := GetHandler()
	if handler == nil {
		t.Error("unexpected return value")
	}

	mockOrchestration := orchemock.NewMockOrcheInternalAPI(ctrl)
	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)

	r := httptest.NewRequest("POST", "http://test.test", nil)
	w := httptest.NewRecorder()

	req := make(map[string]interface{})
	req["ServiceName"] = "test_service"
	req["ServiceID"] = 1.0
	req["Requester"] = "test_requester"
	args := make([]interface{}, 2)
	args[0] = "test_execute"
	args[1] = "test_executetype"
	req["UserArgs"] = args

	serviceInfo := configuremgrtypes.ServiceInfo{
		ServiceName:        "test_service",
		ExecutableFileName: "test_execute",
		AllowedRequester:   []string{"test_requester"},
	}

	commandvalidator.CommandValidator{}.AddWhiteCommand(serviceInfo)
	requestervalidator.RequesterValidator{}.StoreRequesterInfo(
		serviceInfo.ServiceName,
		serviceInfo.AllowedRequester,
	)

	t.Run("Error", func(t *testing.T) {
		t.Run("IsNotSetApi", func(t *testing.T) {
			handler.setHelper(mockHelper)
			mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable))

			handler.isSetAPI = false
			handler.APIV1ServicemgrServicesPost(w, r)
		})
		t.Run("IsNotSetKey", func(t *testing.T) {
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)
			mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable))

			handler.IsSetKey = false
			handler.APIV1ServicemgrServicesPost(w, r)
		})
		t.Run("DecryptionFail", func(t *testing.T) {
			handler.SetCipher(mockCipher)
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)
			gomock.InOrder(
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(nil, errors.New("")),
				mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable)),
			)

			handler.APIV1ServicemgrServicesPost(w, r)
		})
		t.Run("EncryptionFail", func(t *testing.T) {
			handler.SetCipher(mockCipher)
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)

			gomock.InOrder(
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(req, nil),
				mockOrchestration.EXPECT().ExecuteAppOnLocal(gomock.Any()),
				mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, errors.New("")),
				mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable)),
			)

			handler.APIV1ServicemgrServicesPost(w, r)
		})
	})

	t.Run("Success", func(t *testing.T) {
		handler.SetCipher(mockCipher)
		handler.SetOrchestrationAPI(mockOrchestration)
		handler.setHelper(mockHelper)

		gomock.InOrder(
			mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(req, nil),
			mockOrchestration.EXPECT().ExecuteAppOnLocal(gomock.Any()),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
			mockHelper.EXPECT().ResponseJSON(gomock.Any(), gomock.Any(), gomock.Eq(http.StatusOK)),
		)

		handler.APIV1ServicemgrServicesPost(w, r)
	})
}

func TestAPIV1ServicemgrServicesNotificationServiceIDPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := GetHandler()
	if handler == nil {
		t.Error("unexpected return value")
	}

	mockOrchestration := orchemock.NewMockOrcheInternalAPI(ctrl)
	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)

	serviceID := float64(1.0)
	status := "Testing"
	notification := make(map[string]interface{})
	notification["ServiceID"] = serviceID
	notification["Status"] = status

	r := httptest.NewRequest("POST", "http://test.test", nil)
	w := httptest.NewRecorder()

	t.Run("Error", func(t *testing.T) {
		t.Run("IsNotSetApi", func(t *testing.T) {
			handler.setHelper(mockHelper)
			mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable))

			handler.isSetAPI = false
			handler.APIV1ServicemgrServicesNotificationServiceIDPost(w, r)
		})
		t.Run("IsNotSetKey", func(t *testing.T) {
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)
			mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable))

			handler.IsSetKey = false
			handler.APIV1ServicemgrServicesNotificationServiceIDPost(w, r)
		})
		t.Run("DecryptionFail", func(t *testing.T) {
			handler.SetCipher(mockCipher)
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)
			gomock.InOrder(
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(nil, errors.New("")),
				mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable)),
			)

			handler.APIV1ServicemgrServicesNotificationServiceIDPost(w, r)
		})
		t.Run("HandleNotificationFail", func(t *testing.T) {
			handler.SetCipher(mockCipher)
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)
			gomock.InOrder(
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(notification, nil),
				mockOrchestration.EXPECT().HandleNotificationOnLocal(gomock.Eq(serviceID), gomock.Eq(status)).Return(errors.New("")),
				mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusInternalServerError)),
			)

			handler.APIV1ServicemgrServicesNotificationServiceIDPost(w, r)
		})
	})

	t.Run("Success", func(t *testing.T) {
		handler.SetCipher(mockCipher)
		handler.SetOrchestrationAPI(mockOrchestration)
		handler.setHelper(mockHelper)
		gomock.InOrder(
			mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(notification, nil),
			mockOrchestration.EXPECT().HandleNotificationOnLocal(gomock.Eq(serviceID), gomock.Eq(status)).Return(nil),
			mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusOK)),
		)

		handler.APIV1ServicemgrServicesNotificationServiceIDPost(w, r)
	})
}

func TestAPIV1ScoringmgrScoreLibnameGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := GetHandler()
	if handler == nil {
		t.Error("unexpected return value")
	}

	mockOrchestration := orchemock.NewMockOrcheInternalAPI(ctrl)
	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)

	serviceID := float64(1.0)
	//	status := "Testing"
	//	notification := make(map[string]interface{})
	//	notification["ServiceID"] = serviceID
	//	notification["Status"] = status

	appNameInfo := make(map[string]interface{})
	appNameInfo["devID"] = "deviceID"

	r := httptest.NewRequest("POST", "http://test.test", nil)
	w := httptest.NewRecorder()

	t.Run("Error", func(t *testing.T) {
		t.Run("IsNotSetApi", func(t *testing.T) {
			handler.setHelper(mockHelper)
			mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable))

			handler.isSetAPI = false
			handler.APIV1ScoringmgrScoreLibnameGet(w, r)
		})
		t.Run("IsNotSetKey", func(t *testing.T) {
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)
			mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable))

			handler.IsSetKey = false
			handler.APIV1ScoringmgrScoreLibnameGet(w, r)
		})
		t.Run("DecryptionFail", func(t *testing.T) {
			handler.SetCipher(mockCipher)
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)
			gomock.InOrder(
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(nil, errors.New("")),
				mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable)),
			)

			handler.APIV1ScoringmgrScoreLibnameGet(w, r)
		})
		t.Run("GetScoreFail", func(t *testing.T) {
			handler.SetCipher(mockCipher)
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)
			gomock.InOrder(
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(appNameInfo, nil),
				mockOrchestration.EXPECT().GetScore(gomock.Any()).Return(serviceID, errors.New("")),
				mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusInternalServerError)),
			)

			handler.APIV1ScoringmgrScoreLibnameGet(w, r)
		})
		t.Run("EncryptionFail", func(t *testing.T) {
			handler.SetCipher(mockCipher)
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)
			gomock.InOrder(
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(appNameInfo, nil),
				mockOrchestration.EXPECT().GetScore(gomock.Any()).Return(serviceID, nil),
				mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, errors.New("")),
				mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable)),
			)

			handler.APIV1ScoringmgrScoreLibnameGet(w, r)
		})
	})

	t.Run("Success", func(t *testing.T) {
		handler.SetCipher(mockCipher)
		handler.SetOrchestrationAPI(mockOrchestration)
		handler.setHelper(mockHelper)
		gomock.InOrder(
			mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(appNameInfo, nil),
			mockOrchestration.EXPECT().GetScore(gomock.Any()).Return(serviceID, nil),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
			mockHelper.EXPECT().ResponseJSON(gomock.Any(), gomock.Any(), gomock.Eq(http.StatusOK)),
		)

		handler.APIV1ScoringmgrScoreLibnameGet(w, r)
	})
}

func TestAPIV1DiscoveryFromVPNServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := GetHandler()
	if handler == nil {
		t.Error("unexpected return value")
	}

	mockOrchestration := orchemock.NewMockOrcheInternalAPI(ctrl)
	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)

	deviceDetailsInfo := make(map[string]interface{})
	deviceDetailsInfo["DeviceID"] = "deviceID"
	deviceDetailsInfo["PrivateAddr"] = "privateIP"
	deviceDetailsInfo["VirtualAddr"] = "virtualIP"

	r := httptest.NewRequest("POST", "http://test.test", nil)
	w := httptest.NewRecorder()

	t.Run("IsNotSetApi", func(t *testing.T) {
		handler.setHelper(mockHelper)
		mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable))

		handler.isSetAPI = false
		handler.APIV1DiscoverymgrMNEDCDeviceInfoPost(w, r)
	})
	t.Run("IsNotSetKey", func(t *testing.T) {
		handler.SetOrchestrationAPI(mockOrchestration)
		handler.setHelper(mockHelper)
		mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable))

		handler.IsSetKey = false
		handler.APIV1DiscoverymgrMNEDCDeviceInfoPost(w, r)
	})
	t.Run("DecryptionFail", func(t *testing.T) {
		handler.SetCipher(mockCipher)
		handler.SetOrchestrationAPI(mockOrchestration)
		handler.setHelper(mockHelper)
		gomock.InOrder(
			mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(nil, errors.New("")),
			mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable)),
		)

		handler.APIV1DiscoverymgrMNEDCDeviceInfoPost(w, r)
	})
	t.Run("Success", func(t *testing.T) {
		handler.SetCipher(mockCipher)
		handler.SetOrchestrationAPI(mockOrchestration)
		handler.setHelper(mockHelper)
		gomock.InOrder(
			mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(deviceDetailsInfo, nil),
			mockOrchestration.EXPECT().HandleDeviceInfo(gomock.Any(), gomock.Any(), gomock.Any()),
			mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusOK)),
		)

		handler.APIV1DiscoverymgrMNEDCDeviceInfoPost(w, r)
	})
}

func TestAPIV1DiscoverymgrOrchInfoGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := GetHandler()
	if handler == nil {
		t.Error("unexpected return value")
	}

	mockOrchestration := orchemock.NewMockOrcheInternalAPI(ctrl)
	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)

	r := httptest.NewRequest("POST", "http://test.test", nil)
	w := httptest.NewRecorder()

	t.Run("IsNotSetApi", func(t *testing.T) {
		handler.setHelper(mockHelper)
		mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable))

		handler.isSetAPI = false
		handler.APIV1DiscoverymgrOrchestrationInfoGet(w, r)
	})
	t.Run("IsNotSetKey", func(t *testing.T) {
		handler.SetOrchestrationAPI(mockOrchestration)
		handler.setHelper(mockHelper)
		mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable))

		handler.IsSetKey = false
		handler.APIV1DiscoverymgrOrchestrationInfoGet(w, r)
	})
	t.Run("GetOrchInfoError", func(t *testing.T) {
		handler.SetCipher(mockCipher)
		handler.SetOrchestrationAPI(mockOrchestration)
		handler.setHelper(mockHelper)
		gomock.InOrder(
			mockOrchestration.EXPECT().GetOrchestrationInfo().Return("", "", []string{""}, errors.New("")),
			mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable)),
		)

		handler.APIV1DiscoverymgrOrchestrationInfoGet(w, r)
	})
	t.Run("EncryptFail", func(t *testing.T) {
		handler.SetCipher(mockCipher)
		handler.SetOrchestrationAPI(mockOrchestration)
		handler.setHelper(mockHelper)
		gomock.InOrder(
			mockOrchestration.EXPECT().GetOrchestrationInfo().Return("", "", []string{""}, nil),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, errors.New("")),
			mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable)),
		)

		handler.APIV1DiscoverymgrOrchestrationInfoGet(w, r)
	})
	t.Run("Success", func(t *testing.T) {
		handler.SetCipher(mockCipher)
		handler.SetOrchestrationAPI(mockOrchestration)
		handler.setHelper(mockHelper)
		gomock.InOrder(
			mockOrchestration.EXPECT().GetOrchestrationInfo().Return("", "", []string{""}, nil),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
			mockHelper.EXPECT().ResponseJSON(gomock.Any(), gomock.Any(), gomock.Eq(http.StatusOK)),
		)

		handler.APIV1DiscoverymgrOrchestrationInfoGet(w, r)
	})

}

func TestSetCertificateFilePath(t *testing.T) {
	testHandler := new(Handler)
	testHandler.SetCertificateFilePath("test")

	testURL := testHandler.helper.MakeTargetURL("", 1, "")

	if strings.Contains(testURL, "https") != true {
		t.Error("expected key is set, but not set on helper: ", testURL)
	}
}
