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
	"errors"
	"testing"

	clientMocks "github.com/lf-edge/edge-home-orchestration-go/internal/controller/discoverymgr/mnedc/client/mocks"

	"github.com/golang/mock/gomock"
)

var (
	defaultConfigPath       = "server.config"
	defaultDeviceID         = "dummyID"
	defaultDeviceIDFilePath = "testdata/deviceID.txt"

	mockMnedcClient *clientMocks.MockMNEDCClient
)

func init() {
}

func TestStartMNEDCClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("CreateClientError", func(t *testing.T) {
		c := GetClientInstance()
		mockMnedcClient.EXPECT().SetClient(gomock.Any())
		mockMnedcClient.EXPECT().CreateClient(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New(""))
		c.StartMNEDCClient(defaultDeviceIDFilePath, defaultConfigPath)
	})
	t.Run("Success", func(t *testing.T) {
		c := GetClientInstance()
		mockMnedcClient.EXPECT().SetClient(gomock.Any())
		mockMnedcClient.EXPECT().CreateClient(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
		mockMnedcClient.EXPECT().Run()
		mockMnedcClient.EXPECT().NotifyBroadcastServer(gomock.Any()).Return(nil).AnyTimes()

		c.StartMNEDCClient(defaultDeviceIDFilePath, defaultConfigPath)
	})
}

func createMockIns(ctrl *gomock.Controller) {
	mockMnedcClient = clientMocks.NewMockMNEDCClient(ctrl)
	mnedcClientIns = mockMnedcClient
}
