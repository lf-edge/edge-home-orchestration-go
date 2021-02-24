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

package restclient

import (
	"errors"
	"net/http"
	"testing"

	ciphermock "github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/cipher/mocks"
	helpermock "github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/resthelper/mocks"

	"github.com/golang/mock/gomock"
)

func TestGetRestClient(t *testing.T) {
	client := GetRestClient()
	if client == nil {
		t.Error("unexpected return value")
	}
}

func TestDoExecuteRemoteDevice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := restClient
	if client == nil {
		t.Error("unexpected return value")
	}

	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)

	decryptJSON := make(map[string]interface{})
	decryptJSON["Status"] = "Failed"

	t.Run("Error", func(t *testing.T) {
		t.Run("IsNotSetKey", func(t *testing.T) {
			client.setHelper(mockHelper)

			client.IsSetKey = false
			err := client.DoExecuteRemoteDevice(make(map[string]interface{}), "")
			if err == nil {
				t.Error("expect error is not nil, but nil")
			}
		})
		t.Run("EncryptionFail", func(t *testing.T) {
			client.SetCipher(mockCipher)
			client.setHelper(mockHelper)
			gomock.InOrder(
				mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
				mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, errors.New("")),
			)

			err := client.DoExecuteRemoteDevice(make(map[string]interface{}), "")
			if err == nil {
				t.Error("expect error is not nil, but nil")
			}
		})
		t.Run("DoPost", func(t *testing.T) {
			t.Run("ReturnError", func(t *testing.T) {
				client.SetCipher(mockCipher)
				client.setHelper(mockHelper)
				gomock.InOrder(
					mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
					mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
					mockHelper.EXPECT().DoPost(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, errors.New("")),
				)

				err := client.DoExecuteRemoteDevice(make(map[string]interface{}), "")
				if err == nil {
					t.Error("expect error is not nil, but nil")
				}
			})
			t.Run("StatusNotOk", func(t *testing.T) {
				client.SetCipher(mockCipher)
				client.setHelper(mockHelper)
				gomock.InOrder(
					mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
					mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
					mockHelper.EXPECT().DoPost(gomock.Any(), gomock.Any()).Return(nil, http.StatusInternalServerError, nil),
				)

				err := client.DoExecuteRemoteDevice(make(map[string]interface{}), "")
				if err == nil {
					t.Error("expect error is not nil, but nil")
				}
			})
		})
		t.Run("DecryptionFail", func(t *testing.T) {
			client.SetCipher(mockCipher)
			client.setHelper(mockHelper)
			gomock.InOrder(
				mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
				mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
				mockHelper.EXPECT().DoPost(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, nil),
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(nil, errors.New("")),
			)

			err := client.DoExecuteRemoteDevice(make(map[string]interface{}), "")
			if err == nil {
				t.Error("expect error is not nil, but nil")
			}
		})
		t.Run("RemoteExecuteFail", func(t *testing.T) {
			client.SetCipher(mockCipher)
			client.setHelper(mockHelper)
			gomock.InOrder(
				mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
				mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
				mockHelper.EXPECT().DoPost(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, nil),
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(decryptJSON, nil),
			)

			err := client.DoExecuteRemoteDevice(make(map[string]interface{}), "")
			if err == nil {
				t.Error("expect error is not nil, but nil")
			}
		})
	})

	t.Run("Success", func(t *testing.T) {
		client.SetCipher(mockCipher)
		client.setHelper(mockHelper)
		decryptJSON["Status"] = "Test"
		gomock.InOrder(
			mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
			mockHelper.EXPECT().DoPost(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, nil),
			mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(decryptJSON, nil),
		)

		err := client.DoExecuteRemoteDevice(make(map[string]interface{}), "")
		if err != nil {
			t.Error("expect error is nil, but not nil")
		}
	})
}

func TestDoNotifyAppStatusRemoteDevice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := restClient
	if client == nil {
		t.Error("unexpected return value")
	}

	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)

	decryptJSON := make(map[string]interface{})
	decryptJSON["Status"] = "Failed"

	t.Run("Error", func(t *testing.T) {
		t.Run("IsNotSetKey", func(t *testing.T) {
			client.setHelper(mockHelper)

			client.IsSetKey = false
			err := client.DoNotifyAppStatusRemoteDevice(make(map[string]interface{}), 1, "")
			if err == nil {
				t.Error("expect error is not nil, but nil")
			}
		})
		t.Run("EncryptionFail", func(t *testing.T) {
			client.SetCipher(mockCipher)
			client.setHelper(mockHelper)
			gomock.InOrder(
				mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
				mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, errors.New("")),
			)

			err := client.DoNotifyAppStatusRemoteDevice(make(map[string]interface{}), 1, "")
			if err == nil {
				t.Error("expect error is not nil, but nil")
			}
		})
		t.Run("DoPost", func(t *testing.T) {
			t.Run("ReturnError", func(t *testing.T) {
				client.SetCipher(mockCipher)
				client.setHelper(mockHelper)
				gomock.InOrder(
					mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
					mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
					mockHelper.EXPECT().DoPost(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, errors.New("")),
				)

				err := client.DoNotifyAppStatusRemoteDevice(make(map[string]interface{}), 1, "")
				if err == nil {
					t.Error("expect error is not nil, but nil")
				}
			})
			t.Run("StatusNotOk", func(t *testing.T) {
				client.SetCipher(mockCipher)
				client.setHelper(mockHelper)
				gomock.InOrder(
					mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
					mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
					mockHelper.EXPECT().DoPost(gomock.Any(), gomock.Any()).Return(nil, http.StatusInternalServerError, nil),
				)

				err := client.DoNotifyAppStatusRemoteDevice(make(map[string]interface{}), 1, "")
				if err == nil {
					t.Error("expect error is not nil, but nil")
				}
			})
		})
	})

	t.Run("Success", func(t *testing.T) {
		client.SetCipher(mockCipher)
		client.setHelper(mockHelper)
		gomock.InOrder(
			mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
			mockHelper.EXPECT().DoPost(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, nil),
		)

		err := client.DoNotifyAppStatusRemoteDevice(make(map[string]interface{}), 1, "")
		if err != nil {
			t.Error("expect error is nil, but not nil")
		}
	})
}

func TestDoGetScoreRemoteDevice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := restClient
	if client == nil {
		t.Error("unexpected return value")
	}

	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)

	t.Run("Error", func(t *testing.T) {
		t.Run("IsNotSetKey", func(t *testing.T) {
			client.setHelper(mockHelper)

			client.IsSetKey = false
			_, err := client.DoGetScoreRemoteDevice("", "")
			if err == nil {
				t.Error("expect error is not nil, but nil")
			}
		})
		t.Run("DoGetWithBody", func(t *testing.T) {
			t.Run("ReturnError", func(t *testing.T) {
				client.SetCipher(mockCipher)
				client.setHelper(mockHelper)
				gomock.InOrder(
					mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
					mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
					mockHelper.EXPECT().DoGetWithBody(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, errors.New("")),
				)

				_, err := client.DoGetScoreRemoteDevice("", "")
				if err == nil {
					t.Error("expect error is not nil, but nil")
				}
			})
			t.Run("StatusNotOk", func(t *testing.T) {
				client.SetCipher(mockCipher)
				client.setHelper(mockHelper)
				gomock.InOrder(
					mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
					mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
					mockHelper.EXPECT().DoGetWithBody(gomock.Any(), gomock.Any()).Return(nil, http.StatusInternalServerError, nil),
				)

				_, err := client.DoGetScoreRemoteDevice("", "")
				if err == nil {
					t.Error("expect error is not nil, but nil")
				}
			})
		})
		t.Run("DecryptionFail", func(t *testing.T) {
			client.SetCipher(mockCipher)
			client.setHelper(mockHelper)
			gomock.InOrder(
				mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
				mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
				mockHelper.EXPECT().DoGetWithBody(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, nil),
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(nil, errors.New("")),
			)

			_, err := client.DoGetScoreRemoteDevice("", "")
			if err == nil {
				t.Error("expect error is not nil, but nil")
			}
		})
		t.Run("ReturnZero", func(t *testing.T) {
			client.SetCipher(mockCipher)
			client.setHelper(mockHelper)

			respMsg := make(map[string]interface{})
			respMsg["ScoreValue"] = float64(0.0)

			gomock.InOrder(
				mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
				mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
				mockHelper.EXPECT().DoGetWithBody(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, nil),
				mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(respMsg, nil),
			)

			_, err := client.DoGetScoreRemoteDevice("", "")
			if err == nil {
				t.Error("expect error is not nil, but nil")
			}
		})
	})

	t.Run("Success", func(t *testing.T) {
		client.SetCipher(mockCipher)
		client.setHelper(mockHelper)

		respMsg := make(map[string]interface{})
		respMsg["ScoreValue"] = float64(1.0)

		gomock.InOrder(
			mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
			mockHelper.EXPECT().DoGetWithBody(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, nil),
			mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(respMsg, nil),
		)

		score, err := client.DoGetScoreRemoteDevice("", "")
		if err != nil {
			t.Error("expect error is nil, but not nil")
		} else if score != float64(1.0) {
			t.Error("unexpected score value")
		}
	})
}

func TestDoGetOrchestrationInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := restClient
	if client == nil {
		t.Error("unexpected return value")
	}

	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)

	t.Run("IsNotSetKey", func(t *testing.T) {
		client.setHelper(mockHelper)

		client.IsSetKey = false
		_, _, _, err := client.DoGetOrchestrationInfo("")
		if err == nil {
			t.Error("expect error is not nil, but nil")
		}
	})
	t.Run("EncryptionFail", func(t *testing.T) {
		client.SetCipher(mockCipher)
		client.setHelper(mockHelper)

		gomock.InOrder(
			mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, errors.New("")),
		)
		_, _, _, err := client.DoGetOrchestrationInfo("")
		if err == nil {
			t.Error("expect error is not nil, but nil")
		}
	})

	t.Run("DoGetWithBody", func(t *testing.T) {
		client.SetCipher(mockCipher)
		client.setHelper(mockHelper)

		gomock.InOrder(
			mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
			mockHelper.EXPECT().DoGetWithBody(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, errors.New("")),
		)
		_, _, _, err := client.DoGetOrchestrationInfo("")
		if err == nil {
			t.Error("expect error is not nil, but nil")
		}
	})
	t.Run("DecryptionFail", func(t *testing.T) {
		client.SetCipher(mockCipher)
		client.setHelper(mockHelper)

		gomock.InOrder(
			mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
			mockHelper.EXPECT().DoGetWithBody(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, nil),
			mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(nil, errors.New("")),
		)
		_, _, _, err := client.DoGetOrchestrationInfo("")
		if err == nil {
			t.Error("expected error is not nil, but nil")
		}
	})
	t.Run("Success", func(t *testing.T) {
		client.SetCipher(mockCipher)
		client.setHelper(mockHelper)

		respJSONMsg := make(map[string]interface{})
		respJSONMsg["Platform"] = "platform"
		respJSONMsg["ExecutionType"] = "execution"
		respJSONMsg["ServiceList"] = []string{"service1", "service2"}

		gomock.InOrder(
			mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
			mockHelper.EXPECT().DoGetWithBody(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, nil),
			mockCipher.EXPECT().DecryptByteToJSON(gomock.Any()).Return(respJSONMsg, nil),
		)
		_, _, _, err := client.DoGetOrchestrationInfo("")
		if err != nil {
			t.Error("expected error is nil, but not nil")
		}
	})
}

func TestDoNotifyMNEDCBroadcastServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := restClient
	if client == nil {
		t.Error("unexpected return value")
	}

	defaultServerIP := "1.1.1.1"
	defaultPort := 3333
	defaultDeviceID := "dummyID"
	defaultPrivateIP := "2.2.2.2"
	defaultVirtualIP := "10.10.10.10"

	mockCipher := ciphermock.NewMockIEdgeCipherer(ctrl)
	mockHelper := helpermock.NewMockRestHelper(ctrl)

	t.Run("IsNotSetKey", func(t *testing.T) {
		client.setHelper(mockHelper)

		client.IsSetKey = false
		err := client.DoNotifyMNEDCBroadcastServer(defaultServerIP, defaultPort, defaultDeviceID, defaultPrivateIP, defaultVirtualIP)
		if err == nil {
			t.Error("expect error is not nil, but nil")
		}
	})
	t.Run("EncryptionFail", func(t *testing.T) {
		client.SetCipher(mockCipher)
		client.setHelper(mockHelper)

		mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, errors.New(""))

		err := client.DoNotifyMNEDCBroadcastServer(defaultServerIP, defaultPort, defaultDeviceID, defaultPrivateIP, defaultVirtualIP)
		if err == nil {
			t.Error("expect error is not nil, but nil")
		}
	})
	t.Run("DoPostError", func(t *testing.T) {
		client.SetCipher(mockCipher)
		client.setHelper(mockHelper)

		gomock.InOrder(
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
			mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
			mockHelper.EXPECT().DoPost(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, errors.New("")),
		)
		err := client.DoNotifyMNEDCBroadcastServer(defaultServerIP, defaultPort, defaultDeviceID, defaultPrivateIP, defaultVirtualIP)
		if err == nil {
			t.Error("expect error is not nil, but nil")
		}
	})

	t.Run("DoPostError", func(t *testing.T) {
		client.SetCipher(mockCipher)
		client.setHelper(mockHelper)

		gomock.InOrder(
			mockCipher.EXPECT().EncryptJSONToByte(gomock.Any()).Return(nil, nil),
			mockHelper.EXPECT().MakeTargetURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(""),
			mockHelper.EXPECT().DoPost(gomock.Any(), gomock.Any()).Return(nil, http.StatusOK, nil),
		)
		err := client.DoNotifyMNEDCBroadcastServer(defaultServerIP, defaultPort, defaultDeviceID, defaultPrivateIP, defaultVirtualIP)
		if err != nil {
			t.Error("expect error is nil, but not nil")
		}
	})
}
