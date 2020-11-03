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
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	networkhelper "common/networkhelper"
	"controller/mnedcmgr/server"
	"restinterface/cipher"
	"restinterface/resthelper"
	"restinterface/route/tlspskserver"
	"restinterface/tls"
)

//ServerImpl structure
type ServerImpl struct {
	cipher.HasCipher
	tls.HasCertificate
}

var (
	serverIns      *ServerImpl
	networkIns     networkhelper.Network
	mnedcServerIns server.MNEDCServer
	helper         resthelper.RestHelper
)

func init() {
	networkIns = networkhelper.GetInstance()
	mnedcServerIns = server.GetInstance()
	serverIns = new(ServerImpl)
	helper = resthelper.GetHelper()
}

// GetServerInstance gives the ServerImpl singletone instance
func GetServerInstance() *ServerImpl {
	return serverIns
}

//StartMNEDCServer starts the MNEDC server on the machine
func (ServerImpl) StartMNEDCServer(deviceIDPath string) {

	deviceID, err := discoveryIns.GetDeviceID()
	if err != nil {
		log.Println(logPrefix, "Couldn't start MNEDC server", err.Error())
		return
	}

	_, err = mnedcServerIns.CreateServer("", strconv.Itoa(mnedcServerPort), serverIns.IsSetCert)
	if err != nil {
		log.Println(logPrefix, "Couldn't start MNEDC server", err.Error())
		return
	}

	fatalErrChan := make(chan error)
	mnedcServerIns.Run()
	go serverWaitInterrupt(fatalErrChan)

	privateIP, err := networkIns.GetOutboundIP()
	if err != nil {
		log.Println(logPrefix, "Couldn't start MNEDC server, Error in getting private IP ", err.Error())
		return
	}

	startMNEDCBroadcastServer()
	mnedcServerIns.SetClientIP(deviceID, privateIP, mnedcServerVirtualIP)

	return
}

func startMNEDCBroadcastServer() {
	if !serverIns.IsSetCert {
		http.HandleFunc("/register", handleClientInfo)
		go http.ListenAndServe(":"+strconv.Itoa(broadcastServerPort), nil)
	} else {
		go tlspskserver.TLSPSKServer{}.ListenAndServe(":"+strconv.Itoa(broadcastServerPort), http.HandlerFunc(handleClientInfo))
	}
}

func handleClientInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("Registered")

	encryptBytes, _ := ioutil.ReadAll(r.Body)
	Info, err := serverIns.Key.DecryptByteToJSON(encryptBytes)

	if err != nil {
		log.Printf("[%s] can not decryption %s", logPrefix, err.Error())
		helper.Response(w, http.StatusServiceUnavailable)
		return
	}
	helper.Response(w, http.StatusOK)

	mnedcServerIns.SetClientIP(Info["DeviceID"].(string), Info["PrivateIP"].(string), Info["VirtualIP"].(string))
	broadcastPeers(Info["DeviceID"].(string), Info["PrivateIP"].(string), Info["VirtualIP"].(string))
}

func broadcastPeers(newDeviceID, newPrivateIP, newVirtualIP string) {

	clientInfoMap := mnedcServerIns.GetClientIPMap()

	for uuid, ipInfo := range clientInfoMap {
		virtualIP := ipInfo.VirtualIP
		privateIP := ipInfo.PrivateIP
		deviceID := uuid

		if uuid == newDeviceID {
			continue
		}

		log.Println(logPrefix, "map content: ", virtualIP, privateIP)

		jsonMap := make(map[string]interface{})
		jsonMap["PrivateAddr"] = newPrivateIP
		jsonMap["VirtualAddr"] = newVirtualIP
		jsonMap["DeviceID"] = newDeviceID

		jsonStr, err := serverIns.Key.EncryptJSONToByte(jsonMap)
		if err != nil {
			log.Println(logPrefix, "Error in encrypting jsonMap", err.Error())
			continue
		}
		postInfoToClient(virtualIP, jsonStr)
		jsonMapDB := make(map[string]interface{})
		jsonMapDB["PrivateAddr"] = privateIP
		jsonMapDB["VirtualAddr"] = virtualIP
		jsonMapDB["DeviceID"] = deviceID

		jsonStrDB, err := serverIns.Key.EncryptJSONToByte(jsonMapDB)
		if err != nil {
			log.Println(logPrefix, "Error in encrypting jsonMapDB", err.Error())
			continue
		}
		postInfoToClient(newVirtualIP, jsonStrDB)
	}

}

func postInfoToClient(target string, jsonData []byte) {

	restapi := "/api/v1/discoverymgr/register"

	targetURL := helper.MakeTargetURL(target, internalPort, restapi)

	_, code, err := helper.DoPost(targetURL, jsonData)
	if err != nil || code != http.StatusOK {
		log.Println(logPrefix, "Error in post", err.Error())
	}
}

func serverWaitInterrupt(fatalErrChan chan error) {
	sig := make(chan os.Signal, 2)
	done := make(chan bool, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		done <- true
	}()

	select {
	case <-done:
		log.Println(logPrefix, "Received interrupt, shutting down.")
		err := mnedcServerIns.Close()
		if err != nil {
			log.Println(logPrefix, "Error closing server", err.Error())
		}
	case err := <-fatalErrChan:
		log.Println(logPrefix, "Fatal internal error: ", err)
	}
}
