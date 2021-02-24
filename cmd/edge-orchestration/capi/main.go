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

// Package main provides C interface for orchestration
package main

/*
#include <stdlib.h>

#ifndef __ORCHESTRATION_H__
#define __ORCHESTRATION_H__

#ifdef __cplusplus
extern "C"
{
#endif

#define MAX_SVC_INFO_NUM 3
typedef struct {
	char* ExecutionType;
	char* ExeCmd;
} RequestServiceInfo;

typedef struct {
	char* ExecutionType;
	char* Target;
} TargetInfo;

typedef struct {
	char*      Message;
	char*      ServiceName;
	TargetInfo RemoteTargetInfo;
} ResponseService;

typedef char* (*identityGetterFunc)();
typedef char* (*keyGetterFunc)(char* id);

identityGetterFunc iGetter;
keyGetterFunc kGetter;

static void setPSKHandler(identityGetterFunc ihandle, keyGetterFunc khandle){
	iGetter = ihandle;
	kGetter = khandle;
}

static char* bridge_iGetter(){
	return iGetter();
}

static char* bridge_kGetter(char* id){
	return kGetter(id);
}
#ifdef __cplusplus
}

#endif

#endif // __ORCHESTRATION_H__

*/
import "C"
import (
	"errors"
	"flag"
	"math"
	"strings"
	"sync"
	"unsafe"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"

	configuremgr "github.com/lf-edge/edge-home-orchestration-go/internal/controller/configuremgr/native"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/discoverymgr"
	mnedcmgr "github.com/lf-edge/edge-home-orchestration-go/internal/controller/discoverymgr/mnedc"
	scoringmgr "github.com/lf-edge/edge-home-orchestration-go/internal/controller/scoringmgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/securemgr/authenticator"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/securemgr/authorizer"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/securemgr/verifier"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/servicemgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/servicemgr/executor/nativeexecutor"

	"github.com/lf-edge/edge-home-orchestration-go/internal/orchestrationapi"

	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/cipher/dummy"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/cipher/sha256"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/client/restclient"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/externalhandler"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/internalhandler"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/route"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/tls"

	"github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/wrapper"
)

const logPrefix = "interface"

// Handle Platform Dependencies
const (
	platform      = "linux"
	executionType = "native"

	edgeDir = "/var/edge-orchestration"

	logPath                = edgeDir + "/log"
	configPath             = edgeDir + "/apps"
	dbPath                 = edgeDir + "/data/db"
	certificateFilePath    = edgeDir + "/data/cert"
	containerWhiteListPath = edgeDir + "/data/cwl"
	passPhraseJWTPath      = edgeDir + "/data/jwt"
	rbacRulePath           = edgeDir + "/data/rbac"

	cipherKeyFilePath = edgeDir + "/user/orchestration_userID.txt"
	deviceIDFilePath  = edgeDir + "/device/orchestration_deviceID.txt"
	mnedcServerConfig = edgeDir + "/mnedc/client.config"
)

var (
	flagVersion                  bool
	commitID, version, buildTime string
	buildTags                    string
	log                          = logmgr.GetInstance()

	orcheEngine orchestrationapi.Orche
)

//export OrchestrationInit
func OrchestrationInit() C.int {
	flag.BoolVar(&flagVersion, "v", false, "if true, print version and exit")
	flag.BoolVar(&flagVersion, "version", false, "if true, print version and exit")
	flag.Parse()

	logmgr.InitLogfile(logPath)
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
	builder.SetExecutor(nativeexecutor.GetInstance())
	builder.SetClient(restIns)
	orcheEngine = builder.Build()
	if orcheEngine == nil {
		log.Fatalf("[%s] Orchestaration initalize fail", logPrefix)
		return -1
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

	if isSecured {
		mnedcmgr.GetServerInstance().SetCipher(dummy.GetCipher(cipherKeyFilePath))
		mnedcmgr.GetServerInstance().SetCertificateFilePath(certificateFilePath)
		mnedcmgr.GetClientInstance().SetCertificateFilePath(certificateFilePath)
	} else {
		mnedcmgr.GetServerInstance().SetCipher(sha256.GetCipher(cipherKeyFilePath))
	}
	isMNEDCServer := false
	isMNEDCClient := false
	if strings.Contains(buildTags, "mnedcserver") {
		isMNEDCServer = true
	} else if strings.Contains(buildTags, "mnedcclient") {
		isMNEDCClient = true
	}

	go func() {
		if isMNEDCServer {
			mnedcmgr.GetServerInstance().StartMNEDCServer(deviceIDFilePath)
		} else if isMNEDCClient {
			mnedcmgr.GetClientInstance().StartMNEDCClient(deviceIDFilePath, mnedcServerConfig)
		}
	}()
	return 0
}

//export OrchestrationRequestService
func OrchestrationRequestService(cAppName *C.char, cSelfSelection C.int, cRequester *C.char, serviceInfo *C.RequestServiceInfo, count C.int) C.ResponseService {
	log.Printf("[%s] OrchestrationRequestService", logPrefix)

	appName := C.GoString(cAppName)

	requestInfos := make([]orchestrationapi.RequestServiceInfo, count)
	CServiceInfo := (*[(math.MaxInt16 - 1) / unsafe.Sizeof(serviceInfo)]C.RequestServiceInfo)(unsafe.Pointer(serviceInfo))[:count:count]

	for idx, requestInfo := range CServiceInfo {
		requestInfos[idx].ExecutionType = C.GoString(requestInfo.ExecutionType)

		args := strings.Split(C.GoString(requestInfo.ExeCmd), " ")
		if strings.Compare(args[0], "") == 0 {
			args = nil
		}
		requestInfos[idx].ExeCmd = append([]string{}, args...)
	}

	externalAPI, err := orchestrationapi.GetExternalAPI()
	if err != nil {
		log.Fatalf("[%s] Orchestaration external api : %s", logPrefix, err.Error())
	}

	selfSel := true
	if cSelfSelection == 0 {
		selfSel = false
	}

	requester := C.GoString(cRequester)

	log.Printf("[OrchestrationRequestService] appName:%s", appName)
	log.Printf("[OrchestrationRequestService] selfSel:%v", selfSel)
	log.Printf("[OrchestrationRequestService] requester:%s", requester)
	log.Printf("[OrchestrationRequestService] infos:%v", requestInfos)

	res := externalAPI.RequestService(orchestrationapi.ReqeustService{
		ServiceName:      appName,
		SelfSelection:    selfSel,
		ServiceInfo:      requestInfos,
		ServiceRequester: requester,
	})
	log.Println("requestService handle : ", res)

	ret := C.ResponseService{}
	ret.Message = C.CString(res.Message)
	ret.ServiceName = C.CString(res.ServiceName)
	ret.RemoteTargetInfo.ExecutionType = C.CString(res.RemoteTargetInfo.ExecutionType)
	ret.RemoteTargetInfo.Target = C.CString(res.RemoteTargetInfo.Target)

	return ret
}

type customPSKHandler struct{}

func (cHandler customPSKHandler) GetIdentity() string {
	var cIdentity *C.char
	cIdentity = C.bridge_iGetter()
	identity := C.GoString(cIdentity)
	return identity
}

func (cHandler customPSKHandler) GetKey(id string) ([]byte, error) {
	var cKey *C.char
	cStr := C.CString(id)
	defer C.free(unsafe.Pointer(cStr))

	cKey = C.bridge_kGetter(cStr)
	key := C.GoString(cKey)
	if len(key) == 0 {
		return nil, errors.New("key is empty")
	}
	return []byte(key), nil
}

//export SetPSKHandler
func SetPSKHandler(iGetter C.identityGetterFunc, kGetter C.keyGetterFunc) {
	C.setPSKHandler(iGetter, kGetter)
	tls.SetPSKHandler(customPSKHandler{})
}

var count int
var mtx sync.Mutex

//export PrintLog
func PrintLog(cMsg *C.char) (count C.int) {
	mtx.Lock()
	msg := C.GoString(cMsg)
	defer mtx.Unlock()
	log.Printf(msg)
	count++
	return
}

func main() {
	// Do nothing because it is only to build static lWatcher interface
}
