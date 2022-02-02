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

// Package mqtt provides functionalities to handle the client connection to the broker using MQTT
package mqtt

import (
	"encoding/json"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
)

var (
	log       = logmgr.GetInstance()
	logPrefix = "[MQTTConnectionMgr] "
)

// Connect creates a new mqtt client and uses the ClientOptions generated in the NewClient function to connect with the provided host and port.
func (client *Client) Connect() error {

	mqttClient := MQTT.NewClient(client.ClientOptions)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	log.Info(logPrefix, "MQTT Connected")

	client.Lock()
	client.Client = mqttClient

	client.Unlock()

	return nil
}

//onConnect callback is called on successful connection to broker
func (client *Client) onConnect() MQTT.OnConnectHandler {
	log.Info("Running MQTT.OnConnectHandler")
	//Adding the subcription if needed for client
	return nil
}

//onConnectionLost callback is called in case of disconnection from broker
func (client *Client) onConnectionLost() MQTT.ConnectionLostHandler {
	log.Warn(logPrefix, "Connection to the client is disconnected")
	return nil
}

// IsConnected checks if the client is connected or not
func (client *Client) IsConnected() bool {
	//check as client can be null
	if client == nil || client.Client == nil {
		return false
	}

	return client.Client.IsConnected()
}

// Disconnect disconnects the connection
func (client *Client) Disconnect(quiesce uint) {
	log.Info(logPrefix, "Disconnect requested from client")
	if client.Client != nil {
		client.Client.Disconnect(quiesce)
	}
}

// StartMQTTClient is used to initiate the client and set the configuration
func StartMQTTClient(brokerURL string, clientID string, mqttPort uint) string {
	log.Info(logPrefix, "Starting the MQTT Client")
	//Check if the client connection exist
	mqttClient := CheckifClientExist(clientID)
	// Check if the connection exist for same url
	ifConn := checkforConnection(brokerURL, mqttClient, mqttPort)
	if mqttClient != nil && ifConn == 0 {
		log.Info(logPrefix, "Connection Object exist", mqttClient)
		if mqttClient.IsConnected() {
			log.Info(logPrefix, "Client is already connected")
			return ""
		}
		connectErr := mqttClient.Connect()
		if connectErr != nil {
			log.Warn(logPrefix, connectErr)
			return connectErr.Error()
		}

	} else if mqttClient != nil && ifConn != 0 {
		//close previous connection and remove from table
		mqttClient.Disconnect(1)
		delete(clientData, clientID)
	}

	clientConfig, err := NewClient(
		SetHost(brokerURL),
		SetPort(uint(mqttPort)),
		SetClientID(clientID),
	)
	if err != nil {
		log.Warn(logPrefix, err)
		return err.Error()
	}
	clientConfig.ClientOptions.SetOnConnectHandler(clientConfig.onConnect())
	clientConfig.setProtocol()
	URL := clientConfig.SetBrokerURL()
	log.Info(logPrefix, "The broker is", URL)
	clientConfig.URL = URL
	clientConfig.ClientOptions.AddBroker(URL)

	connectErr := clientConfig.Connect()
	if connectErr != nil {
		log.Warn(logPrefix, connectErr)
		return connectErr.Error()
	}
	addClientData(clientConfig, clientID)
	return ""
}

//Publish is used to publish the client data to the cloud
func (client *Client) Publish(message Message, topic string) error {

	log.Info(logPrefix, "Publishing the data to cloud")
	payload, err := json.Marshal(message)
	if err != nil {
		log.Warn(logPrefix, "Error in Json Marshalling", err)
		return err
	}
	mqttClient := client.Client
	for mqttClient == nil {
		time.Sleep(time.Second * 2)
	}
	token := mqttClient.Publish(topic, 0, true, payload)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	time.Sleep(time.Second)
	return nil
}
