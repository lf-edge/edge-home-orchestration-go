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
		orcheExternalAPIMock := orchemock.NewMockOrcheExternalApi(ctrl)

		handler.SetOrchestrationAPI(orcheExternalAPIMock)
		if handler.api != orcheExternalAPIMock {
			t.Error("unexpected difference of copied value")
		}
		if handler.isSetAPI == false {
			t.Error("unexpected flag value")
		}
	})
}

func TestAPIV1RequestServicePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := GetHandler()
	if handler == nil {
		t.Error("unexpected return value")
	}

	mockOrchestration := orchemock.NewMockOrcheExternalApi(ctrl)
	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)

	command := "test"
	args := []string{"-a", "-b"}

	appCommand := make(map[string]interface{})
	appCommand["Name"] = command

	iargs := make([]interface{}, len(args))
	for idx, arg := range args {
		iargs[idx] = arg
	}
	appCommand["Args"] = iargs

	resp := []byte{'1'}

	r := httptest.NewRequest("POST", "http://test.test", nil)
	w := httptest.NewRecorder()

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
			gomock.InOrder(
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(nil, errors.New("")),
				mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable)),
			)

			handler.APIV1RequestServicePost(w, r)
		})
		t.Run("EncryptionFail", func(t *testing.T) {
			handler.SetCipher(mockCipher)
			handler.SetOrchestrationAPI(mockOrchestration)
			handler.setHelper(mockHelper)
			gomock.InOrder(
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(appCommand, nil),
				mockOrchestration.EXPECT().RequestService(gomock.Eq(command), gomock.Eq(args)),
				mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(resp, errors.New("")),
				mockHelper.EXPECT().Response(gomock.Any(), gomock.Eq(http.StatusServiceUnavailable)),
			)

			handler.APIV1RequestServicePost(w, r)
		})
	})

	t.Run("Success", func(t *testing.T) {
		handler.SetCipher(mockCipher)
		handler.SetOrchestrationAPI(mockOrchestration)
		handler.setHelper(mockHelper)

		gomock.InOrder(
			mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(appCommand, nil),
			mockOrchestration.EXPECT().RequestService(gomock.Eq(command), gomock.Eq(args)),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(resp, nil),
			mockHelper.EXPECT().ResponseJSON(gomock.Any(), gomock.Eq(resp), gomock.Eq(http.StatusOK)),
		)

		handler.APIV1RequestServicePost(w, r)
	})
}
