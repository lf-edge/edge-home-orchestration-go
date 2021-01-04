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

// Package notification provides functions to manage notification after running service application
package notification

import (
	"errors"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
	"strings"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/networkhelper"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/client"
)

// Notification is the interface for notification
type Notification interface {
	InvokeNotification(target string, serviceID float64, status string) error
	AddNotificationChan(serviceID uint64, notiChan chan string)
	HandleNotificationOnLocal(serviceID float64, status string) (err error)

	// for client
	client.Setter
}

// HasNotification struct
type HasNotification struct {
	NotiImplIns Notification
}

// SetNotiImpl for setting notification implementation
func (n *HasNotification) SetNotiImpl(noti Notification) {
	n.NotiImplIns = noti
}

// NotiImpl Structure
type NotiImpl struct {
	client.HasClient
}

var (
	log              = logmgr.GetInstance()
	// notificaitonImpl is instance of NotiImpl
	notificationImpl = &NotiImpl{}

	// NotificationMap is map strucutre (serviceID / Notification Channel)
	notificationMap  = ConcurrentMap{items: make(map[uint64]interface{})}
)

// GetInstance returns the singleton NotiImpl instance
func GetInstance() *NotiImpl {
	return notificationImpl
}

// AddNotificationChan is adding notification channel value with service ID key
func (NotiImpl) AddNotificationChan(serviceID uint64, notiChan chan string) {
	value := make(map[string]interface{})

	value[ConstKeyNotiChan] = notiChan
	notificationMap.Set(serviceID, value)

}

// InvokeNotification is processing notification
func (n NotiImpl) InvokeNotification(target string, serviceID float64, status string) (err error) {
	outboundIP, outboundIPErr := networkhelper.GetInstance().GetOutboundIP()
	if outboundIPErr != nil {
		outboundIP = ""
	}

	if strings.Compare(target, outboundIP) == 0 {
		return n.HandleNotificationOnLocal(serviceID, status)
	}
	return n.handleNotificationOnRemote(target, serviceID, status)
}

// HandleNotificationOnLocal is invoking notification on local
func (NotiImpl) HandleNotificationOnLocal(serviceID float64, status string) (err error) {
	id := uint64(serviceID)
	notiChan, err := getNotiChan(id)
	if notiChan == nil {
		return
	}

	notiChan <- status
	notificationMap.Remove(id)

	return
}

func (n NotiImpl) handleNotificationOnRemote(target string, serviceID float64, status string) (err error) {
	statusNotificationInfo := make(map[string]interface{})
	statusNotificationInfo["ServiceID"] = serviceID
	statusNotificationInfo["Status"] = status

	err = n.Clienter.DoNotifyAppStatusRemoteDevice(statusNotificationInfo, uint64(serviceID), target)

	if err != nil {
		log.Println(logPrefix, err.Error())
	}
	return
}

func getNotiChan(serviceID uint64) (notiChan chan string, err error) {
	value, _ := notificationMap.Get(serviceID)
	if value == nil {
		return nil, errors.New("invalid serviceID")
	}

	valueList := value.(map[string]interface{})
	return valueList[ConstKeyNotiChan].(chan string), nil
}
