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

package application

import (
	"encoding/json"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/errors"
	"reflect"
	"testing"

	wrapperMock "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/wrapper/mocks"

	"github.com/golang/mock/gomock"
)

const (
	validName          = "valid"
	invalidName        = "invalid"
	executableFileName = "test_file"
	execType           = "container"

	serviceJSON = "{\"serviceName\":\"valid\",\"executableFileName\":\"test_file\",\"allowedRequester\":[\"test_req\"],\"execType\":\"container\",\"execCmd\":[\"docker\",\"run\",\"valid\"]}"
)

var (
	notFoundErr = errors.NotFound{Message: invalidName + " does not exist"}
	dbOPErr     = errors.DBOperationError{}

	allowedRequester = []string{"test_req"}
	execCmd          = []string{"docker", "run", "valid"}

	serviceStruct = Info{
		ServiceName:        validName,
		ExecutableFileName: executableFileName,
		AllowedRequester:   allowedRequester,
		ExecType:           execType,
		ExecCmd:            execCmd,
	}
)

func TestGet_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Get([]byte(validName)).Return([]byte(serviceJSON), nil),
	)

	db = wrapperMockObj
	query := Query{}

	data, err := query.Get(validName)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(serviceStruct, data) {
		t.Error("Expected res: ", serviceStruct, "actual res: %s", data)
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

func TestGetList_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	serviceInfoMap := map[string]interface{}{
		validName: serviceJSON,
	}
	serviceInfoStructList := []Info{serviceStruct}

	gomock.InOrder(
		wrapperMockObj.EXPECT().List().Return(serviceInfoMap, nil),
	)

	db = wrapperMockObj
	query := Query{}

	data, err := query.GetList()
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(serviceInfoStructList, data) {
		t.Error("Expected res: ", serviceInfoStructList, "actual res: ", data)
	}
}

func TestGetList_WhenDBReturnError_ExpectedErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().List().Return(nil, notFoundErr),
	)

	db = wrapperMockObj
	query := Query{}

	_, err := query.GetList()
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

	serviceByte, _ := json.Marshal(serviceStruct)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Put([]byte(serviceStruct.ServiceName), serviceByte).Return(nil),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Set(serviceStruct)
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

	err := query.Set(Info{})
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
