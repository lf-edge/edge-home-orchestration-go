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

package commands

import (
	"errors"
	"sync"
)

type CommandStore interface {
	GetServiceFileName(serviceName string) (string, error)
	StoreServiceInfo(serviceName, command string)
}

type commands struct {
	serviceInfos map[string]string
	mutex        *sync.Mutex
}

const notFoundRegisteredService = "not found registered service"

var commandList commands

func GetInstance() CommandStore {
	return &commandList
}

func init() {
	commandList.mutex = &sync.Mutex{}
	commandList.serviceInfos = make(map[string]string)
}

func (c *commands) GetServiceFileName(serviceName string) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	val, ok := c.serviceInfos[serviceName]
	if !ok {
		return "", errors.New(notFoundRegisteredService)
	}

	return val, nil
}

func (c *commands) StoreServiceInfo(serviceName, command string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.serviceInfos[serviceName] = command
}
