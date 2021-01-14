/*******************************************************************************
 * Copyright 2020 Samsung Electronics All Rights Reserved.
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

package sigmgr

import (
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
	"os"
	"os/signal"
	"syscall"

	"github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/client"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/server"
)

const (
	logPrefix = "[sigmgr]"
)

var (
	log = logmgr.GetInstance()
)

// Watch operating system signals
func Watch() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	s := <-sig
	log.Println(logPrefix, "Received Signal:", s)
	err := server.GetInstance().Close()
	if err != nil {
		log.Println(logPrefix, "[MNEDC Server]", err.Error())
	} else {
		log.Println(logPrefix, "[MNEDC Server]", "Server Closed")
	}
	err = client.GetInstance().Close()
	if err != nil {
		log.Println(logPrefix, "[MNEDC Client]", err.Error())
	} else {
		log.Println(logPrefix, "[MNEDC Client]", "Client Closed")
	}
}
