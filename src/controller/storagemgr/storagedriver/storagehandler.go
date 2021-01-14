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
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cast"

	sdk "github.com/edgexfoundry/device-sdk-go"
	"github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

const (
	deviceNameKey     = "deviceName"
	resourceNameKey   = "resourceName"
	apiResourceRoute  = clients.ApiBase + "/resource/{" + deviceNameKey + "}/{" + resourceNameKey + "}"
	handlerContextKey = "RestHandler"
)

type StorageHandler struct {
	service     *sdk.Service
	logger      logger.LoggingClient
	asyncValues chan<- *models.AsyncValues
}

func NewStorageHandler(service *sdk.Service, logger logger.LoggingClient, asyncValues chan<- *models.AsyncValues) *StorageHandler {
	handler := StorageHandler{
		service:     service,
		logger:      logger,
		asyncValues: asyncValues,
	}

	return &handler
}

func (handler StorageHandler) Start() error {
	if err := handler.service.AddRoute(apiResourceRoute, handler.addContext(deviceHandler), http.MethodPost); err != nil {
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

func (handler StorageHandler) processAsyncRequest(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	deviceName := vars[deviceNameKey]
	resourceName := vars[resourceNameKey]

	handler.logger.Debug(fmt.Sprintf("Received POST for Device=%s Resource=%s", deviceName, resourceName))

	_, err := handler.service.GetDeviceByName(deviceName)
	if err != nil {
		handler.logger.Error(fmt.Sprintf("Incoming reading ignored. Device '%s' not found", deviceName))
		http.Error(writer, fmt.Sprintf("Device not found"), http.StatusNotFound)
		return
	}

	deviceResource, ok := handler.service.DeviceResource(deviceName, resourceName, "get")
	if !ok {
		handler.logger.Error(fmt.Sprintf("Incoming reading ignored. Resource '%s' not found", resourceName))
		http.Error(writer, fmt.Sprintf("Resource not found"), http.StatusNotFound)
		return
	}

	if deviceResource.Properties.Value.MediaType != "" {
		contentType := request.Header.Get(clients.ContentType)
		if contentType == "" {
			http.Error(writer, "No Content-Type", http.StatusBadRequest)
			return
		}

		handler.logger.Debug(fmt.Sprintf("Content Type is '%s' & Media Type is '%s' and Type is '%s'",
			contentType, deviceResource.Properties.Value.MediaType, deviceResource.Properties.Value.Type))

		if contentType != deviceResource.Properties.Value.MediaType {
			handler.logger.Error(fmt.Sprintf("Incoming reading ignored. Content Type '%s' doesn't match %s resource's Media Type '%s'",
				contentType, resourceName, deviceResource.Properties.Value.MediaType))

			http.Error(writer, "Wrong Content-Type", http.StatusBadRequest)
			return
		}
	}

	var reading interface{}
	readingType := models.ParseValueType(deviceResource.Properties.Value.Type)

	if readingType == models.Binary {
		reading, err = handler.readBodyAsBinary(writer, request)

	} else {
		reading, err = handler.readBodyAsString(writer, request)
	}

	if err != nil {
		handler.logger.Error(fmt.Sprintf("Incoming reading ignored. Unable to read request body: %s", err.Error()))
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	value, err := handler.newCommandValue(resourceName, reading, readingType)
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

	handler.processAsyncRequest(writer, request)
}

func (handler StorageHandler) newCommandValue(resourceName string, reading interface{}, valueType models.ValueType) (*models.CommandValue, error) {
	var result = &models.CommandValue{}
	var errn error
	var timestamp = time.Now().UnixNano()
	castError := "fail to parse %v reading, %v"

	if !checkValueInRange(valueType, reading) {
		errn = fmt.Errorf("parse reading fail. Reading %v is out of the value type(%v)'s range", reading, valueType)
		handler.logger.Error(errn.Error())
		return result, errn
	}

	switch valueType {
	case models.Binary:
		val, ok := reading.([]byte)
		if !ok {
			return nil, fmt.Errorf(castError, resourceName, "not []byte")
		}
		result, errn = models.NewCommandValue(resourceName, timestamp, val, valueType)

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
	default:
		errn = fmt.Errorf("return result fail, none supported value type: %v", valueType)
	}

	return result, errn
}

func checkValueInRange(valueType models.ValueType, reading interface{}) bool {
	isValid := false

	if valueType == models.String || valueType == models.Bool || valueType == models.Binary {
		return true
	} else if valueType == models.Int8 || valueType == models.Int16 ||
		valueType == models.Int32 || valueType == models.Int64 {
		val := cast.ToInt64(reading)
		isValid = checkIntValueRange(valueType, val)
	} else if valueType == models.Uint8 || valueType == models.Uint16 ||
		valueType == models.Uint32 || valueType == models.Uint64 {
		val := cast.ToUint64(reading)
		isValid = checkUintValueRange(valueType, val)
	}
	return isValid
}

func checkUintValueRange(valueType models.ValueType, val uint64) bool {
	var isValid = false
	switch valueType {
	case models.Uint8:
		if val >= 0 && val <= math.MaxUint8 {
			isValid = true
		}
	case models.Uint16:
		if val >= 0 && val <= math.MaxUint16 {
			isValid = true
		}
	case models.Uint32:
		if val >= 0 && val <= math.MaxUint32 {
			isValid = true
		}
	case models.Uint64:
		maxiMum := uint64(math.MaxUint64)
		if val >= 0 && val <= maxiMum {
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
		if val >= math.MinInt64 && val <= math.MaxInt64 {
			isValid = true
		}
	}
	return isValid
}
