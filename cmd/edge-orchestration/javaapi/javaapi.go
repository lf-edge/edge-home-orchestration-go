/*******************************************************************************
 * Copyright 2019-2020 Samsung Electronics All Rights Reserved.
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
	"bytes"
	"github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/wrapper"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/networkhelper"

	configuremgr "github.com/lf-edge/edge-home-orchestration-go/internal/controller/configuremgr/native"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/discoverymgr"
	scoringmgr "github.com/lf-edge/edge-home-orchestration-go/internal/controller/scoringmgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/securemgr/authenticator"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/securemgr/authorizer"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/securemgr/verifier"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/servicemgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/servicemgr/executor/androidexecutor"

	"github.com/lf-edge/edge-home-orchestration-go/internal/orchestrationapi"

	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/cipher/dummy"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/cipher/sha256"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/client/restclient"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/internalhandler"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/route"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/tls"
)

// Handle Platform Dependencies
const (
	logPrefix     = "interface"
	platform      = "android"
	executionType = "android"

	logStr                 = "/log"
	configStr              = "/apps"
	dbStr                  = "/data/db"
	certificateFile        = "/data/cert"
	containerWhiteListPath = "/data/cwl"
	passPhraseJWTPath      = "/data/jwt"
	rbacRulePath           = "/data/rbac"

	cipherKeyFile = "/user/orchestration_userID.txt"
	deviceIDFile  = "/device/orchestration_deviceID.txt"
)

var (
	orcheEngine         orchestrationapi.Orche
	edgeDir             string
	logPath             string
	configPath          string
	dbPath              string
	certificateFilePath string
	cipherKeyFilePath   string
	deviceIDFilePath    string
	log                 = logmgr.GetInstance()
)

func initPlatformPath(edgeDir string) {
	logPath = edgeDir + logStr
	configPath = edgeDir + configStr
	dbPath = edgeDir + dbStr
	certificateFilePath = edgeDir + certificateFile

	cipherKeyFilePath = edgeDir + cipherKeyFile
	deviceIDFilePath = edgeDir + deviceIDFile
}

type RequestServiceInfo struct {
	ExecutionType string
	ExeCmd        []string
}

type ReqeustService struct {
	ServiceName      string
	SelfSelection    bool
	ServiceRequester string
	ServiceInfo      []RequestServiceInfo
}

func (r *ReqeustService) SetExecutionCommand(execType string, command string) {
	switch execType {
	case "native", "android", "container":
	default:
		log.Printf("[%s] Invalid execution type: %s", logPrefix, execType)
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
	RemoteTargetInfo *TargetInfo
}

func (r ResponseService) GetExecutedType() string {
	return r.RemoteTargetInfo.ExecutionType
}

func (r ResponseService) GetTarget() string {
	return r.RemoteTargetInfo.Target
}

// ExecuteCallback is required to launch application in java layer
type ExecuteCallback interface {
	androidexecutor.ExecuteCallback
}

// OrchestrationInit runs orchestration service and discovers remote orchestration services
func OrchestrationInit(executeCallback ExecuteCallback, edgeDir string, isSecured bool) (errCode int) {
	initPlatformPath(edgeDir)

	logmgr.InitLogfile(logPath)
	log.Printf("[%s] OrchestrationInit", logPrefix)

	wrapper.SetBoltDBPath(dbPath)

	if isSecured {
		verifier.Init(containerWhiteListPath)
		authenticator.Init(passPhraseJWTPath)
		authorizer.Init(rbacRulePath)

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
	builder.SetExecutor(androidexecutor.GetInstance())
	builder.SetClient(restIns)

	orcheEngine = builder.Build()
	if orcheEngine == nil {
		log.Fatalf("[%s] Orchestration initialize fail", logPrefix)
		return
	}

	// set the android executor callback
	androidexecutor.GetInstance().SetExecuteCallback(executeCallback)

	orcheEngine.Start(deviceIDFilePath, platform, executionType)

	var restEdgeRouter *route.RestRouter
	if isSecured {
		restEdgeRouter = route.NewRestRouterWithCerti(certificateFilePath)
	} else {
		restEdgeRouter = route.NewRestRouter()
	}

	internalapi, err := orchestrationapi.GetInternalAPI()
	if err != nil {
		log.Fatalf("[%s] Orchestration internal api : %s", logPrefix, err.Error())
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

	restEdgeRouter.Start()

	log.Println(logPrefix, "Orchestration init done")

	errCode = 0

	return
}

// OrchestrationRequestService performs request from service applications which uses orchestration service
func OrchestrationRequestService(request *ReqeustService) *ResponseService {
	log.Printf("[%s] OrchestrationRequestService", logPrefix)
	log.Println("Service name: ", request.ServiceName)

	externalAPI, err := orchestrationapi.GetExternalAPI()
	if err != nil {
		log.Fatalf("[%s] Orchestaration external api : %s", logPrefix, err.Error())
	}

	changed := orchestrationapi.ReqeustService{
		ServiceName:      request.ServiceName,
		SelfSelection:    request.SelfSelection,
		ServiceRequester: request.ServiceRequester,
	}

	changed.ServiceInfo = make([]orchestrationapi.RequestServiceInfo, len(request.ServiceInfo))
	for idx, info := range request.ServiceInfo {
		changed.ServiceInfo[idx].ExecutionType = info.ExecutionType
		changed.ServiceInfo[idx].ExeCmd = info.ExeCmd
	}

	response := externalAPI.RequestService(changed)
	log.Println("Response : ", response)

	ret := &ResponseService{
		Message:     response.Message,
		ServiceName: response.ServiceName,
		RemoteTargetInfo: &TargetInfo{
			ExecutionType: response.RemoteTargetInfo.ExecutionType,
			Target:        response.RemoteTargetInfo.Target,
		},
	}
	return ret
}

//RegisterToBroadcastServer registers to the discovery server
func RegisterToBroadcastServer() int {
	log.Println(logPrefix, "Initiating Registration to Broadcast server")
	for {
		if discoverymgr.GetInstance() != nil {
			err := discoverymgr.GetInstance().NotifyMNEDCBroadcastServer()
			if err != nil {
				log.Println(logPrefix, "Registering to Broadcast server Error", err.Error(), ", retrying")
				time.Sleep(2 * time.Second)
				continue
			}
			return 0
		}
		break
	}
	return 1
}

//EncryptToByteAndPost encryps json data to byte array
func EncryptToByteAndPost(data string, target string) int {
	splitted := strings.Split(data, ",")
	jsonMap := make(map[string]interface{})

	if len(splitted) < 3 {
		log.Println(logPrefix, "Improper request data")
		return 1
	}
	jsonMap["VirtualAddr"] = splitted[0]
	jsonMap["PrivateAddr"] = splitted[1]
	jsonMap["DeviceID"] = splitted[2]

	cipher := sha256.GetCipher(cipherKeyFilePath)
	jsonStr, err := cipher.EncryptJSONToByte(jsonMap)
	if err != nil {
		log.Println(logPrefix, "Error in encrypting jsonMap:", err.Error())
		return 1
	}

	restapi := "/api/v1/discoverymgr/register"
	url := fmt.Sprintf("http://%s%s", target+":56002", restapi)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println(logPrefix, "Failed to create a new request:", err.Error())
		return 1
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error in Post", err.Error())
		return 1
	}
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}

	return 0
}

//MNEDCConnectionClosed notifies discovery manager that MNEDC connection is closed
func MNEDCConnectionClosed() {
	if discoverymgr.GetInstance() != nil {
		discoverymgr.GetInstance().MNEDCClosedCallback()
		return
	}
	log.Println(logPrefix, "discoverymgr instance is nil")
}

//MNEDCConnectionReEstablished notifies discovery manager that MNEDC connection is re-established
func MNEDCConnectionReEstablished() {
	if discoverymgr.GetInstance() != nil {
		discoverymgr.GetInstance().MNEDCReconciledCallback()
		return
	}
	log.Println(logPrefix, "discoverymgr instance is nil")
}

//GetPrivateIP returns private IP of the device
func GetPrivateIP() string {
	networkIns := networkhelper.GetInstance()
	if networkIns != nil {
		privateIP, err := networkIns.GetOutboundIP()
		if err != nil {
			return ""
		}
		return privateIP
	}
	return ""

}

type PSKHandler interface {
	tls.PSKHandler
}

func OrchestrationSetPSKHandler(pskHandler PSKHandler) {
	tls.SetPSKHandler(pskHandler)
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
