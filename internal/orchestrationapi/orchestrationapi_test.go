/*******************************************************************************
 * Copyright 2019-2020 Samsung Electronics All Rights Reserved.
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

/**************** HOW TO USE ********************
$ source workspaceProfile.sh
$ go test -failfast -v -count=1 orchestrationapi
*************************************************/

package orchestrationapi

import (
	"errors"

	"github.com/golang/mock/gomock"

	"testing"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/requestervalidator"
	sysDB "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/system"
	dbhelper "github.com/lf-edge/edge-home-orchestration-go/internal/db/helper"
)

func TestRequestService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	appName := "MyApp"
	args := []string{"a", "-b", "-c"}

	var requestServiceInfo ReqeustService
	requestServiceInfo.ServiceName = appName
	requestServiceInfo.SelfSelection = true
	requestServiceInfo.ServiceRequester = "my"
	requestServiceInfo.ServiceInfo = []RequestServiceInfo{
		{
			ExecutionType: "platform",
			ExeCmd:        args,
		},
	}
	requestervalidator.RequesterValidator{}.StoreRequesterInfo(appName, []string{"my"})

	candidateInfos := make([]dbhelper.ExecutionCandidate, 0)
	candidateInfos = append(candidateInfos, dbhelper.ExecutionCandidate{
		Id:       "ID1",
		ExecType: "platform",
		Endpoint: []string{"endpoint1"},
	})
	candidateInfos = append(candidateInfos, dbhelper.ExecutionCandidate{
		Id:       "ID2",
		ExecType: "platform",
		Endpoint: []string{"endpoint2"},
	})
	candidateInfos = append(candidateInfos, dbhelper.ExecutionCandidate{
		Id:       "ID3",
		ExecType: "platform",
		Endpoint: []string{"endpoint3"},
	})

	sysInfo := sysDB.SystemInfo{
		Name:  "ID",
		Value: "ID",
	}

	t.Run("Success", func(t *testing.T) {
		scores := []float64{float64(1.0), float64(2.0), float64(3.0)}

		gomock.InOrder(
			mockService.EXPECT().SetLocalServiceExecutor(mockExecutor),
			mockDiscovery.EXPECT().SetRestResource(),
			mockDBHelper.EXPECT().GetDeviceInfoWithService(gomock.Eq(appName), gomock.Any()).Return(candidateInfos, nil),
			mockSystemDBExecutor.EXPECT().Get("id").Return(sysInfo, nil),
			mockNetwork.EXPECT().GetIPs().Return([]string{""}, nil),
			mockClient.EXPECT().DoGetScoreRemoteDevice(gomock.Any(), gomock.Any()).Return(scores[0], nil),
			mockClient.EXPECT().DoGetScoreRemoteDevice(gomock.Any(), gomock.Any()).Return(scores[1], nil),
			mockClient.EXPECT().DoGetScoreRemoteDevice(gomock.Any(), gomock.Any()).Return(scores[2], nil),
			mockNetwork.EXPECT().GetIPs().Return([]string{""}, nil),
			mockService.EXPECT().Execute(gomock.Any(), appName, gomock.Any(), gomock.Any(), gomock.Any()),
		)

		o := getOcheIns(ctrl)
		if o == nil {
			t.Error("ochestration object is nil, expected is not nil")
		}

		oche := getOrcheImple()

		oche.Ready = true

		res := oche.RequestService(requestServiceInfo)
		if res.Message != ErrorNone {
			t.Error("unexpected handle")
		}
	})

	t.Run("Error", func(t *testing.T) {
		t.Run("NotReady", func(t *testing.T) {
			mockService.EXPECT().SetLocalServiceExecutor(mockExecutor)
			mockDiscovery.EXPECT().SetRestResource()
			o := getOcheIns(ctrl)
			if o == nil {
				t.Error("ochestration object is nil, expected is not nil")
			}

			oche := getOrcheImple()
			res := oche.RequestService(requestServiceInfo)
			if res.Message != InternalServerError {
				t.Error("unexpected Error")
			}
		})
		t.Run("DiscoveryFail", func(t *testing.T) {
			gomock.InOrder(
				mockService.EXPECT().SetLocalServiceExecutor(mockExecutor),
				mockDiscovery.EXPECT().SetRestResource(),
				mockDBHelper.EXPECT().GetDeviceInfoWithService(gomock.Eq(appName), gomock.Any()).Return(nil, errors.New("-3")),
			)
			o := getOcheIns(ctrl)
			if o == nil {
				t.Error("ochestration object is nil, expected is not nil")
			}

			oche := getOrcheImple()

			oche.Ready = true

			res := oche.RequestService(requestServiceInfo)
			if res.Message == ErrorNone {
				t.Error("unexpected Error")
			}
		})
	})
}
