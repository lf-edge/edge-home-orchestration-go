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
package mqtt

import (
	"strings"
	"testing"
)

const Host = "broker.hivemq.com"

var port uint = 1883

const InvalidHost = "invalid"

func initializeTest(Host string, clientID string) {
	InitMQTTData()
	StartMQTTClient(Host, clientID, port)
}

func TestStartMQTTClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		initializeTest(Host, "testClient")
		client := publishData["testClient"]
		if client != nil {
			message := "Test data for testing"
			client.Publish(message, "home/livingroom")
		}
		err := StartMQTTClient(Host, "testClient", port)
		expected := ""
		if !strings.Contains(err, expected) {
			t.Error("Unexpected err", err)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		InitMQTTData()
		err := StartMQTTClient(InvalidHost, "testClientFailure", port)
		expected := "dial tcp: lookup invalid: Temporary failure in name resolution"
		if !strings.Contains(err, expected) {
			t.Error("Unexpected err", err)
		}
	})
}

func TestCheckforConnection(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		initializeTest(Host, "CheckConnection")
		client := publishData["CheckConnection"]
		isConnected := checkforConnection(Host, client, port)
		expected := 0
		if isConnected != expected {
			t.Errorf("Expected %d But received %d", expected, isConnected)
		}
		client.Disconnect(1)
	})
}

func TestIsClientConnected(t *testing.T) {
	t.Run("Fail", func(t *testing.T) {
		client := publishData["CheckConnection"]
		connStatus := client.IsConnected()
		expected := false
		if connStatus != expected {
			t.Errorf("Expected %v but received %v", expected, connStatus)
		}
	})
}
