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

package webui

import (
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	"net/http"
	"strconv"
	"time"
)

const timeout = 15

var (
	assetFolder = "web"
	uiPort      = 49153
	log         = logmgr.GetInstance()
)

// Start starts the server for web UI
func Start() {
	s := &http.Server{
		Handler:      uiHandler(),
		Addr:         ":" + strconv.Itoa(uiPort),
		WriteTimeout: timeout * time.Second,
		ReadTimeout:  timeout * time.Second,
	}

	go s.ListenAndServe()
	log.Debug("Start UI server")
}

func uiHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(assetFolder)).ServeHTTP(w, r)
	})
}
