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
	"db/bolt/configuration"
	"db/bolt/network"
	"errors"
	"log"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"common/networkhelper"
	"controller/configuremgr"
	"controller/discoverymgr"
	"controller/scoringmgr"
	"controller/servicemgr"
	"controller/servicemgr/notification"
	"restinterface/client"

	sysDB "db/bolt/system"
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

	sysDBExecutor sysDB.DBInterface
)

func init() {
	sysDBExecutor = sysDB.Query{}
}

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

	handle := int(orchClientID)

	serviceClient := addServiceClient(handle, serviceInfo.ServiceName)
	go serviceClient.listenNotify()

	var executionTypes []string
	for _, info := range serviceInfo.ServiceInfo {
		executionTypes = append(executionTypes, info.ExecutionType)
	}

	endpoints, err := orcheEngine.getEndpointDevices(serviceInfo.ServiceName, executionTypes)
	if err != nil {
		return ReponseService{
			Message:          err.Error(),
			ServiceName:      serviceInfo.ServiceName,
			RemoteTargetInfo: TargetInfo{},
		}
	}

	deviceScores := sortByScore(orcheEngine.gatherDevicesScore(endpoints, serviceInfo.ServiceName))
	if len(deviceScores) > 0 {
		args, err := getExecCmds(deviceScores[0].endpoint, serviceInfo.ServiceInfo)
		if err != nil {
			log.Println(err.Error())
			return ReponseService{
				Message:          err.Error(),
				ServiceName:      serviceInfo.ServiceName,
				RemoteTargetInfo: TargetInfo{},
			}
		}

		orcheEngine.executeApp(deviceScores[0].endpoint, serviceInfo.ServiceName, args, serviceClient.notiChan)
		log.Println("[orchestrationapi] ", deviceScores)
	}

	return ReponseService{
		Message:          ERROR_NONE,
		ServiceName:      serviceInfo.ServiceName,
		RemoteTargetInfo: TargetInfo{},
	}
}

func getExecCmds(endpoint string, requestServiceInfo []RequestServiceInfo) ([]string, error) {
	netQuery := network.Query{}
	id, err := netQuery.GetIDWithIP(endpoint)
	if err != nil {
		return nil, err
	}

	confQuery := configuration.Query{}
	confInfo, err := confQuery.Get(id)
	if err != nil {
		return nil, err
	}

	for _, serviceInfo := range requestServiceInfo {
		if strings.Compare(serviceInfo.ExecutionType, confInfo.ExecType) == 0 {
			return serviceInfo.ExeCmd, nil
		}
	}

	return nil, errors.New("Not Found")
}

func (orcheEngine orcheImpl) getEndpointDevices(appName string, execType []string) (deviceList []string, err error) {
	return orcheEngine.discoverIns.GetDeviceIPListWithService(appName, execType)
}

func (orcheEngine orcheImpl) gatherDevicesScore(endpoints []string, appName string) (deviceScores []deviceScore) {

	scores := make(chan deviceScore, len(endpoints))
	count := len(endpoints)
	index := 0

	info, err := sysDBExecutor.Get(sysDB.ID)
	if err != nil {
		log.Println("[orchestrationapi] ", "localhost devid gettering fail")
		return
	}

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
				score, err = orcheEngine.GetScore(info.Value, appName)
			} else {
				score, err = orcheEngine.clientAPI.DoGetScoreRemoteDevice(info.Value, appName, endpoint)
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

func addServiceClient(clientID int, appName string) (client *orcheClient) {
	// orcheClients[clientID].args = args
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
