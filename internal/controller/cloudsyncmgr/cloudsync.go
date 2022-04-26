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
	"os"
	"strings"
	"sync"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	mqttmgr "github.com/lf-edge/edge-home-orchestration-go/internal/common/mqtt"
)

const (
	logPrefix             = "[cloudsyncmgr] "
	cloudsyncNotActiveLog = "CloudSync is not Active. Please stop the container and rerun the container with cloudsync set"
)

// CloudSync is the interface for starting Cloud synchronization
type CloudSync interface {
	InitiateCloudSync(isCloudSet string) error
	//implemented by external REST API
	RequestPublish(host string, appID string, message string, topic string) string
	RequestSubscribe(host string, appID string, topic string) string
	RequestSubscribedData(appID string, topic string, host string) string
}

//CloudSyncImpl struct
type CloudSyncImpl struct{}

var (
	cloudsyncIns   *CloudSyncImpl
	log            = logmgr.GetInstance()
	mqttClient     *mqttmgr.Client
	isCloudSyncSet bool
	mqttPort       uint = 1883
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
			log.Info(logPrefix, "CloudSync init set")
			isCloudSyncSet = true
			secure := os.Getenv("SECURE")
			if len(secure) > 0 {
				if strings.Compare(strings.ToLower(secure), "true") == 0 {
					log.Info(logPrefix, "Orchestration init with secure option")
					mqttPort = 8883
				}
			}
			//Intialize the client and hashmap storing client data
			mqttmgr.InitMQTTData()
		}
	}
	return nil
}

// RequestPublish is  configuration request handler
func (c *CloudSyncImpl) RequestPublish(host string, appID string, message string, topic string) string {
	log.Info(logPrefix, "Publishing the data to the cloud")
	resp := ""
	var wg sync.WaitGroup
	if !isCloudSyncSet {
		resp = cloudsyncNotActiveLog
		return resp
	}
	if len(host) == 0 {
		return "No broker host defined"
	}
	wg.Add(1)
	errs := make(chan string, 1)
	go func() {
		errs <- mqttmgr.StartMQTTClient(host, appID, mqttPort)
		resp = <-errs
		wg.Done()

	}()
	wg.Wait()
	if resp != "" {
		errresp := fmt.Sprintf("Error Connecting MQTT -> %s", resp)
		return errresp
	}

	mqttClient = mqttmgr.CheckifClientExist(host)
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

// RequestSubscribe is  configuration request handler
func (c *CloudSyncImpl) RequestSubscribe(host string, appID string, topic string) string {
	log.Info(logPrefix, "Subscribing the data to ", logmgr.SanitizeUserInput(host)) // lgtm [go/log-injection]
	resp := ""
	var wg sync.WaitGroup
	if !isCloudSyncSet {
		resp = cloudsyncNotActiveLog
		return resp
	}
	if len(host) == 0 {
		return "No broker host defined"
	}
	wg.Add(1)
	errs := make(chan string, 1)
	go func() {
		errs <- mqttmgr.StartMQTTClient(host, appID, mqttPort)
		resp = <-errs
		wg.Done()
	}()
	wg.Wait()
	if resp != "" {
		errresp := fmt.Sprintf("Error Connecting MQTT -> %s", resp)
		return errresp
	}
	mqttClient = mqttmgr.CheckifClientExist(host)
	//subscribe function
	if mqttClient != nil && mqttClient.IsConnected() {
		err := mqttClient.Subscribe(topic, appID)
		if err != nil {
			errMsg := fmt.Sprintf("Error in subscribing the data %s", err)
			resp = errMsg
		} else {
			resp = "Successfully subscribed to the topic " + topic
		}
	}
	return resp
}

// RequestSubscribedData is  to get the subscribed data
func (c *CloudSyncImpl) RequestSubscribedData(appID string, topic string, host string) string {
	log.Info(logPrefix, "Requesting the data for ", logmgr.SanitizeUserInput(topic)) // lgtm [go/log-injection]
	resp := ""
	if !isCloudSyncSet {
		resp = cloudsyncNotActiveLog
		return resp
	}
	if len(topic) == 0 {
		return "No topic defined"
	}
	resp = mqttmgr.GetSubscribedData(topic, appID, host)
	return resp
}
