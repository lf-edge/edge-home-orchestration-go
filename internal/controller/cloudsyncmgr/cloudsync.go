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
	"fmt"
	"strings"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	mqttmgr "github.com/lf-edge/edge-home-orchestration-go/internal/common/mqtt"
)

const (
	logPrefix = "[cloudsyncmgr]"
)

// CloudSync is the interface for starting Cloud synchronization
type CloudSync interface {
	InitiateCloudSync(isCloudSet string) error
	StartCloudSync(host string) error
	//implemented by external REST API
	RequestCloudSyncConf(message mqttmgr.Message, topic string, clientID string) string
}

//CloudSyncImpl struct
type CloudSyncImpl struct{}

var (
	cloudsyncIns   *CloudSyncImpl
	log            = logmgr.GetInstance()
	mqttClient     *mqttmgr.Client
	isCloudSyncSet bool
)

func init() {
	cloudsyncIns = &CloudSyncImpl{}

}

// GetInstance returns cloudSync instaance
func GetInstance() CloudSync {
	return cloudsyncIns
}

// InitiateCloudSync initiate CloudSync
func (c *CloudSyncImpl) InitiateCloudSync(isCloudSet string) (err error) {
	isCloudSyncSet = false
	if len(isCloudSet) > 0 {
		if strings.Compare(strings.ToLower(isCloudSet), "true") == 0 {
			log.Println("CloudSync init set")
			isCloudSyncSet = true
		}
	}
	return nil
}

//StartCloudSync is used to start the sync by connecting to the broker
func (c *CloudSyncImpl) StartCloudSync(host string) (err error) {
	if isCloudSyncSet && len(host) > 0 {
		log.Info("Starting the CLoudsync Mgr")
		go mqttmgr.StartMQTTClient(host)
	}
	return
}

// RequestCloudSyncConf is  configuration request handler
func (c *CloudSyncImpl) RequestCloudSyncConf(message mqttmgr.Message, topic string, clientID string) string {
	log.Info(logPrefix, "Publishing the data to the cloud")
	mqttClient = mqttmgr.GetClient()
	mqttmgr.SetClientID(clientID)
	resp := ""
	if mqttClient.IsConnected() {
		err := mqttClient.Publish(message, topic)
		if err != nil {
			errMsg := fmt.Sprintf("Error in publishing the data %s", err)
			resp = errMsg
		} else {
			resp = "Data published successfully to Cloud"
		}
	} else {
		resp = "Client not connected to Broker URL"
	}

	return resp
}
