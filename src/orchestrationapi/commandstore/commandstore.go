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
package commandstore

import (
	"errors"
	"sync"

	"common/types/configuremgrtypes"
)

type CommandList interface {
	GetServiceFileName(serviceName string) (string, error)
	StoreServiceInfo(serviceInfo configuremgrtypes.ServiceInfo)
}

type commandListImpl struct {
	serviceInfos map[string]string
	mutex        *sync.Mutex
}

var commandList commandListImpl

func GetInstance() *CommandList {
	return &commandList
}

func (c commandListImpl) GetServiceFileName(serviceName string) (string, error) {
	val, ok := c.serviceInfos[serviceName]
	if !ok {
		return "", errors.New("not found registered service")
	}

	return val, nil
}

func (c *commandListImpl) StoreServiceInfo(serviceInfo configuremgrtypes.ServiceInfo) {
	mutex.Lock()
	c.serviceInfos[serviceInfo.ServiceName] = serviceInfo.ExecutableFileName
	mutex.Unlock()
}
