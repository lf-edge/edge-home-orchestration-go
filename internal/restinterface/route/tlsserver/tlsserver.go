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
	"net"
	"net/http"
	"os"

	"crypto/tls"
	"crypto/x509"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
)

const (
	logPrefix = "[tlsserver] "
)

var (
	log = logmgr.GetInstance()
)

// TLSListenerServer provides insterface for tlsserver
type TLSListenerServer interface {
	ListenAndServe(addr string, handler http.Handler)
}

// TLSServer structure
type TLSServer struct {
	Certspath string
	listener  net.Listener
}

func createServerConfig(certspath string) (*tls.Config, error) {
	caCertPEM, err := os.ReadFile(certspath + "/ca-crt.pem")
	if err != nil {
		log.Panic(logPrefix, err.Error())
		return nil, err
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertPEM)
	if !ok {
		log.Panic(logPrefix, "failed to parse root certificate")
	}

	cert, err := tls.LoadX509KeyPair(certspath+"/hen-crt.pem", certspath+"/hen-key.pem")
	if err != nil {
		log.Panic(logPrefix, err.Error())
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
func (s *TLSServer) ListenAndServe(addr string, handler http.Handler) {

	config, _ := createServerConfig(s.Certspath)

	var err error
	s.listener, err = tls.Listen("tcp", addr, config)
	if err != nil {
		log.Panic(logPrefix, "listen failed: ", err.Error())
	}

	defer s.listener.Close()

	(&http.Server{Handler: handler}).Serve(s.listener)
}
