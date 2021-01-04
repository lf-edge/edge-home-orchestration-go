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

// Package servicemgr manages service application execution
package servicemgr

import (
	"strings"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/networkhelper"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/executor"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/notification"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/client"
)

// ServiceMgr is the interface to execute service application
type ServiceMgr interface {
	Execute(target, name, requester string, args []interface{}, notiChan chan string) (err error)
	SetLocalServiceExecutor(s executor.ServiceExecutor)

	// for internal api
	ExecuteAppOnLocal(appInfo map[string]interface{})

	// for client
	client.Setter
}

// SMMgrImpl Structure
type SMMgrImpl struct {
	serviceExecutor executor.ServiceExecutor
	client.HasClient
}

var (
	serviceMgr *SMMgrImpl
)

func init() {
	ServiceMap = ConcurrentMap{items: make(map[uint64]interface{})}
	serviceMgr = &SMMgrImpl{}

}

// GetInstance returns the singletone SMMgrImpl instance
func GetInstance() *SMMgrImpl {
	return serviceMgr
}

// SetLocalServiceExecutor sets executor and client
func (sm *SMMgrImpl) SetLocalServiceExecutor(s executor.ServiceExecutor) {
	s.SetClient(sm.Clienter)
	sm.serviceExecutor = s
}

// Execute selects local execution and remote execution
func (sm SMMgrImpl) Execute(target, name, requester string, args []interface{}, notiChan chan string) (err error) {
	serviceID := createServiceMap(name)
	appInfo := makeAppInfo(target, name, requester, args, float64(serviceID))

	notification.GetInstance().AddNotificationChan(serviceID, notiChan)

	outboundIP, outboundIPErr := networkhelper.GetInstance().GetOutboundIP()
	if outboundIPErr != nil {
		outboundIP = ""
	}

	if strings.Compare(target, outboundIP) == 0 {
		sm.ExecuteAppOnLocal(appInfo)
	} else {
		err = sm.executeAppOnRemote(target, appInfo)
	}

	return
}

// ExecuteAppOnLocal fills out service execution info and deliver it to executor
func (sm SMMgrImpl) ExecuteAppOnLocal(appInfo map[string]interface{}) {
	var serviceExecutionInfo executor.ServiceExecutionInfo

	serviceID, serviceName, args, notitargetURL := parseAppInfo(appInfo)
	args = args[:len(args)-1]

	serviceExecutionInfo = executor.ServiceExecutionInfo{
		ServiceID:             serviceID,
		ServiceName:           serviceName,
		ParamStr:              args,
		NotificationTargetURL: notitargetURL}

	go sm.serviceExecutor.Execute(serviceExecutionInfo)
}

func (sm SMMgrImpl) executeAppOnRemote(target string, appInfo map[string]interface{}) (err error) {
	err = sm.Clienter.DoExecuteRemoteDevice(appInfo, target)
	return
}

func makeAppInfo(target, name, requester string, args []interface{}, serviceID float64) (appInfo map[string]interface{}) {
	appInfo = make(map[string]interface{})

	appInfo[ConstKeyServiceID] = serviceID
	appInfo[ConstKeyServiceName] = name
	appInfo[ConstKeyNotiTargetURL] = target
	appInfo[ConstKeyRequester] = requester

	if args != nil {
		appInfo[ConstKeyUserArgs] = args
	}

	return
}

func parseAppInfo(appInfo map[string]interface{}) (serviceID uint64, serviceName string, args []string, notificationTargetURL string) {
	serviceID = uint64(appInfo[ConstKeyServiceID].(float64))
	serviceName = appInfo[ConstKeyServiceName].(string)
	notificationTargetURL = appInfo[ConstKeyNotiTargetURL].(string)

	userArgs := appInfo[ConstKeyUserArgs]
	if userArgs != nil {
		for _, param := range userArgs.([]interface{}) {
			args = append(args, param.(string))
		}
	} else {
		args = nil
	}

	return
}
