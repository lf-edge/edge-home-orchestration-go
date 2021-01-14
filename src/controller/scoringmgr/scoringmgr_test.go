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
	"testing"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/resourceutil"
	resourceUtilMock "github.com/lf-edge/edge-home-orchestration-go/src/common/resourceutil/mocks"

	"github.com/golang/mock/gomock"
)

var (
	dummyDevID    = "devID"
	expectedScore = 0.5948754760361981
)

func TestGetScore_ExpectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceutilMockObj := resourceUtilMock.NewMockGetResource(ctrl)

	gomock.InOrder(
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(10.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(10.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUFreq).Return(10.0, nil),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.NetBandwidth).Return(10.0, nil),
		resourceutilMockObj.EXPECT().SetDeviceID(dummyDevID),
		resourceutilMockObj.EXPECT().GetResource(resourceutil.NetRTT).Return(10.0, nil),
	)
	resourceIns = resourceutilMockObj

	score, err := GetInstance().GetScore(dummyDevID)
	if err != nil {
		t.Errorf("Unexpected error return : %s", err.Error())
	}
	if score != expectedScore {
		t.Error("score : ", score, " expectedScore : ", expectedScore)
	}
}
