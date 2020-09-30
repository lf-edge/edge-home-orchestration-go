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
	"controller/discoverymgr"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	discoverIns discoverymgr.Discovery
	uuId        string
	test        string
)

const (
	Name       = "Name"
	Profile    = "Profile"
	ProfileDir = "ProfilesDir"
	Host       = "Host"
)
const logPrefix = "[storagemgr]"

func init() {
	discoverIns = discoverymgr.GetInstance()
}

//Get the UUid value for mapping
func getUuid() string {

	log.Println(logPrefix, "GetUuid() +")
	//obtain the uuid from userid file
	deviceId, err := discoverIns.GetDeviceID()
	if err != nil {
		log.Println("err occured is ", err)
		return ""
	}
	log.Println(logPrefix, "The device id obtained :", deviceId)

	//splitting to remove the edge -orchestration
	extract_uuid := regexp.MustCompile("-")
	splitUuid := extract_uuid.Split(deviceId, -1)
	strtIndex := 2
	endIndex := 6

	for i := strtIndex; i <= endIndex; i++ {
		uuId += splitUuid[i]
		uuId += "-"
	}

	//there is enter in the file that is also read
	removeEnter := strings.Split(uuId, "\n")

	uuId = removeEnter[0]

	log.Println(logPrefix, "Final UUid obtained is", uuId)

	return uuId
}

//Map the yaml file maps the edge-orchestration uuid with the name field in yaml file
func MapYamlFile(yamlFileName string) {

	log.Println(logPrefix, "MapYamlFile() +")

	uuid := getUuid()
	uuId = strconv.Quote(uuid)
	curr_dir, _ := os.Getwd()
	log.Println(logPrefix, "the path before directory coming out is", curr_dir)

	err := os.Chdir("../../")
	if err != nil {
		log.Fatalln(err)
	}

	pwd, err := os.Getwd()
	yamlFilePath := pwd + yamlFileName
	log.Println(logPrefix, "the path is ", pwd)
	input, err := ioutil.ReadFile(yamlFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	lines := strings.Split(string(input), "\n")
	//first line is always name field
	lines[0] = "name: " + uuId
	log.Println(lines[0])
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(yamlFilePath, []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

//maps the config file name filed with uuid mapped in yaml file
func MapConfigFile(configFileName string, hostIPAddr []string) {

	var flag bool
	log.Println(logPrefix, "MapConfigFile() +")

	flag = false

	pwd, err := os.Getwd()

	configFilePath := pwd + configFileName

	log.Println(logPrefix, "the path is ", configFilePath)

	input, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, Name) {
			lines[i] = "Name = " + uuId
		}
		if strings.Contains(line, Profile) && !strings.Contains(line, ProfileDir) {
			lines[i] = " Profile = " + uuId
		}
		if strings.Contains(line, Host) && !flag {
			lines[i] = "Host = " + strconv.Quote(hostIPAddr[0])
			flag = true
		}

	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(configFilePath, []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(logPrefix, "Returning from the uuidmapping.go")

}
