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
	cryptotls "crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"restinterface"
	"restinterface/tls"
)

const (
	// ConstWellknownPort is the common port for REST API
	ConstWellknownPort = 56001
)

// RestRouter struct {
type RestRouter struct {
	routes restinterface.Routes
	router *mux.Router

	tls.HasCertificate
}

// NewRestRouter constructs RestRouter instance
func NewRestRouter() *RestRouter {

	edgeRouter := new(RestRouter)
	edgeRouter.router = mux.NewRouter().StrictSlash(true)

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
	r.add(s.GetRoutes())
}

// Start wraps ListenAndServe function
func (r RestRouter) Start() {
	go r.listenAndServe()
}

func (r RestRouter) listenAndServe() {
	switch r.IsSetCert {
	case true:
		log.Printf("ListenAndServeTLS")
		caCert, err := ioutil.ReadFile(r.GetCertificateFilePath() + "/" + tls.CertificateFileName)
		if err != nil {
			log.Println("cert file read fail, run http")
			http.ListenAndServe(":"+strconv.Itoa(ConstWellknownPort), r.router)
		} else {
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			cfg := &cryptotls.Config{
				ClientAuth:         cryptotls.RequireAndVerifyClientCert,
				ClientCAs:          caCertPool,
				RootCAs:            caCertPool,
				InsecureSkipVerify: true,
			}
			srv := &http.Server{
				Addr:      ":" + strconv.Itoa(ConstWellknownPort),
				Handler:   r.router,
				TLSConfig: cfg,
			}
			log.Fatal(srv.ListenAndServeTLS(
				r.GetCertificateFilePath()+"/"+tls.CertificateFileName,
				r.GetCertificateFilePath()+"/"+tls.KeyFileName,
			))
		}
	default:
		log.Printf("ListenAndServe")
		http.ListenAndServe(":"+strconv.Itoa(ConstWellknownPort), r.router)

	}
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

		log.Printf(
			"From [%s] %s %s %s %s",
			readClientIP(r),
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
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
