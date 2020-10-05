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

// Package main provides REST interface for edge-orchestration
package main

import (
	"errors"
	"log"
	"os"
	"time"

	"common/logmgr"
	"common/sigmgr"

	configuremgr "controller/configuremgr/container"
	"controller/discoverymgr"
	"controller/scoringmgr"
	"controller/securemgr/authenticator"
	"controller/securemgr/verifier"
	"controller/servicemgr"
	executor "controller/servicemgr/executor/containerexecutor"
	"controller/storagemgr/storagedriver"
	storagemgr "controller/storagemgr/util"

	"orchestrationapi"

	"restinterface/cipher/dummy"
	"restinterface/cipher/sha256"
	"restinterface/client/restclient"
	"restinterface/externalhandler"
	"restinterface/internalhandler"
	"restinterface/route"

	"db/bolt/wrapper"
	"github.com/edgexfoundry/device-sdk-go"
	"github.com/edgexfoundry/device-sdk-go/pkg/startup"
)

const logPrefix = "interface"

// Handle Platform Dependencies
const (
	platform      = "docker"
	executionType = "container"

	edgeDir = "/var/edge-orchestration"

	logPath                = edgeDir + "/log"
	configPath             = edgeDir + "/apps"
	dbPath                 = edgeDir + "/data/db"
	certificateFilePath    = edgeDir + "/data/cert"
	containerWhiteListPath = edgeDir + "/data/cwl"
	passPhraseJWTPath      = edgeDir + "/data/jwt"
	binPath                = "GoMain/bin"
	YamlFileName           = "/GoMain/bin/res/sample-json-device.yaml"
	ConfigFileName         = "/GoMain/bin/res/configuration.toml"

	cipherKeyFilePath = edgeDir + "/user/orchestration_userID.txt"
	deviceIDFilePath  = edgeDir + "/device/orchestration_deviceID.txt"
	serviceName       = "datastorage"
)

var (
	commitID, version, buildTime string
	buildTags                    string
)

func main() {
	if err := orchestrationInit(); err != nil {
		log.Fatalf("[%s] Orchestaration initalize fail : %s", logPrefix, err.Error())
	}
	sigmgr.Watch()
}

// orchestrationInit runs orchestration service and discovers other orchestration services in other devices
func orchestrationInit() error {
	logmgr.Init(logPath)
	log.Printf("[%s] OrchestrationInit", logPrefix)
	log.Println(">>> commitID  : ", commitID)
	log.Println(">>> version   : ", version)
	log.Println(">>> buildTime : ", buildTime)
	log.Println(">>> buildTags : ", buildTags)
	wrapper.SetBoltDBPath(dbPath)

	isSecured := false
	if buildTags == "secure" {
		log.Println("Orchestration init with secure option")
		isSecured = true
	}

	if isSecured {
		verifier.Init(containerWhiteListPath)
		authenticator.Init(passPhraseJWTPath)
	}

	restIns := restclient.GetRestClient()

	if isSecured {
		restIns.SetCipher(dummy.GetCipher(cipherKeyFilePath))
	} else {
		restIns.SetCipher(sha256.GetCipher(cipherKeyFilePath))
	}

	servicemgr.GetInstance().SetClient(restIns)

	builder := orchestrationapi.OrchestrationBuilder{}
	builder.SetWatcher(configuremgr.GetInstance(configPath))
	builder.SetDiscovery(discoverymgr.GetInstance())
	builder.SetVerifierConf(verifier.GetInstance())
	builder.SetScoring(scoringmgr.GetInstance())
	builder.SetService(servicemgr.GetInstance())
	builder.SetExecutor(executor.GetInstance())
	builder.SetClient(restIns)

	orcheEngine := builder.Build()
	if orcheEngine == nil {
		log.Fatalf("[%s] Orchestaration initalize fail", logPrefix)
		return errors.New("fail to init orchestration")
	}

	orcheEngine.Start(deviceIDFilePath, platform, executionType)

	var restEdgeRouter *route.RestRouter
	if isSecured {
		restEdgeRouter = route.NewRestRouterWithCerti(certificateFilePath)
	} else {
		restEdgeRouter = route.NewRestRouter()
	}

	internalapi, err := orchestrationapi.GetInternalAPI()
	if err != nil {
		log.Fatalf("[%s] Orchestaration internal api : %s", logPrefix, err.Error())
	}
	ihandle := internalhandler.GetHandler()
	ihandle.SetOrchestrationAPI(internalapi)

	if isSecured {
		ihandle.SetCipher(dummy.GetCipher(cipherKeyFilePath))
		ihandle.SetCertificateFilePath(certificateFilePath)
	} else {
		ihandle.SetCipher(sha256.GetCipher(cipherKeyFilePath))
	}
	restEdgeRouter.Add(ihandle)

	// external rest api
	externalapi, err := orchestrationapi.GetExternalAPI()
	if err != nil {
		log.Fatalf("[%s] Orchestaration external api : %s", logPrefix, err.Error())
	}
	ehandle := externalhandler.GetHandler()
	ehandle.SetOrchestrationAPI(externalapi)
	ehandle.SetCipher(dummy.GetCipher(cipherKeyFilePath))
	restEdgeRouter.Add(ehandle)

	restEdgeRouter.Start()

	log.Println(logPrefix, "orchestration init done")
	storagemgr.MapYamlFile(YamlFileName)
	hostIPAddr, _ := discoverymgr.SetNetworkArgument()
	storagemgr.MapConfigFile(ConfigFileName, hostIPAddr)
	//the current directory is set to json file directory so changing it
	err = os.Chdir(binPath)
	if err != nil {
		log.Fatalf("[%s] Error in directory change: %s", logPrefix, err.Error())
	}
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("[%s] Error in directory change: %s", logPrefix, err.Error())
	}
	log.Println(logPrefix, "Path set is now", pwd)
	sd := storagedriver.StorageDriver{}
	startup.Bootstrap(serviceName, device.Version, &sd)

	// TODO remove this line
	// this line is for wait to initialize the mDNS server.
	time.Sleep(time.Second * 2)

	return nil
}
