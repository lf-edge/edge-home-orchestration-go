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
	"fmt"
	"strings"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const mqttPort = 1883

// Client is a wrapper on top of `MQTT.Client`
type Client struct {
	ID   string
	Host string
	URL  string
	Port uint
	Qos  byte
	sync.RWMutex
	ClientOptions *MQTT.ClientOptions
	MQTT.Client
}

//Message is used to wrap the app id and payload into one and publish to broker
type Message struct {
	AppID   string
	Payload string
}

var (
	clientData map[string]*Client
)

// Config represents an attribute config setter for the `Client`.
type Config func(*Client)

// SetClientID sets the mqtt client id.
func SetClientID(id string) Config {
	return func(c *Client) {
		c.ID = id
	}
}

// SetHost sets the host where to connect.
func SetHost(host string) Config {
	return func(c *Client) {
		c.Host = host
	}
}

// SetPort sets the port where to connect.
func SetPort(port uint) Config {
	return func(c *Client) {
		c.Port = port
	}
}

// InitClientData creates an initialized hashmap
func InitClientData() {
	clientData = make(map[string]*Client)
}

// CheckifClientExist used to check if the client conn object exist
func CheckifClientExist(clientID string) *Client {
	client := clientData[clientID]
	return client
}

// addClientData is used to add the client object based on client id
func addClientData(client *Client, clientID string) {
	clientData[clientID] = client
}

//SetBrokerURL returns the broker url for connection
func (c *Client) SetBrokerURL(protocol string) string {
	return fmt.Sprintf("%s://%s:%d", protocol, c.Host, c.Port)
}

func checkforConnection(brokerURL string, mqttClient *Client) int {
	if mqttClient == nil {
		return 0
	}
	log.Info(logPrefix, mqttClient.URL)
	connURL := fmt.Sprintf("%s://%s:%d", "tcp", brokerURL, mqttPort)
	return strings.Compare(connURL, mqttClient.URL)
}

// NewClient returns a configured `Client`. Is mandatory
func NewClient(configs ...Config) (*Client, error) {
	client := &Client{
		Qos: byte(1),
	}

	for _, config := range configs {
		config(client)
	}
	copts := MQTT.NewClientOptions()
	copts.SetAutoReconnect(true)
	copts.SetClientID(client.ID)
	copts.SetMaxReconnectInterval(1 * time.Second)
	copts.SetOnConnectHandler(client.onConnect())
	copts.SetConnectionLostHandler(client.onConnectionLost())
	client.ClientOptions = copts

	return client, nil
}
