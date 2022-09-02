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

package configuremgr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/types/configuremgrtypes"
)

var name string

const (
	expectedName    = "HelloWorldService"
	defaultConfPath = "testdata/apps"
	fakeConfPath    = "fake"
	fakeExecType    = "native"
)

const (
	fakePath          = "fakePath"
	unexpectedSuccess = "unexpected success"
	unexpectedFail    = "unexpected fail"
)

type dummyNoti struct{}

func (d dummyNoti) Notify(serviceinfo configuremgrtypes.ServiceInfo) {
	log.Println(serviceinfo.ServiceName)
	name = serviceinfo.ServiceName
}

func TestSetConfigPath(t *testing.T) {
	os.Mkdir(defaultConfPath, 0775)
	defer os.RemoveAll(defaultConfPath)

	t.Run("Success", func(t *testing.T) {
		testConfigObj := new(ConfigureMgr)

		err := testConfigObj.SetConfigPath(defaultConfPath)
		if err == nil {
			if strings.Compare(defaultConfPath, configuremgrObj.confpath) != 0 {
				t.Errorf("%s != %s", defaultConfPath, configuremgrObj.confpath)
			}
		} else {
			t.Errorf(err.Error())
		}
	})
	t.Run("No File", func(t *testing.T) {
		testConfigObj := new(ConfigureMgr)

		err := testConfigObj.SetConfigPath(fakeConfPath)
		if err == nil {
			if strings.Compare(fakeConfPath, configuremgrObj.confpath) != 0 {
				t.Errorf("%s != %s", fakeConfPath, configuremgrObj.confpath)
			}
		}
	})
}

func TestBasicMockConfigureMgr(t *testing.T) {
	os.Mkdir(defaultConfPath, 0775)
	defer os.RemoveAll(defaultConfPath)

	var contextNoti Notifier = new(dummyNoti)
	src := "testdata/score"

	t.Run("Success", func(t *testing.T) {
		testConfigObj := GetInstance(defaultConfPath, fakeExecType)

		go testConfigObj.Watch(contextNoti)
		time.Sleep(time.Duration(1) * time.Second)

		dir := defaultConfPath + "/score"
		os.RemoveAll(dir)
		err := os.Mkdir(dir, 0775)
		if err != nil {
			t.Errorf(err.Error())
		} else {
			files, err := os.ReadDir(src)
			if err != nil {
				t.Error(err.Error())
			}
			for _, file := range files {
				fileContent, _ := os.ReadFile(filepath.Join(src, file.Name()))
				err = os.WriteFile(filepath.Join(dir, file.Name()), []byte(fileContent), 0664)
				if err != nil {
					t.Errorf(err.Error())
				}
			}
		}
		time.Sleep(time.Duration(5) * time.Second)

		if name != expectedName {
			t.Errorf("Not matched notified serviceName %s != %s", name, expectedName)
		}
	})
}

func TestGetServiceInfo(t *testing.T) {
	t.Run("Fail", func(t *testing.T) {
		_, err := getServiceInfo(fakePath)
		if err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestGetdirname(t *testing.T) {
	t.Run("Fail", func(t *testing.T) {
		_, err := getdirname(fakePath + "/")
		if err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}
