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
package securemgr

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// cwl - Container White List

const (
	cwlFileName 		= "containerwhitelist.txt"
	hashHelloWorld		= "fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b752"
)

// VerifierImpl structure
type VerifierImpl struct{}

var (
	containerWhiteList	[]string
	logPrefix			= "[verifier]"
	verifierIns 		*VerifierImpl
	initialized		  	= false
)

func init() {
	verifierIns = new(VerifierImpl)
}

// initContainerWhiteList fill the containerWhiteList by reading the information
// from the file if it exists or creates it otherwise
func initContainerWhiteList(cwlFilePath string) error {
	fileContent, err := ioutil.ReadFile(cwlFilePath)
	if err != nil {
		containerWhiteList = append(containerWhiteList, hashHelloWorld) // hello-world container image
		err = ioutil.WriteFile(cwlFilePath, []byte(hashHelloWorld + "\n"), 0666)
		if err != nil {
			log.Println(logPrefix, "Can't create " + cwlFileName + ": ", err)
		}
	} else {
		containerWhiteList = strings.Split(string(fileContent),"\n")
		//for _, whitelistItem := range containerWhiteList {
		//	log.Println(logPrefix, whitelistItem)
		//}
	}
	return err
}

// GetInstance gives the VerifierImpl singletone instance
func GetInstance() *VerifierImpl {
	return verifierIns
}

func containerHashIsInWhiteList(hash string) bool {
	for _, whitelistItem := range containerWhiteList {
		if (hash == whitelistItem) {
			return true
		}
	}
	return false
}

func getIndexHashInContainerName(containerName string) (int, error) {
	digestIndex := strings.Index(containerName, "@sha256:")
	if digestIndex == -1 {
		return -1, errors.New("Container name doesn't contain a digest")
	}
	digestIndex = digestIndex + len("@sha256:")
	return digestIndex, nil
}

// Init sets the environments for securemgr
func Init(cwlPath string) {
	if _, err := os.Stat(cwlPath); err != nil {
		err := os.MkdirAll(cwlPath, os.ModePerm)
		if err != nil {
			log.Panicf("Failed to create cwlPath %s: %s\n", cwlPath, err)
		}
	}
	initContainerWhiteList(cwlPath + "/" + cwlFileName)
	initialized = true
}
// ContainerIsInWhiteList checks if the containerName is in containerWhiteList
func (VerifierImpl) ContainerIsInWhiteList(containerName string) error {
	if initialized == false {
		return nil
	}
	index, err := getIndexHashInContainerName(containerName)
	if err != nil {
		return err
	}
	//log.Println(logPrefix, "SHA:" , containerName[index:])
	if containerHashIsInWhiteList(containerName[index:]) {
		log.Println(logPrefix, "Container is in whitelist")
		return nil
	} else {
		return errors.New("Container is not in whitelist")
	}
}
