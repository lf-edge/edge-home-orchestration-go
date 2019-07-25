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

package scoringmgr

import (
	"fmt"
	"testing"
	"time"

	resourceUtilMock "common/resourceutil/mocks"
	types "common/types/configuremgrtypes"

	"github.com/golang/mock/gomock"
)

var (
	dummyDevID         = "devID"
	dummyServiceName1  = "dummyServiceName1"
	dummyServiceName2  = "dummyServiceName2"
	invalidServiceName = "InvalidServiceName"

	expectedScore = 100.0

	dummyServiceInfo1 = types.ServiceInfo{
		IntervalTimeMs: 100,
		ServiceName:    dummyServiceName1,
	}

	dummyServiceInfo2 = types.ServiceInfo{
		IntervalTimeMs: 100,
		ServiceName:    dummyServiceName2,
	}

	gResourceutilMockObjService1 *resourceUtilMock.MockCommand
	gResourceutilMockObjService2 *resourceUtilMock.MockCommand
)

func TestAddScoring_ExpectedCallScoringFunc_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceutilMockObj := resourceUtilMock.NewMockCommand(ctrl)

	gResourceutilMockObjService1 = resourceutilMockObj

	dummyServiceInfo1.ScoringFunc = resourceutilMockObj

	resourceutilMockObj.EXPECT().Run(gomock.Any()).Return(100.0).AnyTimes()

	err := GetInstance().AddScoring(dummyServiceInfo1)
	if err != nil {
		t.Errorf("Unexpected error return : %s", err.Error())
	}
}

func TestAddScoring_WithDifferentServiceName_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceutilMockObj := resourceUtilMock.NewMockCommand(ctrl)

	gResourceutilMockObjService2 = resourceutilMockObj

	dummyServiceInfo2.ScoringFunc = resourceutilMockObj

	resourceutilMockObj.EXPECT().Run(gomock.Any()).Return(100.0).AnyTimes()

	err := GetInstance().AddScoring(dummyServiceInfo2)
	if err != nil {
		t.Errorf("Unexpected error return : %s", err.Error())
	}
}

func TestAddScoring_WithDuplicatedServiceName_ExpectedReturnError(t *testing.T) {
	fmt.Println(dummyServiceInfo1)
	err := GetInstance().AddScoring(dummyServiceInfo1)
	if err == nil {
		t.Errorf("Unexpected error return : %s", err.Error())
	}
}

func TestGetScore_ExpectedSuccess(t *testing.T) {
	score, err := GetInstance().GetScore(dummyDevID, dummyServiceName1)
	if err != nil {
		t.Errorf("Unexpected error return : %s", err.Error())
	}

	if score != expectedScore {
		t.Errorf("Unexpected score return : %f", score)
	}
}

func TestGetScore_WithInvalidServiceName_ExpectedSuccess(t *testing.T) {
	score, err := GetInstance().GetScore(dummyDevID, invalidServiceName)
	if err == nil {
		t.Errorf("Unexpected error return : %s", err.Error())
	}

	if score != 0.0 {
		t.Errorf("Unexpected score return : %f", score)
	}
}

func TestRemoveScoringWithInvalidParam_ExpectedReturnError(t *testing.T) {
	err := GetInstance().RemoveScoring(invalidServiceName)
	if err == nil {
		t.Errorf("Unexpected error return : %s", err.Error())
	}
}

func TestRemoveScoring_ExpectedSuccess(t *testing.T) {
	gomock.InOrder(
		gResourceutilMockObjService1.EXPECT().Run(gomock.Any()).Return(100.0).AnyTimes(),
		gResourceutilMockObjService1.EXPECT().Close(),
	)

	err := GetInstance().RemoveScoring(dummyServiceName1)
	if err != nil {
		t.Errorf("Unexpected error return : %s", err.Error())
	}

	time.Sleep(1 * time.Second)
}

func TestRemoveAllScoring_ExpectedSuccess(t *testing.T) {
	gomock.InOrder(
		gResourceutilMockObjService2.EXPECT().Run(gomock.Any()).Return(100.0).AnyTimes(),
		gResourceutilMockObjService2.EXPECT().Close(),
	)

	err := GetInstance().RemoveAllScoring()
	if err != nil {
		t.Errorf("Unexpected error return : %s", err.Error())
	}

	time.Sleep(1 * time.Second)
}
