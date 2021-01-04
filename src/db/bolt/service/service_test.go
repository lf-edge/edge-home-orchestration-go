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

package service

import (
	"github.com/lf-edge/edge-home-orchestration-go/src/common/errors"
	"encoding/json"
	"reflect"
	"testing"

	wrapperMock "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/wrapper/mocks"

	"github.com/golang/mock/gomock"
)

const (
	validID   = "valid_id"
	invalidID = "invalid_id"

	serviceJSON = "{\"id\":\"valid_id\",\"services\":[\"service\"],\"RTT\":0.0}"
)

var (
	serviceList  = []string{"service"}
	serviceList2 = []string{"service", "service2"}
	notFoundErr  = errors.NotFound{Message: invalidID + " does not exist"}
	dbOPErr      = errors.DBOperationError{}

	serviceStruct = ServiceInfo{
		ID:       validID,
		Services: serviceList,
	}

	serviceStruct2 = ServiceInfo{
		ID:       validID,
		Services: serviceList2,
	}

	serviceInfo = map[string]interface{}{
		"id":   validID,
		"IPv4": serviceList,
	}
)

func TestGet_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Get([]byte(validID)).Return([]byte(serviceJSON), nil),
	)

	db = wrapperMockObj
	query := Query{}

	data, err := query.Get(validID)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(serviceStruct, data) {
		t.Error("Expected res: ", serviceStruct, "actual res: %s", data)
	}
}

func TestGet_WithInvalidID_ExpectedErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Get([]byte(invalidID)).Return(nil, notFoundErr),
	)

	db = wrapperMockObj
	query := Query{}

	_, err := query.Get(invalidID)
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
		validID: serviceJSON,
	}
	serviceInfoStructList := []ServiceInfo{serviceStruct}

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
		wrapperMockObj.EXPECT().Put([]byte(serviceStruct.ID), serviceByte).Return(nil),
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

	err := query.Set(ServiceInfo{})
	if err == nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.DBOperationError:
	}
}

func TestUpdate_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	updatedServiceInfoByte, _ := json.Marshal(serviceStruct2)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Get([]byte(validID)).Return([]byte(serviceJSON), nil),
		wrapperMockObj.EXPECT().Put([]byte(serviceStruct2.ID), updatedServiceInfoByte).Return(nil),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Update(serviceStruct2)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestUpdate_WhenNotfoundMatchedConfWithID_ExpectedErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Get([]byte(validID)).Return(nil, notFoundErr),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Update(serviceStruct2)
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
		wrapperMockObj.EXPECT().Delete([]byte(validID)).Return(nil),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Delete(validID)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestDelete_WithInvalidID_ExpectedErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Delete([]byte(invalidID)).Return(notFoundErr),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Delete(invalidID)
	if err == nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}
