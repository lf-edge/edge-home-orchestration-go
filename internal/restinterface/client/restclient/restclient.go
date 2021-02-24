/*******************************************************************************
* Copyright 2019-2020 Samsung Electronics All Rights Reserved.
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

// Package restclient implements REST client functions to send request to remote orchestration
package restclient

import (
	"errors"
	"fmt"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	"net/http"

	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/cipher"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/client"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/resthelper"

	"github.com/gomodule/redigo/redis"
)

type restClientImpl struct {
	internalPort int
	externalPort int

	helper resthelper.RestHelper
	cipher.HasCipher
}

const (
	constWellknownPort = 56001
	constInternalPort  = 56002
	logPrefix          = "[restclient]"
)

var (
	restClient *restClientImpl
	log        = logmgr.GetInstance()
)

func init() {
	restClient = new(restClientImpl)
	restClient.helper = resthelper.GetHelper()
	restClient.externalPort = constWellknownPort
	restClient.internalPort = constInternalPort
}

// GetRestClient returns the singleton restClientImpl instance
func GetRestClient() client.Clienter {
	return restClient
}

// DoExecuteRemoteDevice sends request to remote orchestration (APIV1ServicemgrServicesPost) to execute service
func (c restClientImpl) DoExecuteRemoteDevice(appInfo map[string]interface{}, target string) (err error) {
	log.Printf("%s DoExecuteRemoteDevice : endpoint[%v]", logPrefix, target)
	if c.IsSetKey == false {
		return errors.New(logPrefix + " does not set key")
	}

	restapi := "/api/v1/servicemgr/services"

	targetURL := c.helper.MakeTargetURL(target, c.internalPort, restapi)

	encryptBytes, err := c.Key.EncryptJSONToByte(appInfo)
	if err != nil {
		return errors.New(logPrefix + " can not encryption " + err.Error())
	}

	respBytes, code, err := c.helper.DoPost(targetURL, encryptBytes)
	if err != nil || code != http.StatusOK {
		return errors.New(logPrefix + " post return error")
	}

	respMsg, err := c.Key.DecryptByteToJSON(respBytes)
	if err != nil {
		return errors.New(logPrefix + " can not decrytion " + err.Error())
	}
	log.Printf("%s respMsg From [%v] : %v", logPrefix, target, respMsg)

	str := respMsg["Status"].(string)
	if str == "Failed" {
		err = errors.New("failed")
	}

	return
}

// DoNotifyAppStatusRemoteDevice sends request to remote orchestration (APIV1ServicemgrServicesNotificationServiceIDPost) to notify status
func (c restClientImpl) DoNotifyAppStatusRemoteDevice(statusNotificationInfo map[string]interface{}, appID uint64, target string) error {
	log.Printf("%s DoNotifyAppStatusRemoteDevice : endpoint[%v]", logPrefix, target)
	if c.IsSetKey == false {
		return errors.New(logPrefix + " does not set key")
	}

	restapi := fmt.Sprintf("/api/v1/servicemgr/services/notification/%d", appID)

	targetURL := c.helper.MakeTargetURL(target, c.internalPort, restapi)

	encryptBytes, err := c.Key.EncryptJSONToByte(statusNotificationInfo)
	if err != nil {
		return errors.New(logPrefix + " can not encryption " + err.Error())
	}

	_, code, err := c.helper.DoPost(targetURL, encryptBytes)
	if err != nil || code != http.StatusOK {
		return errors.New(logPrefix + " post return error")
	}

	return nil
}

// DoGetScoreRemoteDevice  sends request to remote orchestration (APIV1ScoringmgrScoreLibnameGet) to get score
func (c restClientImpl) DoGetScoreRemoteDevice(devID string, endpoint string) (scoreValue float64, err error) {
	log.Printf("%s DoGetScoreRemoteDevice : endpoint[%v]", logPrefix, endpoint)
	if c.IsSetKey == false {
		return scoreValue, errors.New(logPrefix + " does not set key")
	}

	restapi := "/api/v1/scoringmgr/score"

	targetURL := c.helper.MakeTargetURL(endpoint, c.internalPort, restapi)

	info := make(map[string]interface{})
	info["devID"] = devID
	encryptBytes, err := c.Key.EncryptJSONToByte(info)
	if err != nil {
		return scoreValue, errors.New(logPrefix + " can not encryption " + err.Error())
	}

	respBytes, code, err := c.helper.DoGetWithBody(targetURL, encryptBytes)
	if err != nil || code != http.StatusOK {
		return scoreValue, errors.New(logPrefix + " get return error")
	}

	respMsg, err := c.Key.DecryptByteToJSON(respBytes)
	if err != nil {
		return scoreValue, errors.New(logPrefix + " can not decryption " + err.Error())
	}
	log.Printf("%s respMsg From [%v] : %v", logPrefix, endpoint, respMsg)

	scoreValue = respMsg["ScoreValue"].(float64)
	if scoreValue == 0.0 {
		err = errors.New("failed")
	}
	return
}

// DoGetResourceRemoteDevice sends request to remote orchestration (APIV1ScoringmgrResourceGet) to get resource values
func (c restClientImpl) DoGetResourceRemoteDevice(devID string, endpoint string) (respMsg map[string]interface{}, err error) {
	log.Printf("%s DoGetResourceRemoteDevice : endpoint[%v]", logPrefix, endpoint)
	if c.IsSetKey == false {
		return respMsg, errors.New(logPrefix + " does not set key")
	}

	restapi := "/api/v1/scoringmgr/resource"

	targetURL := c.helper.MakeTargetURL(endpoint, c.internalPort, restapi)

	info := make(map[string]interface{})
	info["devID"] = devID
	encryptBytes, err := c.Key.EncryptJSONToByte(info)
	if err != nil {
		return respMsg, errors.New(logPrefix + " can not encryption " + err.Error())
	}

	respBytes, code, err := c.helper.DoGetWithBody(targetURL, encryptBytes)
	if err != nil || code != http.StatusOK {
		return respMsg, errors.New(logPrefix + " get return error")
	}

	respMsg, err = c.Key.DecryptByteToJSON(respBytes)
	if err != nil {
		return respMsg, errors.New(logPrefix + " can not decryption " + err.Error())
	}
	log.Printf("%s respMsg From [%v] : %v", logPrefix, endpoint, respMsg)

	if _, found := respMsg["error"]; found {
		err = errors.New("failed")
	}

	return
}

//DoGetOrchestrationInfo requests for orchestration info from endpoint
func (c restClientImpl) DoGetOrchestrationInfo(endpoint string) (string, string, []string, error) {
	if c.IsSetKey == false {
		return "", "", []string{}, errors.New("[" + logPrefix + "] does not set key")
	}

	log.Println(logPrefix, "DoGetOrchestrationInfo", "for", endpoint)

	restapi := "/api/v1/discoverymgr/orchestrationinfo"

	targetURL := c.helper.MakeTargetURL(endpoint, c.internalPort, restapi)

	info := make(map[string]interface{})
	info["devID"] = "DevID"
	encryptBytes, err := c.Key.EncryptJSONToByte(info)
	if err != nil {
		return "", "", []string{}, errors.New("[" + logPrefix + "] can not encryption " + err.Error())
	}

	respBytes, code, err := c.helper.DoGetWithBody(targetURL, encryptBytes)
	if err != nil || code != http.StatusOK {
		return "", "", []string{}, errors.New("[" + logPrefix + "] get return error")
	}

	respMsg, err := c.Key.DecryptByteToJSON(respBytes)
	if err != nil {
		return "", "", []string{}, errors.New("[" + logPrefix + "] can not decryption " + err.Error())
	}

	log.Println("[JSON] : ", respMsg)
	platform := respMsg["Platform"].(string)
	log.Println(logPrefix, "GetOrchInfoResponse", "Platform:", platform)
	executionType := respMsg["ExecutionType"].(string)
	log.Println(logPrefix, "GetOrchInfoResponse", "ExecutionType:", executionType)
	//serviceList := respMsg["ServiceList"].([]string)

	var serviceList []string
	var errRedis error
	serviceList, err = redis.Strings(respMsg["ServiceList"], errRedis)
	if err != nil {
		log.Println(logPrefix, "GetOrchinfoResp", "ServiceList couldnt be populated")
	}
	log.Println(logPrefix, "GetOrchInfoResponse", "ServiceList:", serviceList)
	return platform, executionType, serviceList, nil
}

//DoNotifyMNEDCBroadcastServer sends the device details to MNEDC server
func (c restClientImpl) DoNotifyMNEDCBroadcastServer(endpoint string, port int, deviceID string, privateIP string, virtualIP string) error {
	if c.IsSetKey == false {
		return errors.New("[" + logPrefix + "] does not set key")
	}

	log.Println(logPrefix, "DoNotifyMNEDCBroadcastServer")
	info := make(map[string]interface{})
	info["DeviceID"] = deviceID
	info["VirtualIP"] = virtualIP
	info["PrivateIP"] = privateIP

	encryptBytes, err := c.Key.EncryptJSONToByte(info)
	if err != nil {
		return errors.New("[" + logPrefix + "] can not encryption " + err.Error())
	}

	restapi := "/register"

	targetURL := c.helper.MakeTargetURL(endpoint, port, restapi)

	_, code, err := c.helper.DoPost(targetURL, encryptBytes)
	if err != nil || code != http.StatusOK {
		return err
	}

	return nil
}

func (c *restClientImpl) setHelper(helper resthelper.RestHelper) {
	c.helper = helper
}
