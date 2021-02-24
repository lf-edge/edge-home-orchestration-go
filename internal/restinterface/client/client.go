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

// Package client implements REST client management functions
package client

import (
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/cipher"
)

// Clienter interface
type Clienter interface {
	cipher.Setter

	// for servicemgr
	DoExecuteRemoteDevice(appInfo map[string]interface{}, target string) (err error)
	DoNotifyAppStatusRemoteDevice(statusNotificationInfo map[string]interface{}, appID uint64, target string) (err error)

	// for scoringmgr
	DoGetScoreRemoteDevice(devID string, endpoint string) (scoreValue float64, err error)
	DoGetResourceRemoteDevice(devID string, endpoint string) (respMsg map[string]interface{}, err error)
	// for discoverymgr
	DoGetOrchestrationInfo(endpoint string) (platform string, executionType string, serviceList []string, err error)
	DoNotifyMNEDCBroadcastServer(endpoint string, port int, deviceID string, privateIP string, virtualIP string) error
}

// Setter interface
type Setter interface {
	SetClient(clientAPI Clienter)
}

// HasClient struct
type HasClient struct {
	Clienter Clienter
}

// SetClient sets function
func (c *HasClient) SetClient(api Clienter) {
	c.Clienter = api
}
