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

/**************** HOW TO USE ********************
$ source workspaceProfile.sh
$ go test -failfast -v -count=1 orchestrationapi
*************************************************/

package orchestrationapi

import (
	"errors"
	"sync"
	"testing"

	"common/errormsg"

	"github.com/golang/mock/gomock"
)

func TestRequestService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	appName := "MyApp"
	args := []string{"-a", "-b", "-c"}
	endpoints := []string{"a", "b", "c"}

	t.Run("Success", func(t *testing.T) {
		scores := []float64{float64(1.0), float64(2.0), float64(3.0)}

		gomock.InOrder(
			mockService.EXPECT().SetLocalServiceExecutor(mockExecutor),
			mockDiscovery.EXPECT().GetDeviceIPListWithService(gomock.Eq(appName)).Return(endpoints, nil),
			mockNetwork.EXPECT().GetOutboundIP().Return("", nil),
			mockClient.EXPECT().DoGetScoreRemoteDevice(gomock.Eq(appName), gomock.Any()).Return(scores[0], nil),
			mockClient.EXPECT().DoGetScoreRemoteDevice(gomock.Eq(appName), gomock.Any()).Return(scores[1], nil),
			mockClient.EXPECT().DoGetScoreRemoteDevice(gomock.Eq(appName), gomock.Any()).Return(scores[2], nil),
			mockService.EXPECT().Execute(gomock.Any(), appName, gomock.Any(), gomock.Any()),
		)

		o := getOcheIns(ctrl)
		if o == nil {
			t.Error("ochestration object is nil, expected is not nil")
		}

		oche := getOrcheImple()

		oche.Ready = true

		handle := oche.RequestService(appName, args)
		if handle == errormsg.ErrorNotReadyOrchestrationInit ||
			handle == errormsg.ErrorNoNetworkInterface {
			t.Error("unexpected handle")
		}
	})

	t.Run("Error", func(t *testing.T) {
		t.Run("NotReady", func(t *testing.T) {
			mockService.EXPECT().SetLocalServiceExecutor(mockExecutor)
			o := getOcheIns(ctrl)
			if o == nil {
				t.Error("ochestration object is nil, expected is not nil")
			}

			oche := getOrcheImple()

			if oche.RequestService(appName, args) != errormsg.ErrorNotReadyOrchestrationInit {
				t.Error("unexpected handle")
			}
		})
		t.Run("DiscoveryFail", func(t *testing.T) {
			gomock.InOrder(
				mockService.EXPECT().SetLocalServiceExecutor(mockExecutor),
				mockDiscovery.EXPECT().GetDeviceIPListWithService(gomock.Eq(appName)).Return(nil, errors.New("-3")),
			)
			o := getOcheIns(ctrl)
			if o == nil {
				t.Error("ochestration object is nil, expected is not nil")
			}

			oche := getOrcheImple()

			oche.Ready = true

			handle := oche.RequestService(appName, args)
			if handle != errormsg.ErrorNoNetworkInterface {
				t.Logf("expect error %d", errormsg.ErrorNoNetworkInterface)
				t.Logf("actual %d", handle)
				t.Error("unexpected handle")
			}
		})
	})
}

func TestGetEndpointDevices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("Success", func(t *testing.T) {
		gomock.InOrder(
			mockService.EXPECT().SetLocalServiceExecutor(mockExecutor),
			mockDiscovery.EXPECT().GetDeviceIPListWithService("TestAppName").Return(nil, nil),
		)

		o := getOcheIns(ctrl)
		if o == nil {
			t.Error("ochestration object is nil, expected is not nil")
		}

		oche := getOrcheImple()

		oche.getEndpointDevices("TestAppName")

	})
}

func TestGatheringDevicesScore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("Success", func(t *testing.T) {
		appName := "MyApp"
		endpoints := []string{"a", "b", "c"}
		expectedScore := []float64{float64(1.0), float64(2.0), float64(3.0)}

		gomock.InOrder(
			mockService.EXPECT().SetLocalServiceExecutor(mockExecutor),
			mockNetwork.EXPECT().GetOutboundIP().Return("", nil),
			mockClient.EXPECT().DoGetScoreRemoteDevice(gomock.Eq(appName), gomock.Any()).Return(expectedScore[0], nil),
			mockClient.EXPECT().DoGetScoreRemoteDevice(gomock.Eq(appName), gomock.Any()).Return(expectedScore[1], nil),
			mockClient.EXPECT().DoGetScoreRemoteDevice(gomock.Eq(appName), gomock.Any()).Return(expectedScore[2], nil),
		)

		o := getOcheIns(ctrl)
		if o == nil {
			t.Error("ochestration object is nil, expected is not nil")
		}

		oche := getOrcheImple()

		scores := oche.gatheringDevicesScore(endpoints, appName)
		if len(scores) != len(expectedScore) {
			t.Error("can not receive some scores")
		}
	})
	// TODO make test about error case after made error notifying
}

type matcher struct{ values []string }

func matching(values []string) gomock.Matcher {
	return &matcher{values}
}
func (m *matcher) Matches(x interface{}) bool {
	sx := (x.([]interface{}))
	for idx, value := range m.values {
		if value != sx[idx].(string) {
			return false
		}
	}
	return true
}

func (m *matcher) String() string {
	return "is same value, []string <-> []interfaces{}"
}

func TestExecuteApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createMockIns(ctrl)

	t.Run("Success", func(t *testing.T) {
		endpoint := "a"
		serviceName := "MyService"
		args := []string{"-1", "-2", "-3"}
		notiChan := make(chan string)

		gomock.InOrder(
			mockService.EXPECT().SetLocalServiceExecutor(mockExecutor),
			mockService.EXPECT().Execute(gomock.Eq(endpoint), gomock.Eq(serviceName), matching(args), gomock.Eq(notiChan)),
		)

		o := getOcheIns(ctrl)
		if o == nil {
			t.Error("ochestration object is nil, expected is not nil")
		}

		oche := getOrcheImple()

		oche.executeApp(endpoint, serviceName, args, notiChan)
	})
}

func TestAddServiceClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		args := []string{"-1", "-2", "-3"}
		appName := "MyApp"
		clientID := 123

		client := addServiceClient(clientID, appName, args)
		if client.appName != appName {
			t.Error("wrong app name")
		}
		for idx, arg := range client.args {
			if arg != args[idx] {
				t.Error("wrong arg")
			}
		}
	})
}

func TestSortByScore(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deviceScores := []deviceScore{{"a", float64(1.0)}, {"b", float64(2.0)}, {"c", float64(3.0)}}
		sortedScores := sortByScore(deviceScores)
		for idx, score := range sortedScores {
			if idx == 0 {
				continue
			}
			if score.score > sortedScores[idx-1].score {
				t.Error("it is not sorted")
			}
		}
	})
}

func TestListenNotify(t *testing.T) {
	//	old := os.Stdout
	//	r, w, _ := os.Pipe()
	//	os.Stdout = w

	client := new(orcheClient)
	client.appName = "MyApp"
	client.args = []string{"-a", "-b", "-c"}
	client.endSignal = make(chan bool)
	client.notiChan = make(chan string)

	var wait sync.WaitGroup
	wait.Add(1)
	go func() {
		client.listenNotify()
		wait.Done()
	}()
	client.notiChan <- "test"
	wait.Wait()

	// TODO capture stdout and do check the app status is changed.
	//	outC := make(chan string)
	//	go func() {
	//		var buf bytes.Buffer
	//		io.Copy(&buf, r)
	//		outC <- buf.String()
	//	}()
	//
	//	w.Close()
	//	os.Stdout = old
	//	out := <-outC
	//
	//	if strings.Contains(out, "[orchestrationapi] service status changed [appNames:MyApp][status:test]") != true {
	//		t.Log(out)
	//		t.Error("expect receive notification, but never received")
	//	}
}
