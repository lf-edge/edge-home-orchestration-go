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

// Package verifier ensures that only allowed containers (images) are launched
package verifier

import (
	"errors"
	"os"
	"strings"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
)

// cwl - Container White List
const (
	cwlFileName = "containerwhitelist.txt"

	ErrorNone         = "ERROR_NONE"
	InvalidParameter  = "INVALID_PARAMETER"
	SecureMgrError    = "INTERNAL_SECUREMGR_ERROR"
	NotAllowedCommand = "NOT_ALLOWED_COMMAND"

	cannotCreateFile = "cannot create file: "
)

// VerificationImpl structure
type VerificationImpl struct{}

var (
	containerWhiteList []string
	logPrefix          = "[securemgr: verifier] "
	log                = logmgr.GetInstance()
	verifierIns        *VerificationImpl
	initialized        = false
	cwlFilePath        = ""
)

// RequestDescInfo describes the requested container
type RequestDescInfo struct {
	//ContainerName string
	ContainerHash string
}

// RequestVerifierConf describes the request configuration
type RequestVerifierConf struct {
	SecureInsName string
	CmdType       string
	Desc          []RequestDescInfo
}

// ResponseVerifierConf describes the verifier configuration response
type ResponseVerifierConf struct {
	Message       string
	SecureCmpName string
}

// Conf is the interface implemented by external REST API
type Conf interface {
	RequestVerifierConf(containerInfo RequestVerifierConf) ResponseVerifierConf
}

func init() {
	verifierIns = new(VerificationImpl)
}

// initContainerWhiteList fills the containerWhiteList by reading the information
// from the file if it exists or creates it otherwise
func initContainerWhiteList() error {
	fileContent, err := os.ReadFile(cwlFilePath)
	if err != nil {
		containerWhiteList = nil
		err = os.WriteFile(cwlFilePath, []byte(""), 0666)
		if err != nil {
			log.Error(logPrefix, cannotCreateFile, cwlFileName, ": ", err)
		}
	} else {
		containerWhiteList = strings.Split(string(fileContent), "\n")
		if len(containerWhiteList) > 0 {
			containerWhiteList = containerWhiteList[:len(containerWhiteList)-1]
		}
		//for _, whitelistItem := range containerWhiteList {
		//	log.Debug(logPrefix, whitelistItem)
		//}
	}
	return err
}

// GetInstance gives the VerificationImpl singletone instance
func GetInstance() *VerificationImpl {
	return verifierIns
}

func containerHashIsInWhiteList(hash string) error {
	for _, whitelistItem := range containerWhiteList {
		if hash == whitelistItem {
			return nil
		}
	}
	return errors.New("container's hash: " + hash + " is not in container white list")
}

func getIndexHashInContainerName(containerName string) (int, error) {
	digestIndex := strings.Index(containerName, "@sha256:")
	if digestIndex == -1 {
		return -1, errors.New("container name does not contain a digest")
	}
	digestIndex = digestIndex + len("@sha256:")
	return digestIndex, nil
}

// Init sets the environments for securemgr
func Init(cwlPath string) {
	if _, err := os.Stat(cwlPath); err != nil {
		err := os.MkdirAll(cwlPath, os.ModePerm)
		if err != nil {
			log.Panic(logPrefix, "Failed to create cwlPath", cwlPath, ": ", err)
		}
	}
	cwlFilePath = cwlPath + "/" + cwlFileName
	initContainerWhiteList()
	initialized = true
}

// ContainerIsInWhiteList checks if the containerName is in containerWhiteList
func (VerificationImpl) ContainerIsInWhiteList(containerName string) error {
	if !initialized {
		return nil
	}
	index, err := getIndexHashInContainerName(containerName)
	if err != nil {
		return err
	}
	return containerHashIsInWhiteList(containerName[index:])
}

// addHashToContainerWhiteList add the hash to containerWhiteList
// if it exists then ignore this command
func addHashToContainerWhiteList(hash string) error {
	fileContent, err := os.ReadFile(cwlFilePath)
	if err != nil {
		fileContentStr := hash + "\n"
		err = os.WriteFile(cwlFilePath, []byte(fileContentStr), 0666)
		if err != nil {
			log.Error(logPrefix, cannotCreateFile, cwlFileName, err)
			return err
		}
		containerWhiteList = append(containerWhiteList, hash)
	} else {
		fileContentStr := string(fileContent)
		log.Debug(logPrefix, "Len filecontentstr = ", len(fileContentStr))
		containerWhiteList = strings.Split(fileContentStr, "\n")
		if len(containerWhiteList) > 0 {
			containerWhiteList = containerWhiteList[:len(containerWhiteList)-1]
		}
		for _, whitelistItem := range containerWhiteList {
			if whitelistItem == hash {
				log.Info(logPrefix, "container's hash ", logmgr.SanitizeUserInput(hash), " already exists in container white list") // lgtm [go/log-injection]
				return nil
			}
		}
		fileContentStr = fileContentStr + hash + "\n"
		err = os.WriteFile(cwlFilePath, []byte(fileContentStr), 0666)
		if err != nil {
			log.Error(logPrefix, cannotCreateFile, cwlFileName, err)
			return err
		}
		containerWhiteList = append(containerWhiteList, hash)
	}
	return nil
}

// delHashFromContainerWhiteList deletes the hash from containerWhiteList file,
// if hash is absent then ignore this command
func delHashFromContainerWhiteList(hash string) error {
	fileContent, err := os.ReadFile(cwlFilePath)
	if err != nil {
		err = os.WriteFile(cwlFilePath, []byte(""), 0666)
		if err != nil {
			log.Error(logPrefix, cannotCreateFile, cwlFileName, err)
			return err
		}
	} else {
		fileContentStr := string(fileContent)
		digestIndex := strings.Index(fileContentStr, hash)
		if digestIndex == -1 {
			log.Info(logPrefix, "hash: ", logmgr.SanitizeUserInput(hash), " not found in ", cwlFileName) // lgtm [go/log-injection]
			return nil
		}
		fileContentStr = fileContentStr[:digestIndex] + fileContentStr[digestIndex+65:]
		err = os.WriteFile(cwlFilePath, []byte(fileContentStr), 0666)
		if err != nil {
			log.Error(logPrefix, cannotCreateFile, cwlFileName, err)
			return err
		}
		containerWhiteList = strings.Split(fileContentStr, "\n")
		if len(containerWhiteList) > 0 {
			containerWhiteList = containerWhiteList[:len(containerWhiteList)-1]
		}
		log.Info(logPrefix, "hash: ", logmgr.SanitizeUserInput(hash), " successfully deleted from ", cwlFileName, " file") // lgtm [go/log-injection]
	}
	return nil
}

// delAllHashFromContainerWhiteList deletes all hashes from containerWhiteList file
func delAllHashFromContainerWhiteList() error {
	err := os.WriteFile(cwlFilePath, []byte(""), 0666)
	if err != nil {
		log.Error(logPrefix, cannotCreateFile, cwlFileName, err)
		return err
	}
	containerWhiteList = nil
	log.Info(logPrefix, "all hashes successfully deleted from ", cwlFileName, " file")
	return err
}

// printAllHashFromContainerWhiteList displays all records from containerWhiteList file,
func printAllHashFromContainerWhiteList() {
	if containerWhiteList != nil {
		for idx, whitelistItem := range containerWhiteList {
			log.Info(logPrefix, "container's hash[", idx, "]: ", logmgr.SanitizeUserInput(whitelistItem)) // lgtm [go/log-injection]
		}
	} else {
		log.Info(logPrefix, "container white list is empty")
	}
}

// RequestVerifierConf is Verifier configuration request handler
func (verifier *VerificationImpl) RequestVerifierConf(containerInfo RequestVerifierConf) ResponseVerifierConf {
	log.Info(logPrefix, "command type: ", logmgr.SanitizeUserInput(containerInfo.CmdType)) // lgtm [go/log-injection]
	switch containerInfo.CmdType {
	case "addHashCWL":
		for _, containerDesc := range containerInfo.Desc {
			err := addHashToContainerWhiteList(containerDesc.ContainerHash)
			if err != nil {
				return ResponseVerifierConf{
					Message:       SecureMgrError,
					SecureCmpName: "verifier",
				}
			}
		}
	case "delHashCWL":
		for _, containerDesc := range containerInfo.Desc {
			err := delHashFromContainerWhiteList(containerDesc.ContainerHash)
			if err != nil {
				return ResponseVerifierConf{
					Message:       SecureMgrError,
					SecureCmpName: "verifier",
				}
			}
		}
	case "delAllHashCWL":
		err := delAllHashFromContainerWhiteList()
		if err != nil {
			return ResponseVerifierConf{
				Message:       SecureMgrError,
				SecureCmpName: "verifier",
			}
		}
	case "printAllHashCWL":
		printAllHashFromContainerWhiteList()
	default:
		log.Info(logPrefix, "command does not supported: ", logmgr.SanitizeUserInput(containerInfo.CmdType)) // lgtm [go/log-injection]
		return ResponseVerifierConf{
			Message:       NotAllowedCommand,
			SecureCmpName: "verifier",
		}
	}
	return ResponseVerifierConf{
		Message:       ErrorNone,
		SecureCmpName: "verifier",
	}
}
