//
// Copyright (c) 2020 Samsung Electronics Co., Ltd All Rights Reserved.
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package storagedriver

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/storagemgr/config"
	dbhelper "github.com/lf-edge/edge-home-orchestration-go/internal/db/helper"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/resthelper"
	"github.com/spf13/cast"

	"github.com/edgexfoundry/device-sdk-go/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type ctxKey string

const (
	deviceNameKey            = "deviceName"
	resourceNameKey          = "resourceName"
	startKey                 = "start"
	endKey                   = "end"
	handlerContextKey ctxKey = "StorageHandler"
	configPath               = "res/configuration.toml"
	numberOfReadings         = "readingCount"
	eventIDKey               = "id"
	event                    = "event"

	apiResourceRoute               = clients.ApiBase + "/device/{" + deviceNameKey + "}/resource/{" + resourceNameKey + "}"
	apiResourceRouteMultiple       = clients.ApiBase + "/device/{" + deviceNameKey + "}/resource/{" + resourceNameKey + "}/{" + numberOfReadings + "}"
	apiWithoutResRoute             = clients.ApiBase + "/resource/{" + resourceNameKey + "}"
	apiWithoutResRouteMultiple     = clients.ApiBase + "/resource/{" + resourceNameKey + "}/{" + numberOfReadings + "}"
	apiWithoutResRouteTimeMultiple = clients.ApiBase + "/start/{" + startKey + "}/end/{" + endKey + "}/{" + numberOfReadings + "}"
	apiEventAddRoute               = clients.ApiBase + "/{" + event + "}"
	apiEventDeleteRoute            = clients.ApiBase + "/event/id/{" + eventIDKey + "}"
)

var (
	dbIns = dbhelper.GetInstance()
)

// StorageHandler provides necessary struct for running DataStorage.
type StorageHandler struct {
	service     *sdk.DeviceService
	logger      logger.LoggingClient
	asyncValues chan<- *models.AsyncValues
	helper      resthelper.RestHelper
}

// NewStorageHandler creates and returns a handler for DataStorage.
func NewStorageHandler(service *sdk.DeviceService, logger logger.LoggingClient, asyncValues chan<- *models.AsyncValues) *StorageHandler {
	handler := StorageHandler{
		service:     service,
		logger:      logger,
		asyncValues: asyncValues,
		helper:      resthelper.GetHelper(),
	}

	return &handler
}

// Start adds routes to DataStorage
func (handler StorageHandler) Start() error {
	if err := handler.service.AddRoute(apiResourceRoute, handler.addContext(deviceHandler), http.MethodPost, http.MethodGet); err != nil {
		return fmt.Errorf("unable to add required route: %s: %s", apiResourceRoute, err.Error())
	}
	if err := handler.service.AddRoute(apiWithoutResRoute, handler.addContext(deviceHandler), http.MethodPost, http.MethodGet); err != nil {
		return fmt.Errorf("unable to add required route: %s: %s", apiWithoutResRoute, err.Error())
	}

	if err := handler.service.AddRoute(apiResourceRouteMultiple, handler.addContext(deviceHandler), http.MethodGet); err != nil {
		return fmt.Errorf("unable to add required route: %s: %s", apiResourceRouteMultiple, err.Error())
	}
	if err := handler.service.AddRoute(apiWithoutResRouteMultiple, handler.addContext(deviceHandler), http.MethodGet); err != nil {
		return fmt.Errorf("unable to add required route: %s: %s", apiWithoutResRouteMultiple, err.Error())
	}
	if err := handler.service.AddRoute(apiWithoutResRouteTimeMultiple, handler.addContext(deviceHandler), http.MethodGet); err != nil {
		return fmt.Errorf("unable to add required route: %s: %s", apiWithoutResRouteTimeMultiple, err.Error())
	}
	if err := handler.service.AddRoute(apiEventAddRoute, handler.addContext(deviceHandler), http.MethodPost); err != nil {
		return fmt.Errorf("unable to add required route: %s: %s", apiEventAddRoute, err.Error())
	}
	if err := handler.service.AddRoute(apiEventDeleteRoute, handler.addContext(deviceHandler), http.MethodDelete); err != nil {
		return fmt.Errorf("unable to add required route: %s: %s", apiEventDeleteRoute, err.Error())
	}
	return nil
}

func (handler StorageHandler) addContext(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	// Add the context with the handler so the endpoint handling code can get back to this handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), handlerContextKey, handler)
		next(w, r.WithContext(ctx))
	})
}

// processGetAsyncRequest is used to handle Async Get Requests
func (handler StorageHandler) processAsyncGetRequest(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	deviceName := vars[deviceNameKey]
	resourceName := vars[resourceNameKey]
	readingCount := vars[numberOfReadings]
	start := vars[startKey]
	end := vars[endKey]

	var readingAPI string
	if readingCount == "" {
		readingCount = "1"
	}
	var err error
	if start != "" && end != "" {
		log.Debug(fmt.Sprintf("Received GET for start=%s end=%s", logmgr.SanitizeUserInput(start), logmgr.SanitizeUserInput(end))) // lgtm [go/log-injection]
		readingAPI = "/api/v1/reading/" + start + "/" + end + "/" + readingCount
	} else {
		if deviceName == "" {
			deviceName, err = dbIns.GetDeviceID()
			if err != nil {
				log.Error("Fail to get deviceName")
				http.Error(writer, "Device not found", http.StatusNotFound)
				return
			}
		}

		log.Debug(fmt.Sprintf("Received POST for Device=%s Resource=%s", logmgr.SanitizeUserInput(deviceName), logmgr.SanitizeUserInput(resourceName))) // lgtm [go/log-injection]

		_, err = handler.service.GetDeviceByName(deviceName)
		if err != nil {
			log.Error(fmt.Sprintf("Incoming reading ignored. Device '%s' not found", logmgr.SanitizeUserInput(deviceName))) // lgtm [go/log-injection]
			http.Error(writer, "Device not found", http.StatusNotFound)
			return
		}
		_, ok := handler.service.DeviceResource(deviceName, resourceName, "get")
		if !ok {
			log.Error(fmt.Sprintf("Incoming reading ignored. Resource '%s' not found", logmgr.SanitizeUserInput(resourceName))) // lgtm [go/log-injection]
			http.Error(writer, "Resource not found", http.StatusNotFound)
			return
		}
		readingAPI = "/api/v1/reading/name/" + resourceName + "/device/" + deviceName + "/" + readingCount
	}
	serverIP, readingPort, err := config.GetServerIP(configPath)

	if err != nil {
		http.Error(writer, "Configuration File Not Found", http.StatusNotFound)
		return
	}

	requestURL := handler.helper.MakeTargetURL(serverIP, readingPort, readingAPI)
	resp, _, err := handler.helper.DoGet(requestURL)
	if err != nil {
		http.Error(writer, "Resource not found", http.StatusNotFound)
		return
	}

	handler.helper.Response(writer, resp, http.StatusOK)
}

// addEvent is used to handle the add event Post requests
func (handler StorageHandler) addEvent(writer http.ResponseWriter, request *http.Request) {
	readingAPI := "/api/v1/" + event
	serverIP, readingPort, err := config.GetServerIP(configPath)

	if err != nil {
		http.Error(writer, "Configuration File Not Found", http.StatusNotFound)
		return
	}

	requestURL := handler.helper.MakeTargetURL(serverIP, readingPort, readingAPI)
	reqBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Error in getting body", http.StatusBadRequest)
	}
	resp, statusCode, err := handler.helper.DoPost(requestURL, reqBody)
	if err != nil {
		http.Error(writer, "Error in doing Post request :", statusCode)
		return
	}
	eventid := string(resp)
	log.Debug(fmt.Sprintf("Event Id is %s", eventid))
	handler.helper.Response(writer, resp, http.StatusOK)
}

// processAsyncPostRequest is used to handle Async POST Requests
func (handler StorageHandler) processAsyncPostRequest(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	event := vars[event]
	if event != "" {
		handler.addEvent(writer, request)
		return
	}
	deviceName := vars[deviceNameKey]
	var err error
	if deviceName == "" {
		deviceName, err = dbIns.GetDeviceID()
		if err != nil {
			log.Error("Fail to get deviceName")
			http.Error(writer, "Device not found", http.StatusNotFound)
		}
	}
	resourceName := vars[resourceNameKey]

	log.Debug(fmt.Sprintf("Received POST for Device=%s Resource=%s", logmgr.SanitizeUserInput(deviceName), logmgr.SanitizeUserInput(resourceName))) // lgtm [go/log-injection]

	_, err = handler.service.GetDeviceByName(deviceName)
	if err != nil {
		log.Error(fmt.Sprintf("Incoming reading ignored. Device '%s' not found", logmgr.SanitizeUserInput(deviceName))) // lgtm [go/log-injection]
		http.Error(writer, "Device not found", http.StatusNotFound)
		return
	}

	deviceResource, ok := handler.service.DeviceResource(deviceName, resourceName, "get")
	if !ok {
		log.Error(fmt.Sprintf("Incoming reading ignored. Resource '%s' not found", logmgr.SanitizeUserInput(resourceName))) // lgtm [go/log-injection]
		http.Error(writer, "Resource not found", http.StatusNotFound)
		return
	}

	if deviceResource.Properties.Value.MediaType != "" {
		contentType := request.Header.Get(clients.ContentType)
		if contentType == "" {
			http.Error(writer, "No Content-Type", http.StatusBadRequest)
			return
		}

		log.Debug(fmt.Sprintf("Content Type is '%s' & Media Type is '%s' and Type is '%s'", logmgr.SanitizeUserInput(contentType), // lgtm [go/log-injection]
			logmgr.SanitizeUserInput(deviceResource.Properties.Value.MediaType), logmgr.SanitizeUserInput(deviceResource.Properties.Value.Type))) // lgtm [go/log-injection]

		if contentType != deviceResource.Properties.Value.MediaType {
			log.Error(fmt.Sprintf("Incoming reading ignored. Content Type '%s' doesn't match %s resource's Media Type '%s'", logmgr.SanitizeUserInput(contentType), // lgtm [go/log-injection]
				logmgr.SanitizeUserInput(resourceName), logmgr.SanitizeUserInput(deviceResource.Properties.Value.MediaType))) // lgtm [go/log-injection]

			http.Error(writer, "Wrong Content-Type", http.StatusBadRequest)
			return
		}
	}

	var reading interface{}
	readingType := models.ParseValueType(deviceResource.Properties.Value.Type)

	if readingType == models.Binary {
		reading, err = handler.readBodyAsBinary(request)
	} else {
		reading, err = handler.readBodyAsString(request)
	}

	if err != nil {
		log.Error(fmt.Sprintf("Incoming reading ignored. Unable to read request body: %s", err.Error()))
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	value, err := handler.newCommandValue(resourceName, reading, readingType)
	if err != nil {
		log.Error(
			fmt.Sprintf("Incoming reading ignored. Unable to create Command Value for Device=%s Command=%s: %s", // lgtm [go/log-injection]
				logmgr.SanitizeUserInput(deviceName), logmgr.SanitizeUserInput(resourceName), err.Error())) // lgtm [go/log-injection]
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	asyncValues := &models.AsyncValues{
		DeviceName:    deviceName,
		CommandValues: []*models.CommandValue{value},
	}

	log.Debug(fmt.Sprintf("Incoming reading received: Device=%s Resource=%s", logmgr.SanitizeUserInput(deviceName), logmgr.SanitizeUserInput(resourceName))) // lgtm [go/log-injection]

	handler.asyncValues <- asyncValues
}

// processDeleteRequest is used to handle the Async Delete Request
func (handler StorageHandler) processDeleteRequest(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	ID := vars[eventIDKey]

	log.Debug(fmt.Sprintf("Received DELETE for Event ID=%s", logmgr.SanitizeUserInput(ID))) // lgtm [go/log-injection]
	var deleteAPI string
	if ID != "" {
		deleteAPI = "/api/v1/event/id/" + ID
	}
	serverIP, readingPort, err := config.GetServerIP(configPath)
	if err != nil {
		http.Error(writer, "Configuration File Not Found", http.StatusNotFound)
		return
	}
	requestURL := handler.helper.MakeTargetURL(serverIP, readingPort, deleteAPI)
	resp, _, err := handler.helper.DoDelete(requestURL)
	if err != nil {
		http.Error(writer, "Resource not found", http.StatusNotFound)
		return
	}
	handler.helper.Response(writer, resp, http.StatusOK)
}

func (handler StorageHandler) readBodyAsString(request *http.Request) (string, error) {
	if request.Body == nil {
		return "", fmt.Errorf("no request body provided")
	}

	defer request.Body.Close()
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return "", err
	}

	if len(body) == 0 {
		return "", fmt.Errorf("no request body provided")
	}

	return string(body), nil
}

func (handler StorageHandler) readBodyAsBinary(request *http.Request) ([]byte, error) {
	if request.Body == nil {
		return nil, fmt.Errorf("no request body provided")
	}

	defer request.Body.Close()
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("no request body provided")
	}

	return body, nil
}

func deviceHandler(writer http.ResponseWriter, request *http.Request) {
	handler, ok := request.Context().Value(handlerContextKey).(StorageHandler)
	if !ok {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Bad context pass to handler"))
		return
	}
	switch request.Method {
	case "GET":
		handler.processAsyncGetRequest(writer, request)
	case "POST":
		handler.processAsyncPostRequest(writer, request)
	case "DELETE":
		handler.processDeleteRequest(writer, request)
	}

}

func convertToBase64(val []byte) string {
	// Convert to Base 64
	Base64Val := b64.StdEncoding.EncodeToString(val)
	return Base64Val
}

func (handler StorageHandler) newCommandValue(resourceName string, reading interface{}, valueType models.ValueType) (*models.CommandValue, error) {
	var result = &models.CommandValue{}
	var errn error
	var timestamp = time.Now().UnixNano()
	castError := "fail to parse %v reading, %v"

	if !checkValueInRange(valueType, reading) {
		errn = fmt.Errorf("parse reading fail. Reading %v is out of the value type(%v)'s range", reading, valueType)
		log.Error(errn.Error())
		return result, errn
	}

	switch valueType {
	case models.Binary:
		val, ok := reading.([]byte)
		if !ok {
			return nil, fmt.Errorf(castError, resourceName, "not []byte")
		}
		if resourceName == "jpeg" || resourceName == "png" {
			Base64Val := convertToBase64(val)
			valueType = models.String
			result, errn = models.NewCommandValue(resourceName, timestamp, Base64Val, valueType)
		} else {
			result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)
		}

	case models.Bool:
		val, err := cast.ToBoolE(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}
		result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)

	case models.String:
		val, err := cast.ToStringE(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}
		result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)

	case models.Uint8:
		val, err := cast.ToUint8E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}
		result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)

	case models.Uint16:
		val, err := cast.ToUint16E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}
		result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)

	case models.Uint32:
		val, err := cast.ToUint32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}
		result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)

	case models.Uint64:
		val, err := cast.ToUint64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}
		result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)

	case models.Int8:
		val, err := cast.ToInt8E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}
		result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)

	case models.Int16:
		val, err := cast.ToInt16E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}
		result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)

	case models.Int32:
		val, err := cast.ToInt32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}
		result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)

	case models.Int64:
		val, err := cast.ToInt64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}
		result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)

	case models.Float32:
		val, err := cast.ToFloat32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}
		result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)

	case models.Float64:
		val, err := cast.ToFloat64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}
		result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)

	default:
		errn = fmt.Errorf("return result fail, none supported value type: %v", valueType)
	}

	return result, errn
}

func checkValueInRange(valueType models.ValueType, reading interface{}) bool {
	isValid := false

	if valueType == models.String || valueType == models.Bool || valueType == models.Binary {
		return true
	}

	if valueType == models.Int8 || valueType == models.Int16 ||
		valueType == models.Int32 || valueType == models.Int64 {
		val := cast.ToInt64(reading)
		isValid = checkIntValueRange(valueType, val)
	}

	if valueType == models.Uint8 || valueType == models.Uint16 ||
		valueType == models.Uint32 || valueType == models.Uint64 {
		val := cast.ToUint64(reading)
		isValid = checkUintValueRange(valueType, val)
	}

	if valueType == models.Float32 || valueType == models.Float64 {
		val := cast.ToFloat64(reading)
		isValid = checkFloatValueRange(valueType, val)
	}

	return isValid
}

func checkUintValueRange(valueType models.ValueType, val uint64) bool {
	var isValid = false
	switch valueType {
	case models.Uint8:
		if val <= math.MaxUint8 {
			isValid = true
		}
	case models.Uint16:
		if val <= math.MaxUint16 {
			isValid = true
		}
	case models.Uint32:
		if val <= math.MaxUint32 {
			isValid = true
		}
	case models.Uint64:
		maxiMum := uint64(math.MaxUint64)
		if val <= maxiMum {
			isValid = true
		}
	}
	return isValid
}

func checkIntValueRange(valueType models.ValueType, val int64) bool {
	var isValid = false
	switch valueType {
	case models.Int8:
		if val >= math.MinInt8 && val <= math.MaxInt8 {
			isValid = true
		}
	case models.Int16:
		if val >= math.MinInt16 && val <= math.MaxInt16 {
			isValid = true
		}
	case models.Int32:
		if val >= math.MinInt32 && val <= math.MaxInt32 {
			isValid = true
		}
	case models.Int64:
		isValid = true
	}
	return isValid
}

func checkFloatValueRange(valueType models.ValueType, val float64) bool {
	var isValid = false
	switch valueType {
	case models.Float32:
		if math.Abs(val) >= math.SmallestNonzeroFloat32 && math.Abs(val) <= math.MaxFloat32 {
			isValid = true
		}
	case models.Float64:
		if math.Abs(val) >= math.SmallestNonzeroFloat64 && math.Abs(val) <= math.MaxFloat64 {
			isValid = true
		}
	}
	return isValid
}
