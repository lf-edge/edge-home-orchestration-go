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

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/types/configuremgrtypes"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/commandvalidator/blacklist"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/commandvalidator/commands"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/commandvalidator/injectionchecker"
)

const (
	notMatchedServiceWithExecutable = "not matched service with executable"
	notAllowedExecutableService     = "not allowed executable service"
	notFoundExecutableFile          = "not found executable file"
	foundInjectionCommand           = "found injection command"
	alreadyRegisteredServiceName    = "already registered service name"
)

type ICommandValidator interface {
	AddWhiteCommand(configuremgrtypes.ServiceInfo) error
	GetCommand(serviceName string) (string, error)
	CheckCommand(command []string) error
}

type CommandValidator struct{}

func (CommandValidator) GetCommand(serviceName string) (string, error) {
	return commands.GetInstance().GetServiceFileName(serviceName)
}

func (CommandValidator) AddWhiteCommand(serviceInfo configuremgrtypes.ServiceInfo) error {
	command, err := getExecutableName(serviceInfo.ExecutableFileName)
	if err != nil {
		return err
	}

	if blacklist.IsBlack(command) {
		return errors.New(notAllowedExecutableService)
	}

	_, err = commands.GetInstance().GetServiceFileName(serviceInfo.ServiceName)
	if err == nil {
		return errors.New(alreadyRegisteredServiceName)
	}

	commands.GetInstance().StoreServiceInfo(serviceInfo.ServiceName, command)
	return nil
}

func (CommandValidator) CheckCommand(serviceName string, command []string) error {
	fullCommand := strings.Join(command, " ")
	if injectionchecker.HasInjectionOperator(fullCommand) {
		return errors.New(foundInjectionCommand)
	}

	expected, err := commands.GetInstance().GetServiceFileName(serviceName)
	if err != nil {
		return err
	}

	realCommand, err := getExecutableName(command[0])
	if err != nil {
		return err
	}

	if expected != realCommand {
		return errors.New(notMatchedServiceWithExecutable)
	}

	return nil
}

func getExecutableName(str string) (string, error) {
	var command string
	commandList := strings.Split(str, "/")
	switch {
	case str == "":
		return "", errors.New(notFoundExecutableFile)
	case len(commandList) == 1:
		command = commandList[0]
	default:
		command = commandList[len(commandList)-1]
	}

	return command, nil
}
