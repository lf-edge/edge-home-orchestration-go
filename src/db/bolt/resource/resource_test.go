/*******************************************************************************
 * Copyright 2019 Samsung Electronics All Rights Reserved.
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

package resource

import (
	"github.com/lf-edge/edge-home-orchestration-go/src/common/errors"
	"encoding/json"
	"reflect"
	"testing"

	wrapperMock "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/wrapper/mocks"

	"github.com/golang/mock/gomock"
)

const (
	validName   = "validName"
	invalidName = "invalidName"
	value       = 0.0
	value2      = 0.1

	resourceJSON = "{\"name\":\"validName\",\"value\":0.0}"
)

var (
	ipv4List    = []string{"192.168.0.1"}
	ipv4List2   = []string{"192.168.0.1", "192.168.0.2"}
	notFoundErr = errors.NotFound{Message: invalidName + " does not exist"}
	dbOPErr     = errors.DBOperationError{}

	resourceStruct = ResourceInfo{
		Name:  validName,
		Value: value,
	}

	resourceStruct2 = ResourceInfo{
		Name:  validName,
		Value: value2,
	}

	resourceInfo = map[string]interface{}{
		"name":  validName,
		"value": ipv4List,
	}
)

func TestGet_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Get([]byte(validName)).Return([]byte(resourceJSON), nil),
	)

	db = wrapperMockObj
	query := Query{}

	data, err := query.Get(validName)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(resourceStruct, data) {
		t.Error("Expected res: ", resourceStruct, "actual res: %s", data)
	}
}

func TestGet_WithInvalidName_ExpectedErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Get([]byte(invalidName)).Return(nil, notFoundErr),
	)

	db = wrapperMockObj
	query := Query{}

	_, err := query.Get(invalidName)
	if err == nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}

func TestSet_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	resourceByte, _ := json.Marshal(resourceStruct)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Put([]byte(resourceStruct.Name), resourceByte).Return(nil),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Set(resourceStruct)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestSet_WhenDBReturnError_ExpectedErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Put(gomock.Any(), gomock.Any()).Return(dbOPErr),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Set(ResourceInfo{})
	if err == nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.DBOperationError:
	}
}

func TestDelete_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Delete([]byte(validName)).Return(nil),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Delete(validName)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestDelete_WithInvalidName_ExpectedErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Delete([]byte(invalidName)).Return(notFoundErr),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Delete(invalidName)
	if err == nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}
