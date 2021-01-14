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

package configuration

import (
	"github.com/lf-edge/edge-home-orchestration-go/src/common/errors"
	"encoding/json"
	"reflect"
	"testing"

	wrapperMock "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/wrapper/mocks"

	"github.com/golang/mock/gomock"
)

const (
	validID        = "valid_id"
	invalidID      = "invalid_id"
	platform       = "test_platform"
	executionType  = "test_exec"
	platform2      = "test_platform2"
	executionType2 = "test_exec2"

	confJSON = "{\"id\":\"valid_id\",\"platform\":\"test_platform\",\"executionType\":\"test_exec\"}"
)

var (
	notFoundErr = errors.NotFound{Message: invalidID + " does not exist"}
	dbOPErr     = errors.DBOperationError{}

	confStruct = Configuration{
		ID:       validID,
		Platform: platform,
		ExecType: executionType,
	}

	confStruct2 = Configuration{
		ID:       validID,
		Platform: platform2,
		ExecType: executionType2,
	}

	conf = map[string]interface{}{
		"id":            validID,
		"platform":      platform,
		"executionType": executionType,
	}
)

func TestGet_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Get([]byte(validID)).Return([]byte(confJSON), nil),
	)

	db = wrapperMockObj
	query := Query{}

	data, err := query.Get(validID)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(confStruct, data) {
		t.Errorf("Expected res: %s, actual res: %s", confStruct, data)
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

	confMap := map[string]interface{}{
		validID: confJSON,
	}
	confStructList := []Configuration{confStruct}

	gomock.InOrder(
		wrapperMockObj.EXPECT().List().Return(confMap, nil),
	)

	db = wrapperMockObj
	query := Query{}

	data, err := query.GetList()
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(confStructList, data) {
		t.Errorf("Expected res: %s, actual res: %s", confStructList, data)
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

	confByte, _ := json.Marshal(confStruct)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Put([]byte(confStruct.ID), confByte).Return(nil),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Set(confStruct)
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

	err := query.Set(Configuration{})
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

	updatedConfByte, _ := json.Marshal(confStruct2)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Get([]byte(validID)).Return([]byte(confJSON), nil),
		wrapperMockObj.EXPECT().Put([]byte(confStruct.ID), updatedConfByte).Return(nil),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Update(confStruct2)
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

	err := query.Update(confStruct2)
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
