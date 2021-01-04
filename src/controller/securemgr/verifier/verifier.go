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
package verifier

import (
	"errors"
	"io/ioutil"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
	"os"
	"strings"
)

// cwl - Container White List
const (
	cwlFileName = "containerwhitelist.txt"

	ERROR_NONE          = "ERROR_NONE"
	INVALID_PARAMETER   = "INVALID_PARAMETER"
	SECUREMGR_ERROR     = "INTERNAL_SECUREMGR_ERROR"
	NOT_ALLOWED_COMMAND = "NOT_ALLOWED_COMMAND"
)

// VerifierImpl structure
type VerifierImpl struct{}

var (
	containerWhiteList []string
	logPrefix          = "[securemgr: verifier]"
	log                = logmgr.GetInstance()
	verifierIns        *VerifierImpl
	initialized        = false
	cwlFilePath        = ""
)

type RequestDescInfo struct {
	//ContainerName string
	ContainerHash string
}

type RequestVerifierConf struct {
	SecureInsName string
	CmdType       string
	Desc          []RequestDescInfo
}

type ResponseVerifierConf struct {
	Message       string
	SecureCmpName string
}

// VerifierConf is the interface implemented by external REST API
type VerifierConf interface {
	RequestVerifierConf(containerInfo RequestVerifierConf) ResponseVerifierConf
}

func init() {
	verifierIns = new(VerifierImpl)
}

// initContainerWhiteList fills the containerWhiteList by reading the information
// from the file if it exists or creates it otherwise
func initContainerWhiteList() error {
	fileContent, err := ioutil.ReadFile(cwlFilePath)
	if err != nil {
		containerWhiteList = nil
		err = ioutil.WriteFile(cwlFilePath, []byte(""), 0666)
		if err != nil {
			log.Println(logPrefix, "cannot create "+cwlFileName+": ", err)
		}
	} else {
		containerWhiteList = strings.Split(string(fileContent), "\n")
		if len(containerWhiteList) > 0 {
			containerWhiteList = containerWhiteList[:len(containerWhiteList)-1]
		}
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
		if hash == whitelistItem {
			return true
		}
	}
	return false
}

func getIndexHashInContainerName(containerName string) (int, error) {
	digestIndex := strings.Index(containerName, "@sha256:")
	if digestIndex == -1 {
		return -1, errors.New("Container name does not contain a digest")
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
	cwlFilePath = cwlPath + "/" + cwlFileName
	initContainerWhiteList()
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
	if containerHashIsInWhiteList(containerName[index:]) {
		log.Printf("%s container's hash: %s is in container white list\n", logPrefix, containerName[index:])
		return nil
	} else {
		log.Printf("%s container's hash: %s is not in container white list\n", logPrefix, containerName[index:])
		return errors.New("container's hash is not in container white list")
	}
}

// addHashToContainerWhiteList add the hash to containerWhiteList
// if it exists then ignore this command
func addHashToContainerWhiteList(hash string) error {
	fileContent, err := ioutil.ReadFile(cwlFilePath)
	if err != nil {
		fileContentStr := hash + "\n"
		err = ioutil.WriteFile(cwlFilePath, []byte(fileContentStr), 0666)
		if err != nil {
			log.Printf("%s cannot create %s file: %s\n", logPrefix, cwlFileName, err)
			return err
		}
		containerWhiteList = append(containerWhiteList, hash)
	} else {
		fileContentStr := string(fileContent)
		log.Println(logPrefix, "Len filecontentstr = ", len(fileContentStr))
		containerWhiteList = strings.Split(fileContentStr, "\n")
		if len(containerWhiteList) > 0 {
			containerWhiteList = containerWhiteList[:len(containerWhiteList)-1]
		}
		for _, whitelistItem := range containerWhiteList {
			if whitelistItem == hash {
				log.Printf("%s container's hash %s already exists in conatiner white list\n", logPrefix, hash)
				return nil
			}
		}
		fileContentStr = fileContentStr + hash + "\n"
		err = ioutil.WriteFile(cwlFilePath, []byte(fileContentStr), 0666)
		if err != nil {
			log.Printf("%s cannot create %s file: %s\n", logPrefix, cwlFileName, err)
			return err
		}
		containerWhiteList = append(containerWhiteList, hash)
	}
	return nil
}

// delHashFromContainerWhiteList deletes the hash from containerWhiteList file,
// if hash is absent then ignore this command
func delHashFromContainerWhiteList(hash string) error {
	fileContent, err := ioutil.ReadFile(cwlFilePath)
	if err != nil {
		err = ioutil.WriteFile(cwlFilePath, []byte(""), 0666)
		if err != nil {
			log.Printf("%s cannot create %s file: %s\n", logPrefix, cwlFileName, err)
			return err
		}
	} else {
		fileContentStr := string(fileContent)
		digestIndex := strings.Index(fileContentStr, hash)
		if digestIndex == -1 {
			log.Printf("%s hash: %s not found in %s\n", logPrefix, hash, cwlFileName)
			return nil
		}
		fileContentStr = fileContentStr[:digestIndex] + fileContentStr[digestIndex+65:]
		err = ioutil.WriteFile(cwlFilePath, []byte(fileContentStr), 0666)
		if err != nil {
			log.Printf("%s cannot create %s file: %s\n", logPrefix, cwlFileName, err)
			return err
		}
		containerWhiteList = strings.Split(fileContentStr, "\n")
		if len(containerWhiteList) > 0 {
			containerWhiteList = containerWhiteList[:len(containerWhiteList)-1]
		}
		log.Printf("%s hash: %s successfully deleted from %s file\n", logPrefix, hash, cwlFileName)
	}
	return nil
}

// delAllHashFromContainerWhiteList deletes all hashes from containerWhiteList file
func delAllHashFromContainerWhiteList() error {
	err := ioutil.WriteFile(cwlFilePath, []byte(""), 0666)
	if err != nil {
		log.Printf("%s cannot create %s file: %s\n", logPrefix, cwlFileName, err)
		return err
	}
	containerWhiteList = nil
	log.Printf("%s all hashes successfully deleted from %s file\n", logPrefix, cwlFileName)
	return err
}

// printAllHashFromContainerWhiteList displays all records from containerWhiteList file,
func printAllHashFromContainerWhiteList() {
	if containerWhiteList != nil {
		for idx, whitelistItem := range containerWhiteList {
			log.Printf("%s container's hash[%d]: len = %d %s\n", logPrefix, idx, len(whitelistItem), whitelistItem)
		}
	} else {
		log.Println(logPrefix, "container white list is empty")
	}
}

func (verifier *VerifierImpl) RequestVerifierConf(containerInfo RequestVerifierConf) ResponseVerifierConf {
	log.Printf("%s command type: %s\n", logPrefix, containerInfo.CmdType)
	switch containerInfo.CmdType {
	case "addHashCWL":
		for _, containerDesc := range containerInfo.Desc {
			err := addHashToContainerWhiteList(containerDesc.ContainerHash)
			if err != nil {
				return ResponseVerifierConf{
					Message:       SECUREMGR_ERROR,
					SecureCmpName: "verifier",
				}
			}
		}
	case "delHashCWL":
		for _, containerDesc := range containerInfo.Desc {
			err := delHashFromContainerWhiteList(containerDesc.ContainerHash)
			if err != nil {
				return ResponseVerifierConf{
					Message:       SECUREMGR_ERROR,
					SecureCmpName: "verifier",
				}
			}
		}
	case "delAllHashCWL":
		err := delAllHashFromContainerWhiteList()
		if err != nil {
			return ResponseVerifierConf{
				Message:       SECUREMGR_ERROR,
				SecureCmpName: "verifier",
			}
		}
	case "printAllHashCWL":
		printAllHashFromContainerWhiteList()
	default:
		log.Println(logPrefix, "command does not supported: ", containerInfo.CmdType)
		return ResponseVerifierConf{
			Message:       NOT_ALLOWED_COMMAND,
			SecureCmpName: "verifier",
		}
	}
	return ResponseVerifierConf{
		Message:       ERROR_NONE,
		SecureCmpName: "verifier",
	}
}
