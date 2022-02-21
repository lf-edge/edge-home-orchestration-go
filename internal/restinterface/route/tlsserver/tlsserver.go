/*******************************************************************************
 * Copyright 2021 Samsung Electronics All Rights Reserved.
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

package tlsserver

import (
	"net/http"

	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
)

const (
	edgeDir = "/var/edge-orchestration"
	caCert  = edgeDir + "/certs/ca-crt.pem"
	henCert = edgeDir + "/certs/hen-crt.pem"
	henKey  = edgeDir + "/certs/hen-key.pem"
)

var (
	log = logmgr.GetInstance()
)

// TLSListenerServer provides insterface for tlsserver
type TLSListenerServer interface {
	ListenAndServe(addr string, handler http.Handler)
}

// TLSServer structure
type TLSServer struct{}

func createServerConfig() (*tls.Config, error) {
	caCertPEM, err := ioutil.ReadFile(caCert)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertPEM)
	if !ok {
		panic("failed to parse root certificate")
	}

	cert, err := tls.LoadX509KeyPair(henCert, henKey)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates:             []tls.Certificate{cert},
		ClientAuth:               tls.RequireAndVerifyClientCert,
		ClientCAs:                roots,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}, nil
}

// ListenAndServe listens HTTPS connection and calls Serve with handler to handle requests on incoming connections.
func (TLSServer) ListenAndServe(addr string, handler http.Handler) {

	config, err := createServerConfig()
	if err != nil {
		log.Fatal("config failed: ", err)
	}

	listener, err := tls.Listen("tcp", addr, config)
	if err != nil {
		log.Fatal("listen failed: ", err)
	}

	defer listener.Close()

	(&http.Server{Handler: handler}).Serve(listener)
}
