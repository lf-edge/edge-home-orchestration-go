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
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/raff/tls-ext"
	"github.com/raff/tls-psk"

	tlscipher "github.com/lf-edge/edge-home-orchestration-go/src/restinterface/tls"
)

var (
	//	config        *tls.Config
	wellKnownPort map[string]string
)

type TLSHelper struct{}

func init() {
	wellKnownPort = map[string]string{
		"http":  "80",
		"https": "443",
	}
}

func (TLSHelper) Do(req *http.Request) (*http.Response, error) {
	if _, err := strconv.Atoi(req.URL.Port()); err != nil {
		return nil, fmt.Errorf("invalid URL port %q", req.URL.Port())
	}

	config := &tls.Config{
		CipherSuites: []uint16{psk.TLS_PSK_WITH_AES_128_CBC_SHA},
		Certificates: []tls.Certificate{tls.Certificate{}},
		Extra: psk.PSKConfig{
			GetKey:      tlscipher.GetKey,
			GetIdentity: tlscipher.GetIdentity,
		},
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
