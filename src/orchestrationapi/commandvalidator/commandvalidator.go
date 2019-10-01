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
	"strings"
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
	NOT_FOUND_EXECUTABLE_FILE      = "not found executable file"
)

var blackList = []string{
	"sudo",
	"su",
	"bash",
	"bsh",
	"csh",
	"adb",
	"sh",
	"ssh",
	"scp",
	"cat",
	"chage",
	"chpasswd",
	"dmidecode",
	"dmsetup",
	"fcinfo",
	"fdisk",
	"iscsiadm",
	"lsof",
	"multipath",
	"oratab",
	"prtvtoc",
	"ps",
	"pburn",
	"pfexec",
	"dzdo",
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
	var command string
	commands := strings.Split(serviceInfo.ExecutableFileName, "/")
	switch len(commands) {
	case 0:
		return errors.New(NOT_FOUND_EXECUTABLE_FILE)
	case 1:
		command = commands[0]
	default:
		command = commands[len(commands)-1]
	}
	if common.HasElem(blackList, command) {
		return errors.New(NOT_ALLOWED_EXECUTABLE_SERVICE)
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.serviceInfos[serviceInfo.ServiceName] = serviceInfo.ExecutableFileName

	return nil
}
