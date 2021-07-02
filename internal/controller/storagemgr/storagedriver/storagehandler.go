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
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/resthelper"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cast"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/pelletier/go-toml"
)

const (
	apiResourceRoute  = common.ApiBase + "/resource/{" + common.DeviceName + "}/{" + common.ResourceName + "}"
	handlerContextKey = "StorageHandler"
	configPath        = "res/configuration.toml"
)

type StorageHandler struct {
	service     *sdk.DeviceService
	logger      logger.LoggingClient
	asyncValues chan<- *models.AsyncValues
	helper      resthelper.RestHelper
}

func NewStorageHandler(service *sdk.DeviceService, logger logger.LoggingClient, asyncValues chan<- *models.AsyncValues) *StorageHandler {
	handler := StorageHandler{
		service:     service,
		logger:      logger,
		asyncValues: asyncValues,
		helper:      resthelper.GetHelper(),
	}

	return &handler
}

func (handler StorageHandler) Start() error {
	if err := handler.service.AddRoute(apiResourceRoute, handler.addContext(deviceHandler), http.MethodPost, http.MethodGet); err != nil {
		return fmt.Errorf("unable to add required route: %s: %s", apiResourceRoute, err.Error())
	}

	handler.logger.Info(fmt.Sprintf("Route %s added.", apiResourceRoute))

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
	deviceName := vars[common.DeviceName]
	resourceName := vars[common.ResourceName]

	handler.logger.Debug(fmt.Sprintf("Received POST for Device=%s Resource=%s", deviceName, resourceName))

	_, err := handler.service.GetDeviceByName(deviceName)
	if err != nil {
		handler.logger.Error(fmt.Sprintf("Incoming reading ignored. Device '%s' not found", deviceName))
		http.Error(writer, fmt.Sprintf("Device not found"), http.StatusNotFound)
		return
	}
	_, ok := handler.service.DeviceResource(deviceName, resourceName)
	if !ok {
		handler.logger.Errorf("Incoming reading ignored. Resource '%s' not found", resourceName)
		http.Error(writer, fmt.Sprintf("Resource not found"), http.StatusNotFound)
		return
	}

	serverIP, readingPort, err := getServerIP(configPath)

	if err != nil {
		http.Error(writer, fmt.Sprintf("Configuration File Not Found"), http.StatusNotFound)
		return
	}

	readingAPI := "/api/v1/reading/name/" + resourceName + "/device/" + deviceName + "/1"

	requestUrl := handler.helper.MakeTargetURL(serverIP, readingPort, readingAPI)
	resp, _, err := handler.helper.DoGet(requestUrl)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Resource not found"), http.StatusNotFound)
		return
	}

	handler.helper.Response(writer, resp, http.StatusOK)
}

func (handler StorageHandler) processAsyncPostRequest(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	deviceName := vars[common.DeviceName]
	resourceName := vars[common.ResourceName]

	handler.logger.Debug(fmt.Sprintf("Received POST for Device=%s Resource=%s", deviceName, resourceName))

	_, err := handler.service.GetDeviceByName(deviceName)
	if err != nil {
		handler.logger.Errorf("Incoming reading ignored. Device '%s' not found", deviceName)
		http.Error(writer, fmt.Sprintf("Device not found"), http.StatusNotFound)
		return
	}

	deviceResource, ok := handler.service.DeviceResource(deviceName, resourceName)
	if !ok {
		handler.logger.Error("Incoming reading ignored. Resource '%s' not found", resourceName)
		http.Error(writer, fmt.Sprintf("Resource not found"), http.StatusNotFound)
		return
	}

	if deviceResource.Properties.MediaType != "" {
		contentType := request.Header.Get(common.ContentType)
		if contentType == "" {
			http.Error(writer, "No Content-Type", http.StatusBadRequest)
			return
		}

		handler.logger.Debug(fmt.Sprintf("Content Type is '%s' & Media Type is '%s' and Type is '%s'",
			contentType, deviceResource.Properties.MediaType, deviceResource.Properties.ValueType))

		if contentType != deviceResource.Properties.MediaType {
			handler.logger.Errorf("Incoming reading ignored. Content Type '%s' doesn't match %s resource's Media Type '%s'",
				contentType, resourceName, deviceResource.Properties.MediaType)

			http.Error(writer, "Wrong Content-Type", http.StatusBadRequest)
			return
		}
	}

	var reading interface{}
	if deviceResource.Properties.ValueType == common.ValueTypeBinary {
		reading, err = handler.readBodyAsBinary(writer, request)
	} else {
		reading, err = handler.readBodyAsString(writer, request)
	}

	if err != nil {
		handler.logger.Error(fmt.Sprintf("Incoming reading ignored. Unable to read request body: %s", err.Error()))
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	value, err := handler.newCommandValue(resourceName, reading, deviceResource.Properties.ValueType)
	if err != nil {
		handler.logger.Error(
			fmt.Sprintf("Incoming reading ignored. Unable to create Command Value for Device=%s Command=%s: %s",
				deviceName, resourceName, err.Error()))
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	asyncValues := &models.AsyncValues{
		DeviceName:    deviceName,
		CommandValues: []*models.CommandValue{value},
	}

	handler.logger.Debug(fmt.Sprintf("Incoming reading received: Device=%s Resource=%s", deviceName, resourceName))

	handler.asyncValues <- asyncValues
}

func (handler StorageHandler) readBodyAsString(writer http.ResponseWriter, request *http.Request) (string, error) {
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

func (handler StorageHandler) readBodyAsBinary(writer http.ResponseWriter, request *http.Request) ([]byte, error) {
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
	}

}

func convertToBase64(val []byte) string {
	// Convert to Base 64
	b64Val := b64.StdEncoding.EncodeToString(val)
	return b64Val
}

func (handler StorageHandler) newCommandValue(resourceName string, reading interface{}, valueType string) (*models.CommandValue, error) {
	var err error
	var result = &models.CommandValue{}
	castError := "fail to parse %v reading, %v"

	if !checkValueInRange(valueType, reading) {
		err = fmt.Errorf("parse reading fail. Reading %v is out of the value type(%v)'s range", reading, valueType)
		handler.logger.Error(err.Error())
		return result, err
	}

	var val, b64Val interface{}
	switch valueType {
	case common.ValueTypeBinary:
		var ok bool
		val, ok = reading.([]byte)
		if !ok {
			return nil, fmt.Errorf(castError, resourceName, "not []byte")
		}
		if resourceName == "jpeg" || resourceName == "png" {
			b64Val = convertToBase64(val.([]byte))
			valueType = common.ValueTypeString
		}

	case common.ValueTypeBool:
		val, err = cast.ToBoolE(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}

	case common.ValueTypeString:
		val, err = cast.ToStringE(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}

	case common.ValueTypeUint8:
		val, err = cast.ToUint8E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}

	case common.ValueTypeUint16:
		val, err = cast.ToUint16E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}

	case common.ValueTypeUint32:
		val, err = cast.ToUint32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}

	case common.ValueTypeUint64:
		val, err = cast.ToUint64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}

	case common.ValueTypeInt8:
		val, err = cast.ToInt8E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}

	case common.ValueTypeInt16:
		val, err = cast.ToInt16E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}

	case common.ValueTypeInt32:
		val, err = cast.ToInt32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}

	case common.ValueTypeInt64:
		val, err = cast.ToInt64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}

	case common.ValueTypeFloat32:
		val, err = cast.ToFloat32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}

	case common.ValueTypeFloat64:
		val, err = cast.ToFloat64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resourceName, err)
		}

	default:
		err = fmt.Errorf("return result fail, none supported value type: %v", valueType)
	}

	if resourceName == "jpeg" || resourceName == "png" {
		result, err = models.NewCommandValue(resourceName, valueType, b64Val)
	} else {
		result, err = models.NewCommandValue(resourceName, valueType, val)
	}
	if err != nil {
		return nil, err
	}

	result.Origin = time.Now().UnixNano()
	return result, nil
}

func checkValueInRange(valueType string, reading interface{}) bool {
	isValid := false

	if valueType == common.ValueTypeString || valueType == common.ValueTypeBool || valueType == common.ValueTypeBinary {
		return true
	}

	if valueType == common.ValueTypeInt8 || valueType == common.ValueTypeInt16 ||
		valueType == common.ValueTypeInt32 || valueType == common.ValueTypeInt64 {
		val := cast.ToInt64(reading)
		isValid = checkIntValueRange(valueType, val)
	}

	if valueType == common.ValueTypeUint8 || valueType == common.ValueTypeUint16 ||
		valueType == common.ValueTypeUint32 || valueType == common.ValueTypeUint64 {
		val := cast.ToUint64(reading)
		isValid = checkUintValueRange(valueType, val)
	}

	if valueType == common.ValueTypeFloat32 || valueType == common.ValueTypeFloat64 {
		val := cast.ToFloat64(reading)
		isValid = checkFloatValueRange(valueType, val)
	}

	return isValid
}

func checkUintValueRange(valueType string, val uint64) bool {
	var isValid = false
	switch valueType {
	case common.ValueTypeUint8:
		if val >= 0 && val <= math.MaxUint8 {
			isValid = true
		}
	case common.ValueTypeUint16:
		if val >= 0 && val <= math.MaxUint16 {
			isValid = true
		}
	case common.ValueTypeUint32:
		if val >= 0 && val <= math.MaxUint32 {
			isValid = true
		}
	case common.ValueTypeUint64:
		maxiMum := uint64(math.MaxUint64)
		if val >= 0 && val <= maxiMum {
			isValid = true
		}
	}
	return isValid
}

func checkIntValueRange(valueType string, val int64) bool {
	var isValid = false
	switch valueType {
	case common.ValueTypeInt8:
		if val >= math.MinInt8 && val <= math.MaxInt8 {
			isValid = true
		}
	case common.ValueTypeInt16:
		if val >= math.MinInt16 && val <= math.MaxInt16 {
			isValid = true
		}
	case common.ValueTypeInt32:
		if val >= math.MinInt32 && val <= math.MaxInt32 {
			isValid = true
		}
	case common.ValueTypeInt64:
		if val >= math.MinInt64 && val <= math.MaxInt64 {
			isValid = true
		}
	}
	return isValid
}

func checkFloatValueRange(valueType string, val float64) bool {
	var isValid = false
	switch valueType {
	case common.ValueTypeFloat32:
		if math.Abs(val) >= math.SmallestNonzeroFloat32 && math.Abs(val) <= math.MaxFloat32 {
			isValid = true
		}
	case common.ValueTypeFloat64:
		if math.Abs(val) >= math.SmallestNonzeroFloat64 && math.Abs(val) <= math.MaxFloat64 {
			isValid = true
		}
	}
	return isValid
}

func getServerIP(ConfigPath string) (string, int, error) {
	config, err := toml.LoadFile(ConfigPath)
	if err != nil {
		return "", 0, err
	}
	return config.Get("Clients.core-data.Host").(string), (int)(config.Get("Clients.core-data.Port").(int64)), nil
}
