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
	"io/ioutil"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/client"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/discoverymgr/mnedc/server"
)

const (
	deviceID   = "edge-orchestration-fake"
	configFile = "client.config"
	saddr      = "localhost"
	sport      = "3334"
)

func TestWatch(t *testing.T) {
	t.Run("Watch with MNEDC", func(t *testing.T) {
		defer func() {
			err := os.Remove(configFile)
			if err != nil {
				log.Println("Could not delete file")
			}
		}()
		_, err := server.GetInstance().CreateServer(saddr, sport, false)
		if err != nil {
			t.Error(err.Error())
		} else {
			server.GetInstance().Run()
			if _, err := os.Stat(configFile); err != nil {
				err = ioutil.WriteFile(configFile, []byte(saddr+"\n"+sport), 0644)
				if err != nil {
					log.Panicf("Failed to create a mnedc config file %s: %s\n", configFile, err)
				}
			}
			_, err = client.GetInstance().CreateClient(deviceID, configFile, false)
			if err != nil {
				t.Error(err.Error())
			}
			go func() {
				time.Sleep(1 * time.Second)
					syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}()
			Watch()
		}
	})

	t.Run("Watch without MNEDC", func(t *testing.T) {
		go func() {
			time.Sleep(1 * time.Second)
			server.GetInstance().Close()
			client.GetInstance().Close()
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}()
		Watch()
	})
}
