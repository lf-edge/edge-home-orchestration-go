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
package storagemgr

import (
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/storagemgr/storagedriver"
	"errors"
	"github.com/edgexfoundry/device-sdk-go"
	"github.com/edgexfoundry/device-sdk-go/pkg/startup"
	"os"
)

const (
	dataStorageService = "datastorage"
	dataStorageFilePath = "/var/edge-orchestration" + "/datastorage/configuration.toml"
)

type Storage interface{
	StartStorage() error
}

type StorageImpl struct {}

var (
	storageIns *StorageImpl
)

func init(){
	storageIns = &StorageImpl{}
}
func GetInstance() Storage {
	return storageIns
}
func (StorageImpl) StartStorage() error {
	if _, err := os.Stat(dataStorageFilePath); err == nil {
		sd := storagedriver.StorageDriver{}
		go startup.Bootstrap(dataStorageService, device.Version, &sd)
		return nil
	}
	return errors.New("could not initiate storageManager")
}
