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
	"strings"

	"common/logmgr"
	"common/sigmgr"

	configuremgr "controller/configuremgr/container"
	"controller/discoverymgr"
	"controller/mnedcmgr"
	"controller/scoringmgr"
	"controller/securemgr/authenticator"
	"controller/securemgr/verifier"
	"controller/servicemgr"
	executor "controller/servicemgr/executor/containerexecutor"

	"orchestrationapi"

	"restinterface/cipher/dummy"
	"restinterface/cipher/sha256"
	"restinterface/client/restclient"
	"restinterface/externalhandler"
	"restinterface/internalhandler"
	"restinterface/route"

	"db/bolt/wrapper"
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

	cipherKeyFilePath = edgeDir + "/user/orchestration_userID.txt"
	deviceIDFilePath  = edgeDir + "/device/orchestration_deviceID.txt"
	mnedcServerConfig = edgeDir + "/mnedc/client.config"
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
	if strings.Contains(buildTags, "secure") {
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
	discoverymgr.GetInstance().SetClient(restIns)

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

	if strings.Contains(buildTags, "mnedcserver") {
		if isSecured {
			mnedcmgr.GetServerInstance().SetCipher(dummy.GetCipher(cipherKeyFilePath))
			mnedcmgr.GetServerInstance().SetCertificateFilePath(certificateFilePath)
		} else {
			mnedcmgr.GetServerInstance().SetCipher(sha256.GetCipher(cipherKeyFilePath))
		}
		go mnedcmgr.GetServerInstance().StartMNEDCServer(deviceIDFilePath)
	} else if strings.Contains(buildTags, "mnedcclient") {
		if isSecured {
			mnedcmgr.GetClientInstance().SetCertificateFilePath(certificateFilePath)
		}
		go mnedcmgr.GetClientInstance().StartMNEDCClient(deviceIDFilePath, mnedcServerConfig)
	}

	return nil
}
