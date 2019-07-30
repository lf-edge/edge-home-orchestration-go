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

// Package orchestrationapi provides orchestration functionalities to handle distributed service in multi-device enviroment
package orchestrationapi

import (
	"errors"
	"log"
	"time"

	"common/networkhelper"
	"common/types/configuremgrtypes"
	"controller/configuremgr"
	"controller/discoverymgr"
	"controller/scoringmgr"
	"controller/servicemgr"
	"controller/servicemgr/executor"
	"controller/servicemgr/notification"
	"restinterface/client"
)

const logtag = "Orchestration"

// Orche is the interface implemented by orchestration start funciton
type Orche interface {
	Start(deviceIDPath string, platform string, executionType string)
}

// OrcheExternalAPI is the interface implemented by external REST API
type OrcheExternalAPI interface {
	RequestService(appName string, args []string) (handle int)
}

// OrcheInternalAPI is the interface implemented by internal REST API
type OrcheInternalAPI interface {
	configuremgr.Notifier
	ExecuteAppOnLocal(appInfo map[string]interface{})
	HandleNotificationOnLocal(serviceID float64, status string) error
	GetScore(target string, name string) (scoreValue float64, err error)
}

var orcheIns *orcheImpl

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

// Build registrers every interface to run orchestration
func (o OrchestrationBuilder) Build() Orche {
	if !o.isSetWatcher || !o.isSetDiscovery || !o.isSetScoring ||
		!o.isSetService || !o.isSetExecutor || !o.isSetClient {
		return nil
	}

	orcheIns.Ready = false
	orcheIns.scoringIns = o.scoringIns
	orcheIns.discoverIns = o.discoveryIns
	orcheIns.watcher = o.watcherIns
	orcheIns.serviceIns = o.serviceIns
	orcheIns.clientAPI = o.clientAPI

	orcheIns.notificationIns = notification.GetInstance()
	orcheIns.serviceIns.SetLocalServiceExecutor(o.executorIns)

	return orcheIns
}

// Start runs the orchestration service itself
func (o *orcheImpl) Start(deviceIDPath string, platform string, executionType string) {
	log.Println(logtag, "Start", "In")
	defer log.Println(logtag, "Start", "Out")

	o.discoverIns.StartDiscovery(deviceIDPath, platform, executionType)
	o.watcher.Watch(o)
	o.Ready = true
	time.Sleep(1000)
}

// Notify gives the notifications to scoringmgr and discoverymgr package after checking installed service applications
func (o orcheImpl) Notify(service configuremgrtypes.ServiceInfo) {
	if err := o.scoringIns.AddScoring(service); err != nil {
		log.Println(logtag, "[Error]", err.Error())
		return
	}
	if err := o.discoverIns.AddNewServiceName(service.ServiceName); err != nil {
		o.scoringIns.RemoveScoring(service.ServiceName)
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
func (o orcheImpl) GetScore(devID string, name string) (scoreValue float64, err error) {
	return o.scoringIns.GetScore(devID, name)
}
