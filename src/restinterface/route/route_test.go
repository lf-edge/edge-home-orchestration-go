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

package route

import (
	"testing"

	"github.com/golang/mock/gomock"

	"net/http"

	"restinterface"
	routemock "restinterface/mocks"
)

func TestNewRestRouter(t *testing.T) {
	router := NewRestRouter()
	if router == nil {
		t.Error("unexpected return value")
	}
}

func getTestRoutes(t *testing.T) restinterface.Routes {
	route1handler := func(w http.ResponseWriter, r *http.Request) {}
	route2handler := func(w http.ResponseWriter, r *http.Request) {}

	return restinterface.Routes{
		restinterface.Route{Name: "route1", Method: "GET", Pattern: "/api/v1/route1", HandlerFunc: route1handler},
		restinterface.Route{Name: "route2", Method: "POST", Pattern: "/api/v1/route2", HandlerFunc: route2handler},
	}
}

func TestAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoute := routemock.NewMockIRestRoutes(ctrl)

	mockRoute.EXPECT().GetRoutes().Return(getTestRoutes(t))

	router := NewRestRouter()
	if router == nil {
		t.Error("unexpected return value")
	}
	router.Add(mockRoute)
}

// TODO check to call expected function as restapi using httpserver mock
