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
	"log"
	"os"
	"syscall"
	"testing"
	"time"

	discoveryMocks "controller/discoverymgr/mocks"
	"controller/mnedcmgr/client"
	clientMocks "controller/mnedcmgr/client/mocks"

	"github.com/golang/mock/gomock"
)

var (
	incorrectDeviceIDFilePath = "/etc/edge-orchestration/deviceID.txt"
	defaultConfigPath         = "server.config"
	defaultDeviceIDFilePath   = "deviceId.txt"
	defaultDeviceID           = "dummyID"

	mockMnedcClient *clientMocks.MockMNEDCClient
	mockDiscovery   *discoveryMocks.MockDiscovery
)

func init() {
}

func TestStartMNEDCClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("ReadFileError", func(t *testing.T) {
		c := GetClientInstance()
		c.StartMNEDCClient(incorrectDeviceIDFilePath, defaultConfigPath)
	})
	t.Run("CreateClientError", func(t *testing.T) {
		c := GetClientInstance()

		err := createDeviceIDFile()
		if err != nil {
			log.Println("Could not write to the device id file", err.Error())
			return
		}
		mockMnedcClient.EXPECT().CreateClient(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New(""))
		c.StartMNEDCClient(defaultDeviceIDFilePath, defaultConfigPath)
	})
	t.Run("TestInterrupt", func(t *testing.T) {
		fatalErrChan := make(chan error)
		mockMnedcClient.EXPECT().Close().Return(nil).AnyTimes()
		go waitInterrupt(fatalErrChan)
		time.Sleep(2 * time.Second)

		syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	})
	t.Run("Success", func(t *testing.T) {
		c := GetClientInstance()

		mockMnedcClient.EXPECT().CreateClient(gomock.Any(), gomock.Any(), gomock.Any()).Return(&client.Client{}, nil)
		mockMnedcClient.EXPECT().Run()
		mockDiscovery.EXPECT().NotifyMNEDCBroadcastServer().Return(errors.New("")).AnyTimes()

		c.StartMNEDCClient(defaultDeviceIDFilePath, defaultConfigPath)

	})

	deleteDeviceIDFile()
}

func deleteDeviceIDFile() {
	err := os.Remove(defaultDeviceIDFilePath)
	if err != nil {
		log.Println("Could not delete file")
	}

}

func createDeviceIDFile() error {
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

func createMockIns(ctrl *gomock.Controller) {
	mockMnedcClient = clientMocks.NewMockMNEDCClient(ctrl)
	mnedcClientIns = mockMnedcClient
	mockDiscovery = discoveryMocks.NewMockDiscovery(ctrl)
	discoveryIns = mockDiscovery
}
