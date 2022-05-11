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
	if client != nil && client.Client != nil {
		client.Client.Disconnect(quiesce)
	}
}

// StartMQTTClient is used to initiate the client and set the configuration
func StartMQTTClient(brokerURL string, appID string, mqttPort uint) string {
	log.Info(logPrefix, "Starting the MQTT Client")
	//Check if the client connection exist
	mqttClient := CheckifClientExist(brokerURL)
	if mqttClient != nil {
		log.Info(logPrefix, "Client is already connected")
		//add it to URL Data if not present
		appList := URLData[brokerURL]
		for _, id := range appList {
			if id == appID {
				return ""
			}
		}
		URLData[brokerURL] = append(URLData[brokerURL], appID)
		return ""
	}
	//new client object is created to add to MQTTClient and connection is established and added to URLData
	clientConfig, _ := NewClient(
		SetHost(brokerURL),
		SetPort(uint(mqttPort)),
		SetClientID(clientID),
	)
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
	MQTTClient[brokerURL] = clientConfig
	URLData[brokerURL] = append(URLData[brokerURL], appID)
	//log.Info(logPrefix, URLData[brokerURL])
	return ""
}

//Publish is used to publish the client data to the cloud
func (client *Client) Publish(message string, topic string) error {

	log.Info(logPrefix, "Publishing the data to cloud")
	mqttClient := client.Client
	for mqttClient == nil {
		time.Sleep(time.Second * 2)
	}
	token := mqttClient.Publish(topic, 0, true, message)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	time.Sleep(time.Second)
	return nil
}

var messageHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {

	log.Info(logPrefix, "Topic ", msg.Topic(), " registered.. ", string(msg.Payload()), " is the payload")
	options := client.OptionsReader()
	path := options.Servers()
	addPublishedData(msg.Topic(), path[0].Hostname(), string(msg.Payload()))
}

//Subscribe is used to subscribe to a topic
func (client *Client) Subscribe(topic string, appID string) error {
	log.Info(logPrefix, "Subscribing to ", logmgr.SanitizeUserInput(topic)) // lgtm [go/log-injection]
	mqttClient := client.Client
	for mqttClient == nil {
		time.Sleep(time.Second * 2)
	}
	token := mqttClient.Subscribe(topic, 0, messageHandler)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	time.Sleep(time.Second)
	addSubscribeClient(appID, topic, client.Host)
	return nil
}
