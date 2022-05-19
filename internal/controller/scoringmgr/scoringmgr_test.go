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
	"errors"
	"testing"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/resourceutil"
	resourceUtilMock "github.com/lf-edge/edge-home-orchestration-go/internal/common/resourceutil/mocks"

	"github.com/golang/mock/gomock"
)

const (
	unexpectedSuccess = "unexpected success"
	unexpectedFail    = "unexpected fail"
)

var (
	dummyDevID    = "devID"
	expectedScore = 0.5948754760361981
)

func TestGetScore_ExpectedSuccess(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
			t.Error(unexpectedFail, err.Error())
		}
		if score != expectedScore {
			t.Error(unexpectedFail, "score : ", score, " expectedScore : ", expectedScore)
		}
	})
}

func TestGetResource(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
		_, err := GetInstance().GetResource(dummyDevID)
		if err != nil {
			t.Error(unexpectedFail, err.Error())
		}
	})
	t.Run("Fail", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resourceutilMockObj := resourceUtilMock.NewMockGetResource(ctrl)

		t.Run("CPUUsage", func(t *testing.T) {
			gomock.InOrder(
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(10.0, errors.New("")),
			)
			resourceIns = resourceutilMockObj
			_, err := GetInstance().GetResource(dummyDevID)
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
		t.Run("CPUCount", func(t *testing.T) {
			gomock.InOrder(
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(10.0, errors.New("")),
			)
			resourceIns = resourceutilMockObj
			_, err := GetInstance().GetResource(dummyDevID)
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
		t.Run("CPUFreq", func(t *testing.T) {
			gomock.InOrder(
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUFreq).Return(10.0, errors.New("")),
			)
			resourceIns = resourceutilMockObj
			_, err := GetInstance().GetResource(dummyDevID)
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
		t.Run("NetBandwidth", func(t *testing.T) {
			gomock.InOrder(
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUFreq).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.NetBandwidth).Return(10.0, errors.New("")),
			)
			resourceIns = resourceutilMockObj
			_, err := GetInstance().GetResource(dummyDevID)
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
		t.Run("NetRTT", func(t *testing.T) {
			gomock.InOrder(
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUFreq).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.NetBandwidth).Return(10.0, nil),
				resourceutilMockObj.EXPECT().SetDeviceID(dummyDevID),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.NetRTT).Return(10.0, errors.New("")),
			)
			resourceIns = resourceutilMockObj
			_, err := GetInstance().GetResource(dummyDevID)
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})

	})
}

func TestCalculateScore(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
		score := calculateScore(dummyDevID)
		if score != expectedScore {
			t.Error(unexpectedFail, "score : ", score, " expectedScore : ", expectedScore)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resourceutilMockObj := resourceUtilMock.NewMockGetResource(ctrl)

		t.Run("CPUUsage", func(t *testing.T) {
			gomock.InOrder(
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(10.0, errors.New("")),
			)
			resourceIns = resourceutilMockObj
			score := calculateScore(dummyDevID)
			if score != InvalidScore {
				t.Error(unexpectedSuccess, "score : ", score, " expectedScore : ", InvalidScore)
			}
		})
		t.Run("CPUCount", func(t *testing.T) {
			gomock.InOrder(
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(10.0, errors.New("")),
			)
			resourceIns = resourceutilMockObj
			score := calculateScore(dummyDevID)
			if score != InvalidScore {
				t.Error(unexpectedSuccess, "score : ", score, " expectedScore : ", InvalidScore)
			}
		})
		t.Run("CPUFreq", func(t *testing.T) {
			gomock.InOrder(
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUFreq).Return(10.0, errors.New("")),
			)
			resourceIns = resourceutilMockObj
			score := calculateScore(dummyDevID)
			if score != InvalidScore {
				t.Error(unexpectedSuccess, "score : ", score, " expectedScore : ", InvalidScore)
			}
		})
		t.Run("NetBandwidth", func(t *testing.T) {
			gomock.InOrder(
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUFreq).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.NetBandwidth).Return(10.0, errors.New("")),
			)
			resourceIns = resourceutilMockObj
			score := calculateScore(dummyDevID)
			if score != InvalidScore {
				t.Error(unexpectedSuccess, "score : ", score, " expectedScore : ", InvalidScore)
			}
		})
		t.Run("NetRTT", func(t *testing.T) {
			gomock.InOrder(
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUUsage).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUCount).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.CPUFreq).Return(10.0, nil),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.NetBandwidth).Return(10.0, nil),
				resourceutilMockObj.EXPECT().SetDeviceID(dummyDevID),
				resourceutilMockObj.EXPECT().GetResource(resourceutil.NetRTT).Return(10.0, errors.New("")),
			)
			resourceIns = resourceutilMockObj
			score := calculateScore(dummyDevID)
			if score != InvalidScore {
				t.Error(unexpectedSuccess, "score : ", score, " expectedScore : ", InvalidScore)
			}
		})
	})
}

func TestGetScoreWithResource(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resource := make(map[string]interface{})
		resource["cpuUsage"] = 10.0
		resource["cpuCount"] = 10.0
		resource["cpuFreq"] = 10.0
		resource["netBandwidth"] = 10.0
		resource["rtt"] = 10.0
		_, err := GetInstance().GetScoreWithResource(resource)
		if err != nil {
			t.Error(unexpectedFail, err.Error())
		}
	})
	t.Run("Fail", func(t *testing.T) {
		resource := make(map[string]interface{})
		resource["error"] = 10.0
		_, err := GetInstance().GetScoreWithResource(resource)
		if err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestRenderingScore(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		score := renderingScore(-0.1)
		if score != 0 {
			t.Error(unexpectedFail, "score : ", score, " expectedScore : 0")
		}

	})
}
