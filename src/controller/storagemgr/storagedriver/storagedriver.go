//
// Copyright (c) 2020 Samsung Electronics Co., Ltd All Rights Reserved.
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package storagedriver

import (
	"fmt"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"

	sdk "github.com/edgexfoundry/device-sdk-go"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

const logPrefix = "[storagedriver]"

type StorageDriver struct {
	logger      logger.LoggingClient
	asyncValues chan<- *dsModels.AsyncValues
}

var (
	log = logmgr.GetInstance()
)

// Initialize performs protocol-specific initialization for the device
func (driver *StorageDriver) Initialize(logger logger.LoggingClient, asyncValues chan<- *dsModels.AsyncValues) error {
	log.Println(logPrefix, "Device service intialize started")
	driver.logger = logger
	handler := NewStorageHandler(sdk.RunningService(), logger, asyncValues)
	return handler.Start()
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (driver *StorageDriver) HandleReadCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest) ([]*dsModels.CommandValue, error) {
	return nil, fmt.Errorf("StorageDriver.HandleReadCommands; read commands not supported")
}

// HandleWriteCommands passes a slice of CommandRequest
func (driver *StorageDriver) HandleWriteCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest,
	params []*dsModels.CommandValue) error {

	return fmt.Errorf("StorageDriver.HandleWriteCommands; write commands not supported")
}

// Stop the protocol-specific DS code to shutdown gracefully
func (driver *StorageDriver) Stop(force bool) error {
	log.Println(logPrefix, "StorageDriver.Stop called:", force)

	return nil
}

// when a new Device associated with this Device Service is added
func (driver *StorageDriver) AddDevice(deviceName string, protocols map[string]contract.ProtocolProperties, adminState contract.AdminState) error {

	log.Println(logPrefix, "Device has been successfully added!!!!!!", deviceName)
	//TODO add the logic for dynamic device addition after discussion
	return nil
}

// when a Device associated with this Device Service is updated
func (driver *StorageDriver) UpdateDevice(deviceName string, protocols map[string]contract.ProtocolProperties, adminState contract.AdminState) error {

	//TODO add the logic for dynamic device updation after discussion
	return nil
}

// when a Device associated with this Device Service is removed
func (driver *StorageDriver) RemoveDevice(deviceName string, protocols map[string]contract.ProtocolProperties) error {

	return nil
}
