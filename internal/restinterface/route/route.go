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
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/securemgr/authenticator"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/externalhandler"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/internalhandler"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/route/tlsserver"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/tls"
)

const (
	// ConstWellknownPort is the common port for REST API
	ConstWellknownPort = 56001
	// ConstInternalPort is the port used for TLS connection
	ConstInternalPort = 56002
	logPrefix         = "[route] "
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

	internalServer *http.Server
	externalServer *http.Server
}

// NewRestRouter constructs RestRouter instance
func NewRestRouter() *RestRouter {

	edgeRouter := new(RestRouter)
	edgeRouter.routerExternal = nil
	edgeRouter.routerInternal = nil

	return edgeRouter
}

// NewRestRouterWithCerti constructs RestRouter instance with Certificate file path
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
		log.Info(logPrefix, "added unknown type, ignore")
	}
}

// Start wraps ListenAndServe function
func (r RestRouter) Start() {
	r.listenAndServe()
}

// Stop shuts down both internal and external servers
func (r RestRouter) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if r.internalServer != nil {
		if err := r.internalServer.Shutdown(ctx); err != nil {
			log.Error(logPrefix, "Failed to shut down internal server")
		}
	}
	if r.externalServer != nil {
		if err := r.externalServer.Shutdown(ctx); err != nil {
			log.Error(logPrefix, "Failed to shut down external server")
		}
	}

}

func (r RestRouter) listenAndServe() {
	// start internal server
	switch r.IsSetCert {
	case true:
		log.Info(logPrefix, "Internal ListenAndServeTLS")
		s := tlsserver.TLSServer{Certspath: r.GetCertificateFilePath()}
		r.internalServer = &http.Server{
			Addr:    ":" + strconv.Itoa(ConstInternalPort),
			Handler: r.routerInternal,
		}
		go s.ListenAndServe(r.internalServer.Addr, r.internalServer.Handler)
		// go s.ListenAndServe(":"+strconv.Itoa(ConstInternalPort), r.routerInternal)
	default:
		log.Info(logPrefix, "Internal ListenAndServe")
		r.internalServer = &http.Server{
			Addr:    ":" + strconv.Itoa(ConstInternalPort),
			Handler: r.routerInternal,
		}
		go r.internalServer.ListenAndServe()
		// go http.ListenAndServe(":"+strconv.Itoa(ConstInternalPort), r.routerInternal)
	}

	if log.Info(logPrefix, "External ListenAndServe"); r.routerExternal != nil {
		r.externalServer = &http.Server{
			Addr:    ":" + strconv.Itoa(ConstWellknownPort),
			Handler: r.routerExternal,
		}
		go r.externalServer.ListenAndServe()
		// go http.ListenAndServe(":"+strconv.Itoa(ConstWellknownPort), r.routerExternal)
	}
}

func logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		if name != "APIV1Ping" {
			log.Printf("From [%s] %s %s %s %s", logmgr.SanitizeUserInput(readClientIP(r)), r.Method, r.RequestURI, name, time.Since(start)) // lgtm [go/log-injection]
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
