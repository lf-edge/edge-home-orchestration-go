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

// Package route implements management functions of REST Router
package route

import (
	"net/http"
	"strconv"
	"time"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"

	"github.com/gorilla/mux"

	"github.com/lf-edge/edge-home-orchestration-go/src/controller/securemgr/authenticator"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/externalhandler"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/internalhandler"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/route/tlspskserver"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/tls"
)

const (
	// ConstWellknownPort is the common port for REST API
	ConstWellknownPort = 56001
	ConstInternalPort  = 56002
)

var (
	log = logmgr.GetInstance()
)

// RestRouter struct {
type RestRouter struct {
	routes         restinterface.Routes
	routerInternal *mux.Router
	routerExternal *mux.Router

	tls.HasCertificate
}

// NewRestRouter constructs RestRouter instance
func NewRestRouter() *RestRouter {

	edgeRouter := new(RestRouter)
	edgeRouter.routerExternal = nil
	edgeRouter.routerInternal = nil

	return edgeRouter
}

// NewRestRouter constructs RestRouter instance with Certificate file path
func NewRestRouterWithCerti(path string) *RestRouter {
	edgeRouter := NewRestRouter()
	edgeRouter.SetCertificateFilePath(path)

	return edgeRouter
}

// Add registers REST API to RestRouter
func (r *RestRouter) Add(s restinterface.IRestRoutes) {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(authenticator.IsAuthorizedRequest)

	for _, route := range s.GetRoutes() {
		handler := logger(route.HandlerFunc, route.Name)

		log.Printf("%v", route)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	switch s.(type) {
	case *internalhandler.Handler:
		r.routerInternal = router
	case *externalhandler.Handler:
		r.routerExternal = router
	default:
		log.Println("added unknown type, ignore")
	}
}

// Start wraps ListenAndServe function
func (r RestRouter) Start() {
	r.listenAndServe()
}

func (r RestRouter) listenAndServe() {
	// start internal server
	switch r.IsSetCert {
	case true:
		log.Printf("ListenAndServeTLS_For_Inter")
		go tlspskserver.TLSPSKServer{}.ListenAndServe(":"+strconv.Itoa(ConstInternalPort), r.routerInternal)
	default:
		log.Printf("ListenAndServe_For_Inter")
		go http.ListenAndServe(":"+strconv.Itoa(ConstInternalPort), r.routerInternal)
	}

	if log.Printf("ListenAndServe"); r.routerExternal != nil {
		go http.ListenAndServe(":"+strconv.Itoa(ConstWellknownPort), r.routerExternal)
	}
}

func logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		if name != "APIV1Ping" {
			log.Printf(
				"From [%s] %s %s %s %s",
				readClientIP(r),
				r.Method,
				r.RequestURI,
				name,
				time.Since(start),
			)
		}
	})
}

func readClientIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}
