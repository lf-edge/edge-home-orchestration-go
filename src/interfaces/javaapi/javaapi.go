/*******************************************************************************
 * Copyright 2019 Samsung Electronics All Rights Reserved.
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

// Package javaapi provides Java interface for orchestration
package javaapi

import (
	"db/bolt/wrapper"
	"log"
	"strings"
	"sync"

	"common/logmgr"

	configuremgr "controller/configuremgr/container"
	"controller/discoverymgr"
	scoringmgr "controller/scoringmgr"
	"controller/servicemgr"
	"controller/servicemgr/executor/nativeexecutor"

	"orchestrationapi"

	"restinterface/cipher/sha256"
	"restinterface/client/restclient"
	"restinterface/internalhandler"
	"restinterface/route"
)

type RequestServiceInfo struct {
	ExecutionType string
	ExeCmd        []string
}

type ReqeustService struct {
	ServiceName string
	ServiceInfo []RequestServiceInfo
}

func (r *ReqeustService) SetExecutionCommand(execType, command string) {
	switch execType {
	case "native", "android", "container":
	default:
		log.Printf("[%s] unexpected execution type: %s", logPrefix, execType)
		return
	}

	args := strings.Split(command, " ")

	for _, info := range r.ServiceInfo {
		if info.ExecutionType == execType {
			info.ExeCmd = make([]string, len(args))
			copy(info.ExeCmd, args)
			return
		}
	}
	info := RequestServiceInfo{ExecutionType: execType}
	info.ExeCmd = make([]string, len(args))
	copy(info.ExeCmd, args)

	r.ServiceInfo = append(r.ServiceInfo, info)
}

func (r ReqeustService) GetExecutionCommand(execType string) string {
	switch execType {
	case "native", "android", "container":
		for _, info := range r.ServiceInfo {
			if info.ExecutionType == execType {
				return strings.Join(info.ExeCmd, " ")
			}
		}
	}
	return ""
}

type TargetInfo struct {
	ExecutionType string
	Target        string
}

type ResponseService struct {
	Message          string
	ServiceName      string
	RemoteTargetInfo TargetInfo
}

func (r ResponseService) GetExecutedType() string {
	return r.RemoteTargetInfo.ExecutionType
}

func (r ResponseService) GetTarget() string {
	return r.RemoteTargetInfo.Target
}

const logPrefix = "interface"

// Handle Platform Dependencies
const (
	platform      = "android"
	executionType = "apk"

	edgeDir = "/storage/emulated/0/Android/data/com.samsung.orchestration.service/files/"

	logPath    = edgeDir + "log/edge-orchestration"
	configPath = edgeDir + "apps"
	dbPath     = edgeDir + "db"

	cipherKeyFilePath = edgeDir + "orchestration_userID.txt"
	deviceIDFilePath  = edgeDir + "orchestration_deviceID.txt"
)

var orcheEngine orchestrationapi.Orche

// OrchestrationInit runs orchestration service and discovers other orchestration services in other devices
func OrchestrationInit() (errCode int) {

	logmgr.Init(logPath)
	log.Printf("[%s] OrchestrationInit", logPrefix)

	wrapper.SetBoltDBPath(dbPath)

	restIns := restclient.GetRestClient()
	restIns.SetCipher(sha256.GetCipher(cipherKeyFilePath))

	servicemgr.GetInstance().SetClient(restIns)

	builder := orchestrationapi.OrchestrationBuilder{}
	builder.SetWatcher(configuremgr.GetInstance(configPath))
	builder.SetDiscovery(discoverymgr.GetInstance())
	builder.SetScoring(scoringmgr.GetInstance())
	builder.SetService(servicemgr.GetInstance())
	builder.SetExecutor(nativeexecutor.GetInstance()) // TODO modify for android if needed
	builder.SetClient(restIns)

	orcheEngine = builder.Build()
	if orcheEngine == nil {
		log.Fatalf("[%s] Orchestaration initalize fail", logPrefix)
		return
	}

	orcheEngine.Start(deviceIDFilePath, platform, executionType)

	restEdgeRouter := route.NewRestRouter()

	internalapi, err := orchestrationapi.GetInternalAPI()
	if err != nil {
		log.Fatalf("[%s] Orchestaration internal api : %s", logPrefix, err.Error())
	}
	ihandle := internalhandler.GetHandler()
	ihandle.SetOrchestrationAPI(internalapi)
	ihandle.SetCipher(sha256.GetCipher(cipherKeyFilePath))
	restEdgeRouter.Add(ihandle)

	restEdgeRouter.Start()

	log.Println(logPrefix, "orchestration init done")

	errCode = 0

	return
}

// OrchestrationRequestService performs request from service applications who uses orchestration service
func OrchestrationRequestService(request *ReqeustService) *ResponseService {
	log.Printf("[%s] OrchestrationRequestService", logPrefix)
	log.Println("Service name: ", request.ServiceName)

	externalAPI, err := orchestrationapi.GetExternalAPI()
	if err != nil {
		log.Fatalf("[%s] Orchestaration external api : %s", logPrefix, err.Error())
	}

	changed := orchestrationapi.ReqeustService{ServiceName: request.ServiceName}

	changed.ServiceInfo = make([]orchestrationapi.RequestServiceInfo, len(request.ServiceInfo))
	for idx, info := range request.ServiceInfo {
		changed.ServiceInfo[idx].ExecutionType = info.ExecutionType
		copy(changed.ServiceInfo[idx].ExeCmd, info.ExeCmd)
	}

	response := externalAPI.RequestService(changed)
	log.Println("response : ", response)

	ret := &ResponseService{
		Message:     response.Message,
		ServiceName: response.ServiceName,
		RemoteTargetInfo: TargetInfo{
			ExecutionType: response.RemoteTargetInfo.ExecutionType,
			Target:        response.RemoteTargetInfo.Target,
		},
	}
	return ret
}

var count int
var mtx sync.Mutex

// PrintLog provides logging interface
func PrintLog(cMsg string) (count int) {
	mtx.Lock()
	msg := cMsg
	defer mtx.Unlock()
	log.Printf(msg)
	count++
	return
}
