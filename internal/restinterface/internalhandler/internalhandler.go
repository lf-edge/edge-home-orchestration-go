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

// Package internalhandler implements REST server functions to communication between orchestrations
package internalhandler

import (
	"io/ioutil"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	"net"
	"net/http"
	"strings"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/commandvalidator"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/requestervalidator"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/types/servicemgrtypes"
	"github.com/lf-edge/edge-home-orchestration-go/internal/orchestrationapi"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/cipher"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/resthelper"
)

const logPrefix = "RestInternalInterface"

// Handler struct
type Handler struct {
	isSetAPI bool
	api      orchestrationapi.OrcheInternalAPI

	helper resthelper.RestHelper

	restinterface.HasRoutes
	cipher.HasCipher
}

var (
	handler *Handler
	log     = logmgr.GetInstance()
)

func init() {
	handler = new(Handler)
	handler.helper = resthelper.GetHelper()
	handler.Routes = restinterface.Routes{
		restinterface.Route{
			Name:        "APIV1Ping",
			Method:      strings.ToUpper("Get"),
			Pattern:     "/api/v1/ping",
			HandlerFunc: handler.APIV1Ping,
		},

		restinterface.Route{
			Name:        "APIV1ServicemgrServicesPost",
			Method:      strings.ToUpper("Post"),
			Pattern:     "/api/v1/servicemgr/services",
			HandlerFunc: handler.APIV1ServicemgrServicesPost,
		},

		restinterface.Route{
			Name:        "APIV1ServicemgrServicesNotificationServiceIDPost",
			Method:      strings.ToUpper("Post"),
			Pattern:     "/api/v1/servicemgr/services/notification/{serviceid}",
			HandlerFunc: handler.APIV1ServicemgrServicesNotificationServiceIDPost,
		},

		restinterface.Route{
			Name:        "APIV1ScoringmgrScoreLibnameGet",
			Method:      strings.ToUpper("Get"),
			Pattern:     "/api/v1/scoringmgr/score",
			HandlerFunc: handler.APIV1ScoringmgrScoreLibnameGet,
		},

		restinterface.Route{
			Name:        "APIV1ScoringmgrResourceGet",
			Method:      strings.ToUpper("Get"),
			Pattern:     "/api/v1/scoringmgr/resource",
			HandlerFunc: handler.APIV1ScoringmgrResourceGet,
		},
		restinterface.Route{
			Name:        "APIV1DiscoverymgrMNEDCDeviceInfoPost",
			Method:      strings.ToUpper("Post"),
			Pattern:     "/api/v1/discoverymgr/register",
			HandlerFunc: handler.APIV1DiscoverymgrMNEDCDeviceInfoPost,
		},

		restinterface.Route{
			Name:        "APIV1DiscoverymgrOrchestrationInfoGet",
			Method:      strings.ToUpper("Get"),
			Pattern:     "/api/v1/discoverymgr/orchestrationinfo",
			HandlerFunc: handler.APIV1DiscoverymgrOrchestrationInfoGet,
		},
	}
}

// GetHandler returns the singleton Handler instance
func GetHandler() *Handler {
	return handler
}

// SetOrchestrationAPI sets OrcheInternalAPI
func (h *Handler) SetOrchestrationAPI(o orchestrationapi.OrcheInternalAPI) {
	h.api = o
	h.isSetAPI = true
}

func (h *Handler) SetCertificateFilePath(path string) {
	rh := resthelper.GetHelperWithCertificate()
	rh.SetCertificateFilePath(path)
	h.helper = resthelper.GetHelper()
}

// APIV1Ping handles ping request from remote orchestration
func (h *Handler) APIV1Ping(w http.ResponseWriter, r *http.Request) {
	h.helper.Response(w, http.StatusOK)
}

// APIV1ServicemgrServicesPost handles service execution request from remote orchestration
func (h *Handler) APIV1ServicemgrServicesPost(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] APIV1ServicemgrServicesPost", logPrefix)
	if h.isSetAPI == false {
		log.Printf("[%s] does not set api", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	} else if h.IsSetKey == false {
		log.Printf("[%s] does not set key", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	remoteAddr, _, _ := net.SplitHostPort(r.RemoteAddr)
	encryptBytes, _ := ioutil.ReadAll(r.Body)

	appInfo, err := h.Key.DecryptByteToJSON(encryptBytes)
	if err != nil {
		log.Printf("[%s] can not decryption", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	appInfo["NotificationTargetURL"] = remoteAddr

	log.Printf("[%s] Requested AppInfo", logPrefix)
	log.Printf("[%s] Requester    : %v", logPrefix, appInfo["Requester"].(string))
	log.Printf("[%s] ServiceID    : %v", logPrefix, appInfo["ServiceID"].(float64))
	log.Printf("[%s] ServiceName  : %v", logPrefix, appInfo["ServiceName"].(string))
	log.Printf("[%s] NotificationTargetURL : %v", logPrefix, appInfo["NotificationTargetURL"].(string))
	log.Printf("[%s] ExecutionCmd : %v", logPrefix, appInfo["UserArgs"].([]interface{}))

	args := make([]string, 0)
	for _, arg := range appInfo["UserArgs"].([]interface{}) {
		args = append(args, arg.(string))
	}
	executionType := args[len(args)-1]
	args = args[:len(args)-1]

	if executionType != "container" {
		serviceName := appInfo["ServiceName"].(string)
		requester := appInfo["Requester"].(string)
		vRequester := requestervalidator.RequesterValidator{}
		if err := vRequester.CheckRequester(serviceName, requester); err != nil {
			log.Printf("[%s] ", err.Error())
			h.helper.Response(w, http.StatusBadRequest)
			return
		}

		validator := commandvalidator.CommandValidator{}
		if err := validator.CheckCommand(serviceName, args); err != nil {
			log.Printf("[%s] ", err.Error())
			h.helper.Response(w, http.StatusBadRequest)
			return
		}
	}

	h.api.ExecuteAppOnLocal(appInfo)

	respJSONMsg := make(map[string]interface{})
	respJSONMsg["Status"] = servicemgrtypes.ConstServiceStatusStarted

	respEncryptBytes, err := h.Key.EncryptJSONToByte(respJSONMsg)
	if err != nil {
		log.Printf("[%s] can not encryption", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	h.helper.ResponseJSON(w, respEncryptBytes, http.StatusOK)
}

// APIV1ServicemgrServicesNotificationServiceIDPost handles service notification request from remote orchestration
func (h *Handler) APIV1ServicemgrServicesNotificationServiceIDPost(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] APIV1ServicemgrServicesNotificationServiceIDPost", logPrefix)
	if h.isSetAPI == false {
		log.Printf("[%s] does not set api", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	} else if h.IsSetKey == false {
		log.Printf("[%s] does not set key", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	encryptBytes, _ := ioutil.ReadAll(r.Body)

	statusNotification, err := h.Key.DecryptByteToJSON(encryptBytes)
	if err != nil {
		log.Printf("[%s] can not decryption", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	serviceID := statusNotification["ServiceID"].(float64)
	status := statusNotification["Status"].(string)

	err = h.api.HandleNotificationOnLocal(serviceID, status)
	if err != nil {
		h.helper.Response(w, http.StatusInternalServerError)
		return
	}

	handler.helper.Response(w, http.StatusOK)
}

// APIV1ScoringmgrScoreLibnameGet handles scoring request from remote orchestration
func (h *Handler) APIV1ScoringmgrScoreLibnameGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] APIV1ScoringmgrScoreLibnameGet", logPrefix)
	if h.isSetAPI == false {
		log.Printf("[%s] does not set api", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	} else if h.IsSetKey == false {
		log.Printf("[%s] does not set key", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	encryptBytes, _ := ioutil.ReadAll(r.Body)
	Info, err := h.Key.DecryptByteToJSON(encryptBytes)
	if err != nil {
		log.Printf("[%s] can not decryption %s", logPrefix, err.Error())
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	devID := Info["devID"]

	scoreValue, err := h.api.GetScore(devID.(string))
	if err != nil {
		log.Printf("[%s] GetScore fail : %s", logPrefix, err.Error())
		h.helper.Response(w, http.StatusInternalServerError)
		return
	}

	respJSONMsg := make(map[string]interface{})
	respJSONMsg["ScoreValue"] = scoreValue

	respEncryptBytes, err := h.Key.EncryptJSONToByte(respJSONMsg)
	if err != nil {
		log.Printf("[%s] can not encryption %s", logPrefix, err.Error())
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	h.helper.ResponseJSON(w, respEncryptBytes, http.StatusOK)
}

// APIV1ScoringmgrResourceGet handles Resource request from remote orchestration
func (h *Handler) APIV1ScoringmgrResourceGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] APIV1ScoringmgrResourceGet", logPrefix)
	if h.isSetAPI == false {
		log.Printf("[%s] does not set api", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	} else if h.IsSetKey == false {
		log.Printf("[%s] does not set key", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	encryptBytes, _ := ioutil.ReadAll(r.Body)
	Info, err := h.Key.DecryptByteToJSON(encryptBytes)
	if err != nil {
		log.Printf("[%s] can not decryption %s", logPrefix, err.Error())
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	devID := Info["devID"]

	resourceValue, err := h.api.GetResource(devID.(string))
	if err != nil {
		log.Printf("[%s] GetResource fail : %s", logPrefix, err.Error())
		h.helper.Response(w, http.StatusInternalServerError)
		return
	}

	respEncryptBytes, err := h.Key.EncryptJSONToByte(resourceValue)
	if err != nil {
		log.Printf("[%s] can not encryption %s", logPrefix, err.Error())
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	h.helper.ResponseJSON(w, respEncryptBytes, http.StatusOK)
}

//APIV1DiscoverymgrMNEDCDeviceInfoPost handles device info from MNEDC server
func (h *Handler) APIV1DiscoverymgrMNEDCDeviceInfoPost(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] APIV1DiscoveryFromMNEDCServer", logPrefix)
	if h.isSetAPI == false {
		log.Printf("[%s] does not set api", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	} else if h.IsSetKey == false {
		log.Printf("[%s] does not set key", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	encryptBytes, _ := ioutil.ReadAll(r.Body)
	Info, err := h.Key.DecryptByteToJSON(encryptBytes)

	if err != nil {
		log.Printf("[%s] can not decryption %s", logPrefix, err.Error())
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	log.Println(logPrefix, "Info from MNEDC server received")
	log.Println(logPrefix, "Device ID:", Info["DeviceID"].(string))
	log.Println(logPrefix, "Private Add:", Info["PrivateAddr"].(string))
	log.Println(logPrefix, "Virtual Add:", Info["VirtualAddr"].(string))

	devID := Info["DeviceID"].(string)
	privateIP := Info["PrivateAddr"].(string)
	virtualIP := Info["VirtualAddr"].(string)

	h.api.HandleDeviceInfo(devID, virtualIP, privateIP)
	handler.helper.Response(w, http.StatusOK)
}

//APIV1DiscoverymgrOrchestrationInfoGet handles device info requests from peers
func (h *Handler) APIV1DiscoverymgrOrchestrationInfoGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] APIV1DiscoverymgrOrchestrationInfoGet", logPrefix)
	if h.isSetAPI == false {
		log.Printf("[%s] does not set api", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	} else if h.IsSetKey == false {
		log.Printf("[%s] does not set key", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	platform, execution, serviceList, err := h.api.GetOrchestrationInfo()

	if err != nil {
		log.Printf("[%s] can not encryption %s", logPrefix, err.Error())
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	respJSONMsg := make(map[string]interface{})
	respJSONMsg["Platform"] = platform
	respJSONMsg["ExecutionType"] = execution
	respJSONMsg["ServiceList"] = serviceList

	respEncryptBytes, err := h.Key.EncryptJSONToByte(respJSONMsg)
	if err != nil {
		log.Printf("[%s] can not encryption %s", logPrefix, err.Error())
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	h.helper.ResponseJSON(w, respEncryptBytes, http.StatusOK)
}

func (h *Handler) setHelper(helper resthelper.RestHelper) {
	h.helper = helper
}
