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

package internalhandler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	orchemock "orchestrationapi/mocks"
	ciphermock "restinterface/cipher/mocks"
	helpermock "restinterface/resthelper/mocks"

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
		orcheExternalAPIMock := orchemock.NewMockOrcheInternalApi(ctrl)

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

	mockOrchestration := orchemock.NewMockOrcheInternalApi(ctrl)
	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)

	r := httptest.NewRequest("POST", "http://test.test", nil)
	w := httptest.NewRecorder()

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
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(make(map[string]interface{}), nil),
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
			mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(make(map[string]interface{}), nil),
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

	mockOrchestration := orchemock.NewMockOrcheInternalApi(ctrl)
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

	mockOrchestration := orchemock.NewMockOrcheInternalApi(ctrl)
	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)

	serviceID := float64(1.0)
	//	status := "Testing"
	//	notification := make(map[string]interface{})
	//	notification["ServiceID"] = serviceID
	//	notification["Status"] = status

	appNameInfo := make(map[string]interface{})
	appNameInfo["appName"] = "test"

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
				mockOrchestration.EXPECT().GetScore(gomock.Any(), gomock.Any()).Return(serviceID, errors.New("")),
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
				mockOrchestration.EXPECT().GetScore(gomock.Any(), gomock.Any()).Return(serviceID, nil),
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
			mockOrchestration.EXPECT().GetScore(gomock.Any(), gomock.Any()).Return(serviceID, nil),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
			mockHelper.EXPECT().ResponseJSON(gomock.Any(), gomock.Any(), gomock.Eq(http.StatusOK)),
		)

		handler.APIV1ScoringmgrScoreLibnameGet(w, r)
	})
}
