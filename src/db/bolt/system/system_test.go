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

package system

import (
	"github.com/lf-edge/edge-home-orchestration-go/src/common/errors"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	wrapperMock "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/wrapper/mocks"

	"github.com/golang/mock/gomock"
)

const (
	invalidKey = "invalid_key"
)

var (
	defaultID = "defaultID"

	notFoundErr = errors.NotFound{Message: invalidKey + " does not exist"}
	dbOPErr     = errors.DBOperationError{}

	idInfo = SystemInfo{
		Name:  ID,
		Value: defaultID,
	}
)

func TestGet_WithVaildIDKey_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	idBytes, _ := json.Marshal(idInfo)
	fmt.Println(idBytes)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Get([]byte(ID)).Return(idBytes, nil),
	)

	db = wrapperMockObj
	query := Query{}

	data, err := query.Get(ID)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(idInfo, data) {
		t.Error("Expected res: ", idInfo, "actual res: %s", data)
	}
}

func TestGet_WithInvalidKey_ExpectedErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Get([]byte(invalidKey)).Return(nil, notFoundErr),
	)

	db = wrapperMockObj
	query := Query{}

	_, err := query.Get(invalidKey)
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

	idBytes, _ := json.Marshal(idInfo)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Put([]byte(idInfo.Name), idBytes).Return(nil),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Set(idInfo)
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

	err := query.Set(SystemInfo{})
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
		wrapperMockObj.EXPECT().Delete([]byte(ID)).Return(nil),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Delete(ID)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestDelete_WithInvalidID_ExpectedErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wrapperMockObj := wrapperMock.NewMockDatabase(ctrl)

	gomock.InOrder(
		wrapperMockObj.EXPECT().Delete([]byte(invalidKey)).Return(notFoundErr),
	)

	db = wrapperMockObj
	query := Query{}

	err := query.Delete(invalidKey)
	if err == nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}
