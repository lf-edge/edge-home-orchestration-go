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

package container

import (
	"common/errors"
	"common/resourceutil"
	resourceUtilMock "common/resourceutil/mocks"
	"testing"

	"github.com/golang/mock/gomock"
)

var getter resourceutil.Command

var (
	dummyError           = errors.SystemError{Message: "dummyError"}
	expectedSuccessScore = float64(0.59487547603619805)
	expectedFailScore    = float64(0.0)
)

func init() {
	getter = Getter{}
}

func TestRun_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceutilMockObj := resourceUtilMock.NewMockGetResource(ctrl)

	resourceIns = resourceutilMockObj

	gomock.InOrder(
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(10.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(10.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUFreq).Return(10.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.NetBandwidth).Return(10.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.NetRTT).Return(10.0, nil),
	)

	score := getter.Run()
	if score != expectedSuccessScore {
		t.Errorf("Unexpected score : %f", score)
	}
}

func TestRun_GetResourceReturnError_WithCPUUsage_ExpectedReturnZeroScore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceutilMockObj := resourceUtilMock.NewMockGetResource(ctrl)

	resourceIns = resourceutilMockObj

	gomock.InOrder(
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(0.0, dummyError),
	)

	score := getter.Run()
	if score != expectedFailScore {
		t.Errorf("Unexpected score : %f", score)
	}
}

func TestRun_GetResourceReturnError_WithCPUCount_ExpectedReturnZeroScore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceutilMockObj := resourceUtilMock.NewMockGetResource(ctrl)

	resourceIns = resourceutilMockObj

	gomock.InOrder(
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(0.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(0.0, dummyError),
	)

	score := getter.Run()
	if score != expectedFailScore {
		t.Errorf("Unexpected score : %f", score)
	}
}

func TestRun_GetResourceReturnError_WithCPUFreq_ExpectedReturnZeroScore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceutilMockObj := resourceUtilMock.NewMockGetResource(ctrl)

	resourceIns = resourceutilMockObj

	gomock.InOrder(
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(0.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(0.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUFreq).Return(0.0, dummyError),
	)

	score := getter.Run()
	if score != expectedFailScore {
		t.Errorf("Unexpected score : %f", score)
	}
}

func TestRun_GetResourceReturnError_WithNetBandwidth_ExpectedReturnZeroScore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceutilMockObj := resourceUtilMock.NewMockGetResource(ctrl)

	resourceIns = resourceutilMockObj

	gomock.InOrder(
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(0.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(0.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUFreq).Return(0.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.NetBandwidth).Return(0.0, dummyError),
	)

	score := getter.Run()
	if score != expectedFailScore {
		t.Errorf("Unexpected score : %f", score)
	}
}

func TestRun_GetResourceReturnError_WithNetRTT_ExpectedReturnZeroScore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceutilMockObj := resourceUtilMock.NewMockGetResource(ctrl)

	resourceIns = resourceutilMockObj

	gomock.InOrder(
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(0.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(0.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUFreq).Return(0.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.NetBandwidth).Return(0.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.NetRTT).Return(0.0, dummyError),
	)

	score := getter.Run()
	if score != expectedFailScore {
		t.Errorf("Unexpected score : %f", score)
	}
}
