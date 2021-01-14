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

// Package executor provides struct and interface for multi-platform execution
package executor

import (
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/notification"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/client"
)

// ServiceExecutor interface
type ServiceExecutor interface {
	Execute(ServiceExecutionInfo) (err error)
	SetNotiImpl(noti notification.Notification)
	client.Setter
}

// ServiceExecutionInfo has all information to execute service
type ServiceExecutionInfo struct {
	ServiceID             uint64
	ServiceName           string
	ParamStr              []string
	NotificationTargetURL string
}

// HasClientNotification struct
type HasClientNotification struct {
	notification.HasNotification
	client.HasClient
}

// SetClient sets ClientAPI
func (c *HasClientNotification) SetClient(clientAPI client.Clienter) {
	c.Clienter = clientAPI
	c.NotiImplIns.SetClient(clientAPI)
}
