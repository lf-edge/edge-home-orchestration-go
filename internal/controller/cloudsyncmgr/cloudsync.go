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
	"sync"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	mqttmgr "github.com/lf-edge/edge-home-orchestration-go/internal/common/mqtt"
)

const (
	logPrefix = "[cloudsyncmgr]"
)

// CloudSync is the interface for starting Cloud synchronization
type CloudSync interface {
	InitiateCloudSync(isCloudSet string) error
	//implemented by external REST API
	RequestCloudSyncConf(host string, clientID string, message mqttmgr.Message, topic string) string
}

//CloudSyncImpl struct
type CloudSyncImpl struct{}

var (
	cloudsyncIns   *CloudSyncImpl
	log            = logmgr.GetInstance()
	mqttClient     *mqttmgr.Client
	isCloudSyncSet bool
	mqttPort       uint
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
			//Intialize the client and hashmap storing client data
			mqttmgr.InitClientData()
		}
	}
	return nil
}

// RequestCloudSyncConf is  configuration request handler
func (c *CloudSyncImpl) RequestCloudSyncConf(host string, clientID string, message mqttmgr.Message, topic string) string {
	log.Info(logPrefix, "Publishing the data to the cloud")
	resp := ""
	var wg sync.WaitGroup
	if !isCloudSyncSet {
		resp = "CloudSync is not Active. Please stop the container and rerun the container with cloudsync set"
		return resp
	}
	if len(host) == 0 {
		return "No broker host defined"
	}
	wg.Add(1)
	errs := make(chan string, 1)
	go func() {
		mqttPort = 8883
		errs <- mqttmgr.StartMQTTClient(host, clientID, mqttPort)
		resp = <-errs
		wg.Done()

	}()
	wg.Wait()
	if resp != "" {
		errresp := fmt.Sprintf("Error Connecting MQTT -> %s", resp)
		return errresp
	}

	mqttClient = mqttmgr.CheckifClientExist(clientID)
	if mqttClient != nil && mqttClient.IsConnected() {
		err := mqttClient.Publish(message, topic)
		if err != nil {
			errMsg := fmt.Sprintf("Error in publishing the data %s", err)
			resp = errMsg
		} else {
			resp = "Data published successfully to Cloud" + mqttClient.URL
		}
	}
	return resp
}
