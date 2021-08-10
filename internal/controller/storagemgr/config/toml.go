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

// Writable contains the configuration for the DataStorage log.
type Writable struct {
	LogLevel string `toml:"LogLevel,omitempty"`
}

// Service contains the configuration for the DataStorage service.
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

// Registry contains the configuration for the DataStorage in terms of the Registry server.
type Registry struct {
	Host          string `toml:"Host"`
	Port          int    `toml:"Port"`
	Type          string `toml:"Type"`
	CheckInterval string `toml:"CheckInterval,omitempty"`
	FailLimit     int    `toml:"FailLimit,omitempty"`
	FailWaitTime  int    `toml:"FailWaitTime,omitempty"`
}

// Device contains the configuration for the DataStorage device.
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

// ProtocolProperties is a map of device protocols.
type ProtocolProperties map[string]string

// DeviceProperties contains the configuration for a specific device.
type DeviceProperties struct {
	Name        string                        `toml:"Name"`
	Profile     string                        `toml:"Profile"`
	Description string                        `toml:"Description,omitempty"`
	Labels      []string                      `toml:"Labels,omitempty"`
	Protocols   map[string]ProtocolProperties `toml:"Protocols,omitempty"`
}

// DeviceList is a list of the DevicesPropertieses.
type DeviceList []DeviceProperties

// Client contains other service information for DataStroage.
type Client struct {
	Host     string `toml:"Host"`
	Port     int    `toml:"Port"`
	Protocol string `toml:"Protocol"`
	Timeout  int    `toml:"Timeout,omitempty"`
}

// Clients is a map of Clients.
type Clients map[string]Client

// Toml contains the struct for building the DataStorage configuration file.
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

// SetWritable configures the Writable information.
func SetWritable(level string) {
	tomlInfo.Writable = Writable{LogLevel: level}
}

// SetService configures the Service information.
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

// SetRegistry configures the Registry information.
func SetRegistry(host string, port int) {
	tomlInfo.Registry = Registry{
		Host:          host,
		Port:          port,
		Type:          "consul",
		CheckInterval: "10s",
		FailLimit:     3,
		FailWaitTime:  10}
}

// SetDevice configures the Device's common information.
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

// SetDeviceList configures the specific Device information
func SetDeviceList(name string, profile string, description string, label []string) {
	tomlInfo.DeviceList = DeviceList{DeviceProperties{
		Name:        name,
		Profile:     profile,
		Description: description,
		Labels:      label,
		Protocols:   map[string]ProtocolProperties{"other": {}}}}
}

// SetClients configures the service information in terms of Data and Metadata.
func SetClients(host string, protocol string, timeout int) {
	tomlInfo.Clients = Clients{
		"Data": Client{Host: host,
			Port:     48080,
			Protocol: protocol,
			Timeout:  timeout},
		"Metadata": Client{Host: host,
			Port:     48081,
			Protocol: protocol,
			Timeout:  timeout}}
}

// TomlMarshal returns bytes for DataStorage configuration.
func TomlMarshal() (b []byte, err error) {
	return toml.Marshal(tomlInfo)
}

// GetServerIP is used to obtain server IP from the configuration.toml file
func GetServerIP(ConfigPath string) (string, int, error) {
	config, err := toml.LoadFile(ConfigPath)
	if err != nil {
		return "", 0, err
	}
	return config.Get("Clients.Data.Host").(string), (int)(config.Get("Clients.Data.Port").(int64)), nil
}
