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
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"restinterface"
)

const (
	// ConstWellknownPort is the common port for REST API
	ConstWellknownPort = 56001
)

// RestRouter struct {
type RestRouter struct {
	routes restinterface.Routes
	router *mux.Router
}

// NewRestRouter constructs RestRouter instance
func NewRestRouter() *RestRouter {

	edgeRouter := new(RestRouter)
	edgeRouter.router = mux.NewRouter().StrictSlash(true)

	return edgeRouter
}

// Add registers REST API to RestRouter
func (r *RestRouter) Add(s restinterface.IRestRoutes) {
	r.add(s.GetRoutes())
}

// Start wraps ListenAndServe function
func (r RestRouter) Start() {
	go r.listenAndServe()
}

func (r RestRouter) listenAndServe() {
	log.Printf("ListenAndServe")
	http.ListenAndServe(":"+strconv.Itoa(ConstWellknownPort), r.router)
}

func (r RestRouter) add(routes restinterface.Routes) {
	for _, route := range routes {
		handler := logger(route.HandlerFunc, route.Name)

		log.Printf("%v", route)

		r.router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
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
