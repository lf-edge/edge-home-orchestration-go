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

package tlshelper

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	// config        *tls.Config
	wellKnownPort map[string]string
)

const (
	edgeDir = "/var/edge-orchestration"
	caCert  = edgeDir + "/certs/ca-crt.pem"
	henCert = edgeDir + "/certs/hen-crt.pem"
	henKey  = edgeDir + "/certs/hen-key.pem"
)

// TLSHelper struct
type TLSHelper struct{}

func init() {
	wellKnownPort = map[string]string{
		"http":  "80",
		"https": "443",
	}

}

func createClientConfig() (*tls.Config, error) {
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
		RootCAs:                  roots,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}, nil
}

// Do is used to initiate TLS connection
func (TLSHelper) Do(req *http.Request) (*http.Response, error) {
	if _, err := strconv.Atoi(req.URL.Port()); err != nil {
		return nil, fmt.Errorf("invalid URL port %q", req.URL.Port())
	}

	config, err := createClientConfig()
	if err != nil {
		log.Fatal("config failed: ", err)
	}

	tlsconn, err := tls.Dial("tcp", req.URL.Host, config)
	if err != nil {
		return nil, err
	}
	defer tlsconn.Close()

	req.Write(tlsconn)

	br := bufio.NewReader(tlsconn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		f := strings.SplitN(resp.Status, " ", 2)
		if len(f) < 2 {
			return nil, errors.New("unknown status code")
		}
		return nil, errors.New(f[1])
	}
	return resp, nil
}
