/*******************************************************************************
 * Copyright 2020 Samsung Electronics All Rights Reserved.
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

package mnedcmgr

import (
	"io/ioutil"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/tls"
	"time"

	restclient "github.com/lf-edge/edge-home-orchestration-go/src/restinterface/client"
	//"controller/discoverymgr"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/client"
)

//ClientImpl structure
type ClientImpl struct {
	clientAPI restclient.Clienter
	tls.HasCertificate
}

var (
	clientIns      *ClientImpl
	mnedcClientIns client.MNEDCClient
	//discoveryIns   discoverymgr.Discovery
)

func init() {
	clientIns = new(ClientImpl)
	mnedcClientIns = client.GetInstance()
	//discoveryIns = discoverymgr.GetInstance()
}

// GetClientInstance gives the ClientImpl singletone instance
func GetClientInstance() *ClientImpl {
	return clientIns
}

//StartMNEDCClient starts the MNEDC client
func (c *ClientImpl) StartMNEDCClient(deviceIDPath string, configPath string) {

	//deviceID, err := discoveryIns.GetDeviceID()
	deviceID, err := getDeviceID(deviceIDPath)
	if err != nil {
		log.Println(logPrefix, "Couldn't start MNEDC client", err.Error())
		return
	}

	mnedcClientIns.SetClient(c.clientAPI)

	err = c.RegisterToMNEDCServer(deviceID, configPath)
	if err != nil {
		log.Println(logPrefix, "Couldn't start MNEDC client", err.Error())
		return
	}

	for attempts := 0; attempts <= maxAttempts; attempts++ {
		//err := discoveryIns.NotifyMNEDCBroadcastServer()
		err := mnedcClientIns.NotifyBroadcastServer(configPath)
		if err != nil {
			log.Println(logPrefix, "Registering to Broadcast server Error", err.Error(), ", retrying")
			time.Sleep(2 * time.Second)
			continue
		}
		return
	}
}

//RegisterToMNEDCServer registers with MNEDC server
func (c *ClientImpl) RegisterToMNEDCServer(deviceID string, configPath string) error {
	_, err := mnedcClientIns.CreateClient(deviceID, configPath, clientIns.IsSetCert)
	if err != nil {
		return err
	}
	mnedcClientIns.Run()
	return nil
}

//SetClient sets the client API
func (c *ClientImpl) SetClient(clientAPI restclient.Clienter) {
	c.clientAPI = clientAPI
	//mnedcClientIns.SetClient(clientAPI)
}

func getDeviceID(path string) (string, error) {
	logPrefix := "[GetDeviceID]"
	UUIDv4, err := ioutil.ReadFile(path)

	if err != nil {
		log.Println(logPrefix, "No saved UUID : ", err.Error())
		return "", err
	}

	log.Println(logPrefix, "Got the UUID")
	UUIDstr := "edge-orchestration-" + string(UUIDv4)

	return UUIDstr, nil
}
