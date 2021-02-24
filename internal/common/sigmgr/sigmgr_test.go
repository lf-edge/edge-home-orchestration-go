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

package sigmgr

import (
	"syscall"
	"testing"
	"time"

	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/discoverymgr/mnedc/client"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/discoverymgr/mnedc/server"
)

func TestWatch(t *testing.T) {
	// TC without MNEDC running
	t.Run("Watch without MNEDC", func(t *testing.T) {
		go func() {
			time.Sleep(1 * time.Second)
			server.GetInstance().Close()
			client.GetInstance().Close()
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}()
		Watch()
	})
	// TODO: TC needs to be implemented when MNEDC server and client are running.
}
