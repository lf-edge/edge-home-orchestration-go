/*******************************************************************************
 * Copyright 2021 Samsung Electronics All Rights Reserved.
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

// Package cloudsyncmgr provides functionalities to handle the cloud synchronization
package cloudsyncmgr

import (
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	mqttmgr "github.com/lf-edge/edge-home-orchestration-go/internal/common/mqtt"
)

const (
	logPrefix = "[cloudsyncmgr]"
)

// CloudSync is the interface for starting Cloud synchronization
type CloudSync interface {
	StartCloudSync(host string) error
}

//CloudSyncImpl struct
type CloudSyncImpl struct{}

var (
	cloudsyncIns *CloudSyncImpl
	log          = logmgr.GetInstance()
)

func init() {
	cloudsyncIns = &CloudSyncImpl{}

}

// GetInstance returns cloudSync instaance
func GetInstance() CloudSync {
	return cloudsyncIns
}

// StartCloudSync starts a server in terms of CloudSync
func (c *CloudSyncImpl) StartCloudSync(host string) (err error) {
	if len(host) > 0 {
		log.Info("Starting the CLoudsync Mgr")
		go mqttmgr.StartMQTTClient(host)
	}

	return nil
}
