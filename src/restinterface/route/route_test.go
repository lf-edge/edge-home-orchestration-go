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
	"net/http/httptest"

	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/cipher"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/externalhandler"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/internalhandler"
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

	router := NewRestRouter()
	if router == nil {
		t.Error("unexpected return value")
	}

	type fakeRouter struct {
		restinterface.HasRoutes
		cipher.HasCipher
	}
	router.Add(&fakeRouter{})
	if router.routerInternal != nil || router.routerExternal != nil {
		t.Error("unexpected not set internal handler")
	}

	router.Add(internalhandler.GetHandler())
	if router.routerInternal == nil {
		t.Error("unexpected not set internal handler")
	}
	router.Add(externalhandler.GetHandler())
	if router.routerExternal == nil {
		t.Error("unexpected not set internal handler")
	}
}

func TestReadClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://0.0.0.0:12345", nil)
	req.RemoteAddr = "RemoteAddr"
	if readClientIP(req) != "RemoteAddr" {
		t.Error("unexpected value: ", readClientIP(req))
	}
	req.Header.Set("X-Forwarded-For", "X-Forwarded-For")
	if readClientIP(req) != "X-Forwarded-For" {
		t.Error("unexpected value: ", readClientIP(req))
	}
	req.Header.Set("X-Real-Ip", "X-Real-Ip")
	if readClientIP(req) != "X-Real-Ip" {
		t.Error("unexpected value: ", readClientIP(req))
	}
}

func TestNewRestRouterWithCerti(t *testing.T) {
	edgeRoute := NewRestRouterWithCerti("test")
	if edgeRoute.IsSetCert != true {
		t.Error("expected certificate is set, but not set")
	}
}
