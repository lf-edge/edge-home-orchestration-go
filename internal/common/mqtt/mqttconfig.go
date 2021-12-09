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
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const clientID = "TestHomeEdge"
const mqttPort = 1883

// Client is a wrapper on top of `MQTT.Client`
type Client struct {
	ID   string
	Host string
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

//SetBrokerURL returns the broker url for connection
func (c *Client) SetBrokerURL(protocol string) string {
	return fmt.Sprintf("%s://%s:%d", protocol, c.Host, c.Port)
}

// NewClient returns a configured `Client`. Is mandatory
func NewClient(configs ...Config) (*Client, error) {
	client := &Client{
		Qos: byte(0),
	}

	for _, config := range configs {
		config(client)
	}

	copts := MQTT.NewClientOptions()
	copts.SetClientID(clientID)
	copts.SetAutoReconnect(true)
	copts.SetMaxReconnectInterval(1 * time.Second)
	copts.SetOnConnectHandler(client.onConnect())
	copts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		log.Warn(logPrefix, " disconnected, reason: "+err.Error())
	})

	client.ClientOptions = copts

	return client, nil
}
