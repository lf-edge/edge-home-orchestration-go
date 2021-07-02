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
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"github.com/edgexfoundry/device-sdk-go/v2"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/startup"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/storagemgr/config"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/storagemgr/storagedriver"
)

const (
	dataStorageService    = "datastorage"
	dataStorageConfFolder = "res"
	deviceIDFilePath      = "/var/edge-orchestration/device/orchestration_deviceID.txt"
)

type Storage interface {
	GetStatus() int
	StartStorage(host string) error
	BuildConfiguration(host string) error
}

// status = 0 : No action
// status = 1 : Completed to build configuration files
// status = 2 : Running DataStorage
type StorageImpl struct {
	sd     storagedriver.StorageDriver
	status int
}

var (
	deviceName string
	storageIns *StorageImpl
)

func init() {
	deviceID, _ := getDeviceID()
	deviceName   = dataStorageService + "-" + deviceID
	storageIns   = &StorageImpl{
		sd    : storagedriver.StorageDriver{},
		status: 0,
	}
}

// GetInstance returns the instance of DataStorage
func GetInstance() Storage {
	return storageIns
}

// GetStatus returns the status value in StorageImpl
func (s *StorageImpl) GetStatus() int {
	return s.status
}

// StartStorage starts a server in terms of DataStorage
func (s *StorageImpl) StartStorage(host string) (err error) {
	if checkServiceInEnv() {
		if err = s.BuildConfiguration("127.0.0.1"); err != nil {
			return
		}
	} else if len(host) > 0 {
		if err = s.BuildConfiguration(host); err != nil {
			return
		}
	}

	if _, err := os.Stat(dataStorageConfFolder+"/configuration.toml"); err == nil {
		if s.status < 2 {
			go startup.Bootstrap(dataStorageService, device.Version, &storageIns.sd)
			s.status = 2
			return nil
		}
	}
	return errors.New("could not initiate storageManager")
}

// BuildConfiguration save configuration files, such as configuration.toml and yaml files in res folder
func (s *StorageImpl) BuildConfiguration(host string) (err error) {
	s.status = 0
	if err = saveToml(host); err != nil {
		return
	}
	if err = saveYaml(); err != nil {
		return
	}
	s.status = 1
	return nil
}

func checkServiceInEnv() bool {
	if strings.Contains(os.Getenv("SERVICE"), "DataStorage") {
		return true
	}
	return false
}

func saveToml(host string) (err error) {
	config.SetWritable("DEBUG")
	config.SetService("", 49986, nil)
	config.SetRegistry(host, 8500)
	config.SetDevice(true, "", "", 128, 256, "", "", "./res")
	config.SetDeviceList(deviceName, deviceName, "RESTful Device", []string{"rest", "json"})
	config.SetClients(host, "http", 5000)

	b, err := config.TomlMarshal()
	if err == nil {
		err = ioutil.WriteFile(dataStorageConfFolder+"/configuration.toml", b, 0644)
	}
	return
}

func saveYaml() (err error) {
	manufacture    := "Home Edge"
	model          := "Home Edge"
	label          := []string{"rest", "json", "int", "float", "jpeg", "png", "string"}
	description    := "REST Device"
	propertyJson   := config.Property{
				Value:config.PropertyDetail{
					Type:"String",
					ReadWrite:"RW",
					MediaType:"application/json"},
				Units:config.PropertyDetail{
					Type:"String",
					ReadWrite:"R"}}
	propertyInt    := config.Property{
				Value:config.PropertyDetail{
					Type:"Int64",
					ReadWrite:"RW",
					MediaType:"text/plain"},
				Units:config.PropertyDetail{
					Type:"String",
					ReadWrite:"R"}}
	propertyFloat  := config.Property{
				Value:config.PropertyDetail{
					Type:"Float64",
					ReadWrite:"RW",
					MediaType:"text/plain"},
				Units:config.PropertyDetail{
					Type:"String",
					ReadWrite:"R"}}
	propertyJpeg   := config.Property{
				Value:config.PropertyDetail{
					Type:"Binary",
					ReadWrite:"RW",
					MediaType:"image/jpeg"},
				Units:config.PropertyDetail{
					Type:"String",
					ReadWrite:"R"}}
	propertyPng    := config.Property{
				Value:config.PropertyDetail{
					Type:"Binary",
					ReadWrite:"RW",
					MediaType:"image/png"},
				Units:config.PropertyDetail{
					Type:"String",
					ReadWrite:"R"}}
	propertyString := config.Property{
				Value:config.PropertyDetail{
					Type:"String",
					ReadWrite:"RW",
					MediaType:"text/plain"},
				Units:config.PropertyDetail{
					Type:"String",
					ReadWrite:"R"}}
	resource       := []config.DeviceResource{
				config.DeviceResource{
					Name:"json",
					Properties:propertyJson},
				config.DeviceResource{
					Name:"int",
					Properties:propertyInt},
				config.DeviceResource{
					Name:"float",
					Properties:propertyFloat},
				config.DeviceResource{
					Name:"jpeg",
					Properties:propertyJpeg},
				config.DeviceResource{
					Name:"png",
					Properties:propertyPng},
				config.DeviceResource{
					Name:"string",
					Properties:propertyString}}

	config.SetYaml(deviceName, manufacture, model, description, label, resource)
	b, err := config.YamlMarshal()
	if err == nil {
		err = ioutil.WriteFile(dataStorageConfFolder+"/datastorage-device.yaml", b, 0644)
	}
	return
}

func getDeviceID() (string, error) {
	UUIDv4, err := ioutil.ReadFile(deviceIDFilePath)
	if err != nil {
		return "", err
	}
	return string(UUIDv4), nil
}
