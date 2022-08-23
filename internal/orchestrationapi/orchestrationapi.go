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

package orchestrationapi

import (
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/commandvalidator"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/networkhelper"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/requestervalidator"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/cloudsyncmgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/configuremgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/discoverymgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/scoringmgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/securemgr/verifier"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/servicemgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/servicemgr/notification"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/storagemgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/common"
	sysDB "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/system"
	dbhelper "github.com/lf-edge/edge-home-orchestration-go/internal/db/helper"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/client"
)

type orcheImpl struct {
	Ready bool

	verifierIns     verifier.Conf
	serviceIns      servicemgr.ServiceMgr
	scoringIns      scoringmgr.Scoring
	discoverIns     discoverymgr.Discovery
	watcher         configuremgr.Watcher
	notificationIns notification.Notification
	storageIns      storagemgr.Storage
	cloudsyncIns    cloudsyncmgr.CloudSync
	networkhelper   networkhelper.Network
	clientAPI       client.Clienter
}

type deviceInfo struct {
	id       string
	endpoint string
	score    float64
	resource map[string]interface{}
	execType string
}

type orcheClient struct {
	appName   string
	args      []string
	notiChan  chan string
	endSignal chan bool
}

// RequestServiceInfo struct
type RequestServiceInfo struct {
	ExecutionType string
	ExeCmd        []string
	ExeOption     map[string]interface{}
}

// ReqeustService struct
type ReqeustService struct {
	SelfSelection    bool
	ServiceName      string
	ServiceRequester string
	ServiceInfo      []RequestServiceInfo
	// TODO add status callback
}

// TargetInfo struct
type TargetInfo struct {
	ExecutionType string
	Target        string
}

// ResponseService struct
type ResponseService struct {
	Message          string
	ServiceName      string
	RemoteTargetInfo TargetInfo
}

const (
	// ErrorNone is key no error
	ErrorNone = "ERROR_NONE"
	// InvalidParameter is key for invalid parameter
	InvalidParameter = "INVALID_PARAMETER"
	//ServiceNotFound is key for service not found
	ServiceNotFound = "SERVICE_NOT_FOUND"
	//InternalServerError is key for internal server error
	InternalServerError = "INTERNAL_SERVER_ERROR"
	// NotAllowedCommand is key for not allowed command
	NotAllowedCommand  = "NOT_ALLOWED_COMMAND"
	cloudsyncLogPrefix = "[RequestCloudSync]"
)

var (
	orchClientID int32 = -1
	orcheClients       = [1024]orcheClient{}

	sysDBExecutor sysDB.DBInterface

	helper dbhelper.MultipleBucketQuery
)

func init() {
	sysDBExecutor = sysDB.Query{}

	helper = dbhelper.GetInstance()
}

// RequestCloudSyncPublish handles the request for cloud syncing
func (orcheEngine *orcheImpl) RequestCloudSyncPublish(host string, appID string, message string, topic string) string {
	log.Info(cloudsyncLogPrefix, "Requesting cloud sync publish")
	return orcheEngine.cloudsyncIns.RequestPublish(host, appID, message, topic)
}

// RequestCloudSyncSubscribe handles the request for cloud subscribing
func (orcheEngine *orcheImpl) RequestCloudSyncSubscribe(host string, appID string, topic string) string {
	log.Info(cloudsyncLogPrefix, "Requesting cloud sync subscribe")
	return orcheEngine.cloudsyncIns.RequestSubscribe(host, appID, topic)
}

// RequestSubscribedData request for the data on subscribed topic
func (orcheEngine *orcheImpl) RequestSubscribedData(appID string, topic string, host string) string {
	log.Info(cloudsyncLogPrefix, "Requesting SubscribedData")
	return orcheEngine.cloudsyncIns.RequestSubscribedData(appID, topic, host)
}

// RequestService handles service request (ex. offloading) from service application
func (orcheEngine *orcheImpl) RequestService(serviceInfo ReqeustService) ResponseService {
	log.Printf("[RequestService] %s: %v\n", logmgr.SanitizeUserInput(serviceInfo.ServiceName), serviceInfo.ServiceInfo) // lgtm [go/log-injection]

	if !orcheEngine.Ready {
		return ResponseService{
			Message:          InternalServerError,
			ServiceName:      serviceInfo.ServiceName,
			RemoteTargetInfo: TargetInfo{},
		}
	}

	atomic.AddInt32(&orchClientID, 1)

	handle := int(orchClientID)

	serviceClient := addServiceClient(handle, serviceInfo.ServiceName)
	go serviceClient.listenNotify()

	executionTypes := make([]string, 0)
	var scoringType string
	installed := true
	for _, info := range serviceInfo.ServiceInfo {
		executionTypes = append(executionTypes, info.ExecutionType)
		scoringType, _ = info.ExeOption["scoringType"].(string)
		if len(info.ExeCmd) > 0 {
			installed = false
		}
	}

	candidates, err := orcheEngine.getCandidate(serviceInfo.ServiceName, executionTypes, installed)

	log.Printf("[RequestService] getCandidate")
	for index, candidate := range candidates {
		log.Printf("[%d] ID       : %v", index, candidate.ID)
		log.Printf("[%d] ExecType : %v", index, candidate.ExecType)
		log.Printf("[%d] Endpoint : %v", index, candidate.Endpoint)
		log.Printf("")
	}

	if err != nil {
		return ResponseService{
			Message:          err.Error(),
			ServiceName:      serviceInfo.ServiceName,
			RemoteTargetInfo: TargetInfo{},
		}
	}

	errorResp := ResponseService{
		Message:          ServiceNotFound,
		ServiceName:      serviceInfo.ServiceName,
		RemoteTargetInfo: TargetInfo{},
	}

	var deviceScores []deviceInfo

	if scoringType == "resource" {
		deviceResources := orcheEngine.gatherDevicesResource(candidates, serviceInfo.SelfSelection)
		if len(deviceResources) <= 0 {
			return errorResp
		}
		for i, dev := range deviceResources {
			deviceResources[i].score, _ = orcheEngine.GetScoreWithResource(dev.resource)
		}
		deviceScores = sortByScore(deviceResources)
	} else {
		deviceScores = sortByScore(orcheEngine.gatherDevicesScore(candidates, serviceInfo.SelfSelection))
	}

	if len(deviceScores) <= 0 {
		return errorResp
	} else if deviceScores[0].score == scoringmgr.InvalidScore {
		return errorResp
	}

	args, err := getExecCmds(deviceScores[0].execType, serviceInfo.ServiceInfo)
	if err != nil {
		log.Println(err.Error())
		errorResp.Message = err.Error()
		return errorResp
	}
	args = append(args, deviceScores[0].execType)

	localhosts, err := orcheEngine.networkhelper.GetIPs()
	if err != nil {
		log.Println("[orchestrationapi] localhost ip gettering fail. maybe skipped localhost")
	}

	if common.HasElem(localhosts, deviceScores[0].endpoint) {
		validator := commandvalidator.CommandValidator{}
		for _, info := range serviceInfo.ServiceInfo {
			if info.ExecutionType == "native" || info.ExecutionType == "android" {
				if err := validator.CheckCommand(serviceInfo.ServiceName, info.ExeCmd); err != nil {
					log.Println(err.Error())
					return ResponseService{
						Message:          err.Error(),
						ServiceName:      serviceInfo.ServiceName,
						RemoteTargetInfo: TargetInfo{},
					}
				}
			}
		}

		vRequester := requestervalidator.RequesterValidator{}
		if err := vRequester.CheckRequester(serviceInfo.ServiceName, serviceInfo.ServiceRequester); err != nil &&
			(deviceScores[0].execType == "native" || deviceScores[0].execType == "android") {
			log.Println(err.Error())
			return ResponseService{
				Message:          err.Error(),
				ServiceName:      serviceInfo.ServiceName,
				RemoteTargetInfo: TargetInfo{},
			}
		}
	}

	orcheEngine.executeApp(
		deviceScores[0].endpoint,
		serviceInfo.ServiceName,
		serviceInfo.ServiceRequester,
		args,
		serviceClient.notiChan,
	)
	log.Println("[orchestrationapi] ", deviceScores)

	return ResponseService{
		Message:     ErrorNone,
		ServiceName: serviceInfo.ServiceName,
		RemoteTargetInfo: TargetInfo{
			ExecutionType: deviceScores[0].execType,
			Target:        deviceScores[0].endpoint,
		},
	}
}

func getExecCmds(execType string, requestServiceInfos []RequestServiceInfo) ([]string, error) {
	for _, requestServiceInfo := range requestServiceInfos {
		if execType == requestServiceInfo.ExecutionType {
			return requestServiceInfo.ExeCmd, nil
		}
	}

	return nil, errors.New("not found")
}

func (orcheEngine orcheImpl) getCandidate(appName string, execType []string, installed bool) (deviceList []dbhelper.ExecutionCandidate, err error) {
	return helper.GetDeviceInfoWithService(appName, execType, installed)
}

func (orcheEngine orcheImpl) gatherDevicesScore(candidates []dbhelper.ExecutionCandidate, selfSelection bool) (deviceScores []deviceInfo) {
	count := len(candidates)
	if !selfSelection {
		count--
	}
	scores := make(chan deviceInfo, count)

	info, err := sysDBExecutor.Get(sysDB.ID)
	if err != nil {
		log.Println("[orchestrationapi] localhost devid gettering fail")
		return
	}

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(3 * time.Second)
		timeout <- true
	}()

	var wait sync.WaitGroup
	wait.Add(1)
	index := 0
	go func() {
		defer wait.Done()
		for {
			select {
			case score := <-scores:
				deviceScores = append(deviceScores, score)
				if index++; count == index {
					return
				}
			case <-timeout:
				return
			}
		}
	}()

	localhosts, err := orcheEngine.networkhelper.GetIPs()
	if err != nil {
		log.Println("[orchestrationapi] localhost ip gettering fail. maybe skipped localhost")
	}

	for _, candidate := range candidates {
		go func(cand dbhelper.ExecutionCandidate) {
			var score float64
			var err error

			if len(cand.Endpoint) == 0 {
				log.Println("[orchestrationapi] cannot getting score, cause by ip list is empty")
				scores <- deviceInfo{endpoint: "", score: float64(0.0), id: cand.ID}
				return
			}

			if isLocalhost(cand.Endpoint, localhosts) {
				if !selfSelection {
					return
				}
				score, err = orcheEngine.GetScore(info.Value)
			} else {
				score, err = orcheEngine.clientAPI.DoGetScoreRemoteDevice(info.Value, cand.Endpoint[0])
			}

			if err != nil {
				log.Println("[orchestrationapi] cannot getting score from :", cand.Endpoint[0], "cause by", err.Error())
				scores <- deviceInfo{endpoint: cand.Endpoint[0], score: float64(0.0), id: cand.ID}
				return
			}
			log.Printf("[orchestrationapi] deviceScore")
			log.Printf("candidate ID       : %v", cand.ID)
			log.Printf("candidate ExecType : %v", cand.ExecType)
			log.Printf("candidate Endpoint : %v", cand.Endpoint[0])
			log.Printf("candidate score    : %v", score)
			scores <- deviceInfo{endpoint: cand.Endpoint[0], score: score, id: cand.ID, execType: cand.ExecType}
		}(candidate)
	}

	wait.Wait()

	return
}

// gatherDevicesResource gathers resource values from edge devices
func (orcheEngine orcheImpl) gatherDevicesResource(candidates []dbhelper.ExecutionCandidate, selfSelection bool) (deviceResources []deviceInfo) {
	count := len(candidates)
	if !selfSelection {
		count--
	}
	resources := make(chan deviceInfo, count)

	info, err := sysDBExecutor.Get(sysDB.ID)
	if err != nil {
		log.Println("[orchestrationapi] localhost devid gettering fail")
		return
	}

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(3 * time.Second)
		timeout <- true
	}()

	var wait sync.WaitGroup
	wait.Add(1)
	index := 0
	go func() {
		defer wait.Done()
		for {
			select {
			case resource := <-resources:
				deviceResources = append(deviceResources, resource)
				if index++; count == index {
					return
				}
			case <-timeout:
				return
			}
		}
	}()

	localhosts, err := orcheEngine.networkhelper.GetIPs()
	if err != nil {
		log.Println("[orchestrationapi] localhost ip gettering fail. maybe skipped localhost")
	}

	for _, candidate := range candidates {
		go func(cand dbhelper.ExecutionCandidate) {
			var resource map[string]interface{}
			var err error

			if len(cand.Endpoint) == 0 {
				log.Println("[orchestrationapi] cannot getting score, cause by ip list is empty")
				resources <- deviceInfo{endpoint: "", resource: resource, id: cand.ID, execType: cand.ExecType}
				return
			}

			if isLocalhost(cand.Endpoint, localhosts) {
				if !selfSelection {
					return
				}
				resource, err = orcheEngine.GetResource(info.Value)
			} else {
				resource, err = orcheEngine.clientAPI.DoGetResourceRemoteDevice(info.Value, cand.Endpoint[0])
			}

			if err != nil {
				log.Println("[orchestrationapi] cannot getting msgs from :", cand.Endpoint[0], "cause by", err.Error())
				resources <- deviceInfo{endpoint: cand.Endpoint[0], resource: resource, id: cand.ID, execType: cand.ExecType}
				return
			}
			log.Printf("[orchestrationapi] deviceResource")
			log.Printf("candidate ID       : %v", cand.ID)
			log.Printf("candidate ExecType : %v", cand.ExecType)
			log.Printf("candidate Endpoint : %v", cand.Endpoint[0])
			log.Printf("candidate resource : %v", resource)
			resources <- deviceInfo{endpoint: cand.Endpoint[0], resource: resource, id: cand.ID, execType: cand.ExecType}
		}(candidate)
	}

	wait.Wait()

	return
}

func (orcheEngine orcheImpl) executeApp(endpoint, serviceName, requester string, args []string, notiChan chan string) {
	ifArgs := make([]interface{}, len(args))
	for i, v := range args {
		ifArgs[i] = v
	}

	orcheEngine.serviceIns.Execute(endpoint, serviceName, requester, ifArgs, notiChan)
}

func (client *orcheClient) listenNotify() {
	select {
	case str := <-client.notiChan:
		log.Printf("[orchestrationapi] service status changed [appNames:%s][status:%s]\n", client.appName, str)
	}
}

func isLocalhost(endpoints1, endpoints2 []string) bool {
	for _, endpoint1 := range endpoints1 {
		for _, endpoint2 := range endpoints2 {
			if endpoint1 == endpoint2 {
				return true
			}
		}
	}
	return false
}

func addServiceClient(clientID int, appName string) (client *orcheClient) {
	// orcheClients[clientID].args = args
	orcheClients[clientID].appName = appName
	orcheClients[clientID].notiChan = make(chan string)

	client = &orcheClients[clientID]
	return
}

func sortByScore(deviceScores []deviceInfo) []deviceInfo {
	sort.Slice(deviceScores, func(i, j int) bool {
		return deviceScores[i].score > deviceScores[j].score
	})

	return deviceScores
}
