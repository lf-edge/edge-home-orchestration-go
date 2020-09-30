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
package util

import (
	"bufio"
	discoveryMocks "controller/discoverymgr/mocks"
	"errors"
	"github.com/golang/mock/gomock"
	"log"
	"os"
	"testing"
)

var (
	defaultYamlFilePath   = "YamlTestFile.txt"
	defaultYamlValue      = "name:TestValue"
	defaultConfigFilePath = "ConfigTestFile.txt"
	defaultConfigValue    = []string{"Name:TestValue",
		"Profile:TestProfile", "Host:TestHost"}
	binPath       = "storagemgr/util"
	mockDiscovery *discoveryMocks.MockDiscovery
)

func createTestFiles() error {
	curr_dir, _ := os.Getwd()
	log.Println("the path before directory coming out is", curr_dir)

	err := os.Chdir("../../")
	if err != nil {
		log.Fatalln(err)
	}

	f, err := os.Create(defaultYamlFilePath)
	if err != nil {
		log.Println("The err obtained are ", err)
		return err
	}

	defer f.Close()

	_, err = f.WriteString(defaultYamlValue)
	if err != nil {
		return err
	}

	f.Sync()
	f, err = os.Create(defaultConfigFilePath)
	if err != nil {
		log.Println("The err obtained are ", err)
		return err
	}

	defer f.Close()
	datawriter := bufio.NewWriter(f)

	for _, data := range defaultConfigValue {
		_, _ = datawriter.WriteString(data + "\n")

	}
	datawriter.Flush()

	err = os.Chdir(binPath)
	if err != nil {
		log.Fatalf("[%s] Error in directory change: %s", logPrefix, err.Error())
	}

	return nil
}

func init() {
	createTestFiles()
}

func TestGetUuid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("Success", func(t *testing.T) {

		mockDiscovery.EXPECT().GetDeviceID().Return("edge-orchestration-7e841a12-a083-48fe-8040-83815e0b7410", nil).AnyTimes()

		deviceID := getUuid()
		log.Println(deviceID)
		if deviceID == "" {
			t.Error("Empty deviceID")
		} else {
			t.Log(deviceID)
		}
	})
	t.Run("Error", func(t *testing.T) {

		mockDiscovery.EXPECT().GetDeviceID().Return("", errors.New("")).AnyTimes()

		deviceID := getUuid()
		log.Println(deviceID)
		if deviceID == "" {
			t.Error("Empty deviceID")
		}
	})
}
func TestMapYamlFile(t *testing.T) {
	err := createTestFiles()
	if err != nil {
		t.Error(err)
	}
	t.Log("Files is created successfully")
	filePath := "/" + defaultYamlFilePath
	MapYamlFile(filePath)
}

func TestMapConfigFile(t *testing.T) {
	filePath := "/" + defaultConfigFilePath
	//strArray1 := [3]string{"Japan", "Australia", "Germany"}
	s := []string{"test", "test1"}

	MapConfigFile(filePath, s)
}

func createMockIns(ctrl *gomock.Controller) {
	mockDiscovery = discoveryMocks.NewMockDiscovery(ctrl)
	discoverIns = mockDiscovery
}
