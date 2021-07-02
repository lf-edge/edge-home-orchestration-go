/*******************************************************************************
 * Copyright 2021 Samsung Electronics All Rights Reserved.
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

package config

import (
	toml "github.com/pelletier/go-toml"
)

type Writable struct {
	LogLevel string `toml:"LogLevel,omitempty"`
}

type Service struct {
	Host                string   `toml:"Host"`
	Port                int      `toml:"Port"`
	ConnectionRetries   int      `toml:"ConnectionRetries,omitempty"`
	Labels              []string `toml:"Labels"`
	OpenMsg             string   `toml:"OpenMsg,omitempty"`
	Timeout             int      `toml:"Timeout,omitempty"`
	EnableAsyncReadings bool     `toml:"EnableAsyncReadings,omitempty"`
	AsyncBufferSize     int      `toml:"AsyncBufferSize,omitempty"`
}

type Registry struct {
	Host          string `toml:"Host"`
	Port          int    `toml:"Port"`
	Type          string `toml:"Type"`
	CheckInterval string `toml:"CheckInterval,omitempty"`
	FailLimit     int    `toml:"FailLimit,omitempty"`
	FailWaitTime  int    `toml:"FailWaitTime,omitempty"`
}

type Device struct {
	DataTransform  bool   `toml:"DataTransform,omitempty"`
	InitCmd        string `toml:"InitCmd"`
	InitCmdArgs    string `toml:"InitCmdArgs"`
	MaxCmdOps      int    `toml:"MaxCmdOps,omitempty"`
	MaxCmdValueLen int    `toml:"MaxCmdValueLen,omitempty"`
	RemoveCmd      string `toml:"RemoveCmd"`
	RemoveCmdArgs  string `toml:"RemoveCmdArgs"`
	ProfilesDir    string `toml:"ProfilesDir,omitempty"`
}

type ProtocolProperties map[string]string

type DeviceProperties struct {
	Name        string                        `toml:"Name"`
	Profile     string                        `toml:"Profile"`
	Description string                        `toml:"Description,omitempty"`
	Labels      []string                      `toml:"Labels,omitempty"`
	Protocols   map[string]ProtocolProperties `toml:"Protocols,omitempty"`
}

type DeviceList []DeviceProperties

type Client struct {
	Host     string `toml:"Host"`
	Port     int    `toml:"Port"`
	Protocol string `toml:"Protocol"`
	Timeout  int    `toml:"Timeout,omitempty"`
}

type Clients map[string]Client

type Toml struct {
	Writable
	Service
	Registry
	Device
	DeviceList
	Clients
}

var (
	tomlInfo Toml
)

func SetWritable(level string) {
	tomlInfo.Writable = Writable{LogLevel: level}
}

func SetService(host string, port int, label []string) {
	tomlInfo.Service = Service{
		Host:                host,
		Port:                port,
		ConnectionRetries:   20,
		Labels:              label,
		OpenMsg:             "REST device started",
		Timeout:             5000,
		EnableAsyncReadings: true,
		AsyncBufferSize:     16}
}

func SetRegistry(host string, port int) {
	tomlInfo.Registry = Registry{
		Host:          host,
		Port:          port,
		Type:          "consul",
		CheckInterval: "10s",
		FailLimit:     3,
		FailWaitTime:  10}
}

func SetDevice(dataTransform bool, initCmd string, initCmdArgs string, maxCmdOps int,
	maxCmdValueLen int, removeCmd string, removeCmdArgs string, profilesDir string) {
	tomlInfo.Device = Device{
		DataTransform:  dataTransform,
		InitCmd:        initCmd,
		InitCmdArgs:    initCmdArgs,
		MaxCmdOps:      maxCmdOps,
		MaxCmdValueLen: maxCmdValueLen,
		RemoveCmd:      removeCmd,
		RemoveCmdArgs:  removeCmdArgs,
		ProfilesDir:    profilesDir}
}

func SetDeviceList(name string, profile string, description string, label []string) {
	tomlInfo.DeviceList = DeviceList{DeviceProperties{
		Name:        name,
		Profile:     profile,
		Description: description,
		Labels:      label,
		Protocols:   map[string]ProtocolProperties{"other": {}}}}
}

func SetClients(host string, protocol string, timeout int) {
	tomlInfo.Clients = Clients{
		"core-data": Client{Host: host,
			Port:     59880,
			Protocol: protocol,
			Timeout:  timeout},
		"core-metadata": Client{Host: host,
			Port:     59881,
			Protocol: protocol,
			Timeout:  timeout}}
}

func TomlMarshal() (b []byte, err error) {
	return toml.Marshal(tomlInfo)
}
