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
package commandvalidator

import (
	"errors"
	"sync"

	"common/types/configuremgrtypes"
	"db/bolt/common"
)

type CommandList interface {
	GetServiceFileName(serviceName string) (string, error)
	StoreServiceInfo(serviceInfo configuremgrtypes.ServiceInfo) error
}

type commandListImpl struct {
	serviceInfos map[string]string
	mutex        *sync.Mutex
	blackList    []string
}

const (
	NOT_FOUND_REGISTERED_SERVICE   = "not found registered service"
	NOT_ALLOWED_EXECUTABLE_SERVICE = "not allowed executable service"
)

var blackList = []string{
	"sudo",
	"su",
	"cat",
}

var commandList commandListImpl

func GetInstance() CommandList {
	return &commandList
}

func init() {
	commandList.mutex = &sync.Mutex{}
	commandList.serviceInfos = make(map[string]string)
}

func (c commandListImpl) GetServiceFileName(serviceName string) (string, error) {
	val, ok := c.serviceInfos[serviceName]
	if !ok {
		return "", errors.New(NOT_FOUND_REGISTERED_SERVICE)
	}

	return val, nil
}

func (c *commandListImpl) StoreServiceInfo(serviceInfo configuremgrtypes.ServiceInfo) error {
	if common.HasElem(blackList, serviceInfo.ExecutableFileName) {
		return errors.New(NOT_ALLOWED_EXECUTABLE_SERVICE)
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.serviceInfos[serviceInfo.ServiceName] = serviceInfo.ExecutableFileName

	return nil
}
