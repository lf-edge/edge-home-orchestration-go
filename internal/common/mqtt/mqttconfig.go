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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	dbhelper "github.com/lf-edge/edge-home-orchestration-go/internal/db/helper"
)

const (
	caCert  = "/ca-crt.pem"
	henCert = "/hen-crt.pem"
	henKey  = "/hen-key.pem"
)

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
	protocol string
}

// Key is composite key for storing topic with url
type Key struct {
	topic, url string
}

var (
	publishData      map[Key]string   //topic-->published data
	subscriptionInfo map[Key][]string //to store the appids mapped to {topic,url}
	//URLData stores the urls mapped to app id
	URLData map[string][]string //to store the appids maped to urls
	//MQTTClient stores the Client object mapped to urls
	MQTTClient          map[string]*Client //to store Homeedge client object for every url
	clientID            string
	certificateFilePath string
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

// InitMQTTData creates an initialized hashmap
func InitMQTTData() {
	certificateFilePath = "/var/edge-orchestration/certs"
	subscriptionInfo = make(map[Key][]string)
	publishData = make(map[Key]string)
	URLData = make(map[string][]string)
	MQTTClient = make(map[string]*Client)
	dbIns := dbhelper.GetInstance()
	clientID, err := dbIns.GetDeviceID()
	if err != nil {
		log.Error(logPrefix, err.Error())
	}
	log.Info(logPrefix, "DeviceId is set as ", clientID)
}

func (c *Client) setProtocol() {
	if c.Port == 8883 {
		c.protocol = "tcps"
	} else {
		c.protocol = "tcp"
	}
}

// CheckifClientExist used to check if the client conn object exist
func CheckifClientExist(url string) *Client {
	client := MQTTClient[url]
	return client
}

// addSubscribeClient is used to add the client info for a topic
func addSubscribeClient(appID string, topic string, url string) {
	subscriptionInfo[Key{topic, url}] = append(subscriptionInfo[Key{topic, url}], appID)
	log.Info(logPrefix, subscriptionInfo[Key{topic, url}])
}

func addPublishedData(topic string, host string, message string) {
	publishData[Key{topic, host}] = message
}

//GetPublishedData is used to get the data for topic subscribed
func GetPublishedData(topic string, clientID string, host string) string {
	list := subscriptionInfo[Key{topic, host}]
	for _, ID := range list {
		if ID == clientID {
			return publishData[Key{topic, host}]
		}
	}
	return "Client is not subscribed to the topic " + topic

}

//SetBrokerURL returns the broker url for connection
func (c *Client) SetBrokerURL() string {
	return fmt.Sprintf("%s://%s:%d", c.protocol, c.Host, c.Port)
}

func checkforConnection(brokerURL string, mqttClient *Client, mqttPort uint) int {
	if mqttClient == nil {
		return 0
	}
	log.Info(logPrefix, mqttClient.URL)
	connURL := fmt.Sprintf("%s://%s:%d", mqttClient.protocol, brokerURL, mqttPort)
	return strings.Compare(connURL, mqttClient.URL)
}

//NewTLSConfig creates a tls config for mqtt client
func NewTLSConfig(certsPath string) (*tls.Config, error) {
	caCertPEM, err := ioutil.ReadFile(certsPath + "/ca-crt.pem")
	if err != nil {
		log.Panic(logPrefix, err.Error())
		return nil, err
	}
	certpool := x509.NewCertPool()
	ok := certpool.AppendCertsFromPEM(caCertPEM)
	if !ok {
		log.Panic(logPrefix, " failed to parse root certificate")
	}

	cert, err := tls.LoadX509KeyPair(certsPath+"/hen-crt.pem", certsPath+"/hen-key.pem")
	if err != nil {
		log.Panic(logPrefix, err.Error())
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		RootCAs:      certpool,
	}, nil
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
	copts.SetClientID(clientID)
	copts.SetMaxReconnectInterval(1 * time.Second)
	copts.SetOnConnectHandler(client.onConnect())
	copts.SetConnectionLostHandler(client.onConnectionLost())
	secure := os.Getenv("SECURE")
	if len(secure) > 0 {
		if strings.Compare(strings.ToLower(secure), "true") == 0 {
			tlsconfig, _ := NewTLSConfig(certificateFilePath)
			copts.SetTLSConfig(tlsconfig)
		}
	}
	client.ClientOptions = copts
	return client, nil
}
