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

// Package orchestrationapi provides orchestration functionalities to handle distributed service in multi-device environment
package orchestrationapi

import (
	"controller/storagemgr/storagedriver"
	"errors"
	"log"
	"time"

	"common/commandvalidator"
	"common/networkhelper"
	"common/requestervalidator"
	"common/resourceutil"
	"common/types/configuremgrtypes"
	"controller/configuremgr"
	"controller/discoverymgr"
	"controller/scoringmgr"
	"controller/securemgr/verifier"
	"controller/servicemgr"
	"controller/servicemgr/executor"
	"controller/servicemgr/notification"
	"restinterface/client"
)

const (
	logtag              = "Orchestration"
	dataStorageService  = "datastorage"
	edgeDir             = "/var/edge-orchestration"
	dataStorageFilePath = edgeDir + "/datastorage/configuration.toml"
)

// Orche is the interface implemented by orchestration start function
type Orche interface {
	Start(deviceIDPath string, platform string, executionType string)
}

// OrcheExternalAPI is the interface implemented by external REST API
type OrcheExternalAPI interface {
	RequestService(serviceInfo ReqeustService) ResponseService
	verifier.VerifierConf
}

// OrcheInternalAPI is the interface implemented by internal REST API
type OrcheInternalAPI interface {
	configuremgr.Notifier
	ExecuteAppOnLocal(appInfo map[string]interface{})
	HandleNotificationOnLocal(serviceID float64, status string) error
	GetScore(target string) (scoreValue float64, err error)
	GetOrchestrationInfo() (platfrom string, executionType string, serviceList []string, err error)
	HandleDeviceInfo(deviceID string, virtualAddr string, privateAddr string)
	GetScoreWithResource(target map[string]interface{}) (scoreValue float64, err error)
	GetResource(target string) (resourceMsg map[string]interface{}, err error)
	GetStorageInstance() (storageInstance storagedriver.StorageDriver)
}

var (
	orcheIns            *orcheImpl
	resourceMonitorImpl resourceutil.Monitor
)

func init() {
	orcheIns = new(orcheImpl)
	orcheIns.networkhelper = networkhelper.GetInstance()
}

// GetExternalAPI registers the orchestration external API
func GetExternalAPI() (OrcheExternalAPI, error) {
	if orcheIns.Ready == false {
		return orcheIns, errors.New("orchestration engine does not ready")
	}
	return orcheIns, nil
}

// GetInternalAPI registers the orchestration internal API
func GetInternalAPI() (OrcheInternalAPI, error) {
	if orcheIns.Ready == false {
		return orcheIns, errors.New("orchestration engine does not ready")
	}
	return orcheIns, nil
}

func getOrcheImple() *orcheImpl {
	return orcheIns
}

// OrchestrationBuilder has every interface to run orchestration
type OrchestrationBuilder struct {
	isSetScoring bool
	scoringIns   scoringmgr.Scoring

	isSetVerifierConf bool
	verifierIns       verifier.VerifierConf

	isSetDiscovery bool
	discoveryIns   discoverymgr.Discovery

	isSetWatcher bool
	watcherIns   configuremgr.Watcher

	isSetService bool
	serviceIns   servicemgr.ServiceMgr

	isSetExecutor bool
	executorIns   executor.ServiceExecutor

	isSetClient bool
	clientAPI   client.Clienter

	isSetStorageDriver bool
	storageDriverIns   storagedriver.StorageDriver
}

// SetVerifierConf registers the interface to setting up verifier configuration
func (o *OrchestrationBuilder) SetVerifierConf(d verifier.VerifierConf) {
	o.isSetVerifierConf = true
	o.verifierIns = d
}

// SetScoring registers the interface to handle resource scoring
func (o *OrchestrationBuilder) SetScoring(s scoringmgr.Scoring) {
	o.isSetScoring = true
	o.scoringIns = s
}

// SetDiscovery registers the interface to handle orchestration discovery
func (o *OrchestrationBuilder) SetDiscovery(d discoverymgr.Discovery) {
	o.isSetDiscovery = true
	o.discoveryIns = d
}

// SetWatcher registers the interface to check if service applications are installed
func (o *OrchestrationBuilder) SetWatcher(w configuremgr.Watcher) {
	o.isSetWatcher = true
	o.watcherIns = w
}

// SetService registers the interface to handle executed service applications
func (o *OrchestrationBuilder) SetService(s servicemgr.ServiceMgr) {
	o.isSetService = true
	o.serviceIns = s
}

// SetExecutor registers the interface to execute platform-specific service application
func (o *OrchestrationBuilder) SetExecutor(e executor.ServiceExecutor) {
	o.isSetExecutor = true
	o.executorIns = e
}

// SetClient registers the interface to send request to remote device
func (o *OrchestrationBuilder) SetClient(c client.Clienter) {
	o.isSetClient = true
	o.clientAPI = c
}

// SetDataStorage registers the interface for DataStorage
func (o *OrchestrationBuilder) SetDataStorage(ds storagedriver.StorageDriver) {
	o.isSetStorageDriver = true
	o.storageDriverIns = ds
}

// Build registers every interface to run orchestration
func (o OrchestrationBuilder) Build() Orche {
	if !o.isSetWatcher || !o.isSetDiscovery || !o.isSetScoring ||
		!o.isSetService || !o.isSetExecutor || !o.isSetClient ||
		!o.isSetVerifierConf || !o.isSetStorageDriver {
		return nil
	}

	orcheIns.Ready = false
	orcheIns.scoringIns = o.scoringIns
	orcheIns.discoverIns = o.discoveryIns
	orcheIns.verifierIns = o.verifierIns
	orcheIns.watcher = o.watcherIns
	orcheIns.serviceIns = o.serviceIns
	orcheIns.clientAPI = o.clientAPI
	resourceMonitorImpl = resourceutil.GetMonitoringInstance()

	orcheIns.notificationIns = notification.GetInstance()
	orcheIns.serviceIns.SetLocalServiceExecutor(o.executorIns)

	orcheIns.discoverIns.SetRestResource()
	orcheIns.storageDriverIns = o.storageDriverIns

	return orcheIns
}

// Start runs the orchestration service itself
func (o *orcheImpl) Start(deviceIDPath string, platform string, executionType string) {
	resourceMonitorImpl.StartMonitoringResource()
	o.discoverIns.StartDiscovery(deviceIDPath, platform, executionType)
	o.watcher.Watch(o)
	o.Ready = true
	time.Sleep(1000)
}

func (o orcheImpl) Notify(serviceInfo configuremgrtypes.ServiceInfo) {
	validator := commandvalidator.CommandValidator{}
	if err := validator.AddWhiteCommand(serviceInfo); err != nil {
		log.Println(logtag, "[Error]", err.Error())
		return
	}
	requestervalidator.RequesterValidator{}.
		StoreRequesterInfo(serviceInfo.ServiceName, serviceInfo.AllowedRequester)

	if err := o.discoverIns.AddNewServiceName(serviceInfo.ServiceName); err != nil {
		log.Println(logtag, "[Error]", err.Error())
		return
	}
}

// ExecuteAppOnLocal executes a service application on local device
func (o orcheImpl) ExecuteAppOnLocal(appInfo map[string]interface{}) {
	o.serviceIns.ExecuteAppOnLocal(appInfo)
}

// HandleNotificationOnLocal handles notifications from local device after executing service application
func (o orcheImpl) HandleNotificationOnLocal(serviceID float64, status string) error {
	return o.notificationIns.HandleNotificationOnLocal(serviceID, status)
}

// GetScore gets a resource score of local device for specific app
func (o orcheImpl) GetScore(devID string) (scoreValue float64, err error) {
	return o.scoringIns.GetScore(devID)
}

// GetScoreWithResource gets a resource score of local device for specific app
func (o orcheImpl) GetScoreWithResource(resource map[string]interface{}) (scoreValue float64, err error) {
	return o.scoringIns.GetScoreWithResource(resource)
}

// GetResource gets resource values of local device for running apps
func (o orcheImpl) GetResource(devID string) (resourceMsg map[string]interface{}, err error) {
	return o.scoringIns.GetResource(devID)
}

// GetStorageInstance gets resource values of local device for running apps
func (o orcheImpl) GetStorageInstance() (storageInstance storagedriver.StorageDriver) {
	return o.storageDriverIns
}

// RequestVerifierConf setting up configuration of white list containers
func (o orcheImpl) RequestVerifierConf(containerInfo verifier.RequestVerifierConf) verifier.ResponseVerifierConf {
	return o.verifierIns.RequestVerifierConf(containerInfo)
}

//GetOrchestrationInfo gets orchestration info of the device
func (o orcheImpl) GetOrchestrationInfo() (platform string, executionType string, serviceList []string, err error) {
	return o.discoverIns.GetOrchestrationInfo()
}

//HandleDeviceInfo gets the peer's public and private Ip from relay server
func (o orcheImpl) HandleDeviceInfo(deviceID string, virtualAddr string, privateAddr string) {
	o.discoverIns.AddDeviceInfo(deviceID, virtualAddr, privateAddr)
}
