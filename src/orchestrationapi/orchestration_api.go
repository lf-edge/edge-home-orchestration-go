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

package orchestrationapi

import (
	"log"
	"sort"
	"sync"
	"sync/atomic"

	"common/networkhelper"
	"controller/configuremgr"
	"controller/discoverymgr"
	"controller/scoringmgr"
	"controller/servicemgr"
	"controller/servicemgr/notification"
	"restinterface/client"
)

type orcheImpl struct {
	Ready bool

	serviceIns      servicemgr.ServiceMgr
	scoringIns      scoringmgr.Scoring
	discoverIns     discoverymgr.Discovery
	watcher         configuremgr.Watcher
	notificationIns notification.Notification

	networkhelper networkhelper.Network

	clientAPI client.Clienter
}

type deviceScore struct {
	endpoint string
	score    float64
}

type orcheClient struct {
	appName   string
	args      []string
	notiChan  chan string
	endSignal chan bool
}

type RequestServiceInfo struct {
	ExecutionType string
	ExeCmd        []string
}

type ReqeustService struct {
	ServiceName string
	ServiceInfo []RequestServiceInfo
	// TODO add status callback
}

type TargetInfo struct {
	ExecutionType string
	Target        string
}

type ReponseService struct {
	Message          string
	ServiceName      string
	RemoteTargetInfo TargetInfo
}

const (
	ERROR_NONE        = "ERROR_NONE"
	INVALID_PARAMETER = "INVALID_PARAMETER"
	SERVICE_NOT_FOUND = "SERVICE_NOT_FOUND"
)

var (
	orchClientID int32 = -1
	orcheClients       = [1024]orcheClient{}
)

// RequestService handles service reqeust (ex. offloading) from service application
func (orcheEngine *orcheImpl) RequestService(serviceInfo ReqeustService) ReponseService {
	log.Printf("[RequestService] %v: %v\n", serviceInfo.ServiceName, serviceInfo.ServiceInfo)
	if orcheEngine.Ready == false {
		return ReponseService{
			Message:          ERROR_NONE,
			ServiceName:      serviceInfo.ServiceName,
			RemoteTargetInfo: TargetInfo{},
		}
	}

	atomic.AddInt32(&orchClientID, 1)

	//	handle := int(orchClientID)

	//TODO
	//	serviceClient := addServiceClient(handle, appName, args)
	//	go serviceClient.listenNotify()

	return ReponseService{}
}

//func (orcheEngine *orcheImpl) RequestService(appName string, args []string) (handle int) {
//
//	log.Printf("[RequestService] %v: %v\n", appName, args)
//
//	if orcheEngine.Ready == false {
//		return errormsg.ErrorNotReadyOrchestrationInit
//	}
//
//	atomic.AddInt32(&orchClientID, 1)
//
//	handle = int(orchClientID)
//
//	serviceClient := addServiceClient(handle, appName, args)
//	go serviceClient.listenNotify()
//	endpoints, err := orcheEngine.getEndpointDevices(appName)
//	if err != nil {
//		return errormsg.ToInt(err)
//	}
//	deviceScores := sortByScore(orcheEngine.gatheringDevicesScore(endpoints, appName))
//
//	if len(deviceScores) > 0 {
//		orcheEngine.executeApp(deviceScores[0].endpoint, appName, args, serviceClient.notiChan)
//		log.Println("[orchestrationapi] ", deviceScores)
//	}
//
//	return
//}

func (orcheEngine orcheImpl) getEndpointDevices(appName string) (deviceList []string, err error) {
	return orcheEngine.discoverIns.GetDeviceIPListWithService(appName)
}

func (orcheEngine orcheImpl) gatheringDevicesScore(endpoints []string, appName string) (deviceScores []deviceScore) {

	scores := make(chan deviceScore, len(endpoints))
	count := len(endpoints)
	index := 0

	var wait sync.WaitGroup
	wait.Add(1)

	go func() {
		for {
			score := <-scores
			deviceScores = append(deviceScores, score)
			if index++; count == index {
				break
			}
		}
		wait.Done()
		return
	}()

	localhost, err := orcheEngine.networkhelper.GetOutboundIP()
	if err != nil {
		log.Println("[orchestrationapi] ", "localhost ip gettering fail", "maybe skipped localhost")
	}

	for _, endpoint := range endpoints {
		go func(endpoint, appName string) {
			var score float64
			var err error

			if endpoint == localhost {
				score, err = orcheEngine.GetScore(endpoint, appName)
			} else {
				score, err = orcheEngine.clientAPI.DoGetScoreRemoteDevice(appName, endpoint)
			}

			if err != nil {
				log.Println("[orchestrationapi] ", "cannot getting score from : ", endpoint, " cause by ", err.Error())
				scores <- deviceScore{endpoint, float64(0.0)}
				return
			}
			scores <- deviceScore{endpoint, score}
		}(endpoint, appName)
	}

	wait.Wait()

	return
}

func (orcheEngine orcheImpl) executeApp(endpoint string, serviceName string, args []string, notiChan chan string) {
	ifArgs := make([]interface{}, len(args))
	for i, v := range args {
		ifArgs[i] = v
	}

	orcheEngine.serviceIns.Execute(endpoint, serviceName, ifArgs, notiChan)
}

func (client *orcheClient) listenNotify() {
	select {
	case str := <-client.notiChan:
		log.Printf("[orchestrationapi] service status changed [appNames:%s][status:%s]\n", client.appName, str)
	}
}

func addServiceClient(clientID int, appName string, args []string) (client *orcheClient) {
	orcheClients[clientID].args = args
	orcheClients[clientID].appName = appName
	orcheClients[clientID].notiChan = make(chan string)

	client = &orcheClients[clientID]
	return
}

func sortByScore(deviceScores []deviceScore) []deviceScore {
	sort.Slice(deviceScores, func(i, j int) bool {
		return deviceScores[i].score > deviceScores[j].score
	})

	return deviceScores
}
