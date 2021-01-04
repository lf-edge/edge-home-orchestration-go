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

package network

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
	rtt       = 0.0
	rtt2      = 0.1

	netJSON = "{\"id\":\"valid_id\",\"IPv4\":[\"192.168.0.1\"],\"RTT\":0.0}"
)

var (
	ipv4List    = []string{"192.168.0.1"}
	ipv4List2   = []string{"192.168.0.1", "192.168.0.2"}
	notFoundErr = errors.NotFound{Message: invalidID + " does not exist"}
	dbOPErr     = errors.DBOperationError{}

	netStruct = NetworkInfo{
		ID:   validID,
		IPv4: ipv4List,
		RTT:  rtt,
	}

	netStruct2 = NetworkInfo{
		ID:   validID,
		IPv4: ipv4List2,
		RTT:  rtt2,
	}

	netinfo = map[string]interface{}{
		"id":   validID,
		"IPv4": ipv4List,
		"RTT":  rtt,
	}
)

func TestGet_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Get([]byte(validID)).Return([]byte(netJSON), nil),
	)

	db = wrapperMockObj
	query := Query{}

	data, err := query.Get(validID)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(netStruct, data) {
		t.Error("Expected res: ", netStruct, "actual res: %s", data)
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

	netInfoMap := map[string]interface{}{
		validID: netJSON,
	}
	netInfoStructList := []NetworkInfo{netStruct}

	gomock.InOrder(
		wrapperMockObj.EXPECT().List().Return(netInfoMap, nil),
	)

	db = wrapperMockObj
	query := Query{}

	data, err := query.GetList()
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(netInfoStructList, data) {
		t.Error("Expected res: ", netInfoStructList, "actual res: ", data)
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

	netByte, _ := json.Marshal(netStruct)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Put([]byte(netStruct.ID), netByte).Return(nil),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Set(netStruct)
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

	err := query.Set(NetworkInfo{})
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

	updatedNetInfoByte, _ := json.Marshal(netStruct2)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Get([]byte(validID)).Return([]byte(netJSON), nil),
		wrapperMockObj.EXPECT().Put([]byte(netStruct2.ID), updatedNetInfoByte).Return(nil),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Update(netStruct2)
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

	err := query.Update(netStruct2)
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
