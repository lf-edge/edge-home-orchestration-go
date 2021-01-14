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

package tlspskserver

import (
	"net/http"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"

	rafftls "github.com/raff/tls-ext"
	"github.com/raff/tls-psk"

	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/tls"
)

var (
	log = logmgr.GetInstance()
)

type TLSPSKServerListener interface {
	ListenAndServe(addr string, handler http.Handler)
}

type TLSPSKServer struct{}

func (TLSPSKServer) ListenAndServe(addr string, handler http.Handler) {
	var config = &rafftls.Config{
		CipherSuites: []uint16{psk.TLS_PSK_WITH_AES_128_CBC_SHA},
		Certificates: []rafftls.Certificate{rafftls.Certificate{}},
		Extra: psk.PSKConfig{
			GetKey:      tls.GetKey,
			GetIdentity: tls.GetIdentity,
		},
	}

	listener, err := rafftls.Listen("tcp", addr, config)
	if err != nil {
		log.Println(err.Error())
	}
	defer listener.Close()

	(&http.Server{Handler: handler}).Serve(listener)
}
