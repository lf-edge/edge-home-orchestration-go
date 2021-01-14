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

	"github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/client"
	clientMocks "github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/client/mocks"
	discoveryMocks "github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mocks"

	"github.com/golang/mock/gomock"
)

var (
	defaultConfigPath       = "server.config"
	defaultDeviceID         = "dummyID"
	defaultDeviceIDFilePath = "deviceID.txt"

	mockMnedcClient *clientMocks.MockMNEDCClient
	mockDiscovery   *discoveryMocks.MockDiscovery
)

func init() {
}

func TestStartMNEDCClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("CreateClientError", func(t *testing.T) {
		c := GetClientInstance()

		mockDiscovery.EXPECT().GetDeviceID().Return(defaultDeviceID, nil)
		mockMnedcClient.EXPECT().CreateClient(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New(""))
		c.StartMNEDCClient(defaultDeviceIDFilePath, defaultConfigPath)
	})
	t.Run("Success", func(t *testing.T) {
		c := GetClientInstance()

		mockDiscovery.EXPECT().GetDeviceID().Return(defaultDeviceID, nil)
		mockMnedcClient.EXPECT().CreateClient(gomock.Any(), gomock.Any(), gomock.Any()).Return(&client.Client{}, nil)
		mockMnedcClient.EXPECT().Run()
		mockDiscovery.EXPECT().NotifyMNEDCBroadcastServer().Return(errors.New("")).AnyTimes()

		c.StartMNEDCClient(defaultDeviceIDFilePath, defaultConfigPath)

	})

}

func createMockIns(ctrl *gomock.Controller) {
	mockMnedcClient = clientMocks.NewMockMNEDCClient(ctrl)
	mnedcClientIns = mockMnedcClient
	mockDiscovery = discoveryMocks.NewMockDiscovery(ctrl)
	discoveryIns = mockDiscovery
}
