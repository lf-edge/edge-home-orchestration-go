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

package native

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/types/configuremgrtypes"
	contextmgr "github.com/lf-edge/edge-home-orchestration-go/src/controller/configuremgr"
)

var name string

const (
	expectedName = "HelloWorldService"
)

type dummyNoti struct{}

func (d dummyNoti) Notify(serviceinfo configuremgrtypes.ServiceInfo) {
	log.Println(serviceinfo.ServiceName)
	name = serviceinfo.ServiceName
}

func TestSetConfigPath(t *testing.T) {
	testConfigObj := new(ConfigureMgr)

	configFilePath := "/etc/edge-orchestration"
	err := testConfigObj.SetConfigPath(configFilePath)
	if err == nil {
		if strings.Compare(configFilePath, configuremgrObj.confpath) != 0 {
			t.Errorf("%s != %s", configFilePath, configuremgrObj.confpath)
		}
	}
}

func TestBasicMockConfigureMgr(t *testing.T) {

	var contextNoti contextmgr.Notifier
	contextNoti = new(dummyNoti)

	//copy event environment
	watchDir := "/tmp/foo"
	src := "./mock/mysum"
	dst := watchDir

	// testConfigObj := &ConfigureMgr{confpath: watchDir}
	testConfigObj := new(ConfigureMgr)
	testConfigObj.confpath = watchDir

	execCommand("mkdir -p /tmp/foo/")

	go testConfigObj.Watch(contextNoti)

	//TODO : push /tmp/foo/simple directory using Cmd package
	time.Sleep(time.Duration(1 * time.Second))

	//init scenario
	execCommand("rm -rf /tmp/foo/mysum")
	time.Sleep(time.Duration(1) * time.Second)

	//user scenario
	execCommand(fmt.Sprintf("cp -ar %s %s", src, dst))

	time.Sleep(time.Duration(5) * time.Second)

	if name != expectedName {
		t.Errorf("Not matched notified serviceName")
	}

	// testConfigObj.Done <- true
}

func execCommand(command string) {
	log.Println(command)
	cmd := exec.Command("sh", "-c", command)
	stdoutStderr, err := cmd.CombinedOutput()
	log.Printf("%s", stdoutStderr)
	if err != nil {
		log.Fatal(err)
	}
}
