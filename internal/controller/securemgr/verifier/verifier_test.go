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
package verifier

import (
	"os"
	"testing"
)

const (
	fakecwlPath            = "fakecwl"
	fakecwlFilePath        = fakecwlPath + "/" + cwlFileName
	hashHelloWorld         = "fc6a51919cfeb2e6763f62b6d9e8815acbf7cd2e476ea353743570610737b752"
	fakehashHelloWorld     = "1834bdb494c6150a9861cf32432df7c5d93fe2bc99e008da83a57a318dc207d7"
	fakehashExtraContainer = "99a55eca2c0afefdb019787b0e8d980e0efdf5c29db0d9004fbfe20612b73b96"
	unexpectedSuccess      = "unexpected success"
	unexpectedFail         = "unexpected fail"
)

func TestGetIndexHashInContainerName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_, err := getIndexHashInContainerName("hello-world@sha256:" + hashHelloWorld)
		if err != nil {
			t.Error(err.Error())
		}
	})
	t.Run("Error", func(t *testing.T) {
		index, err := getIndexHashInContainerName("hello-world@sha56:" + hashHelloWorld)
		if err == nil && index != -1 {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestContainerHashIsInWhiteList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {

		containerWhiteList = append(containerWhiteList, hashHelloWorld) // hello-world container image

		if err := containerHashIsInWhiteList(hashHelloWorld); err != nil {
			t.Error(unexpectedFail)
		}

		if err := containerHashIsInWhiteList("121212"); err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestPrintAllHashFromContainerWhiteList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		printAllHashFromContainerWhiteList()

		containerWhiteListOriginal := containerWhiteList
		containerWhiteList = nil
		printAllHashFromContainerWhiteList()
		containerWhiteList = containerWhiteListOriginal

	})
}

func TestContainerIsInWhiteList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		containerWhiteList = append(containerWhiteList, hashHelloWorld) // hello-world container image

		m := GetInstance()
		if err := m.ContainerIsInWhiteList("hello-world@sha256:" + hashHelloWorld); err != nil {
			t.Error(unexpectedFail)
		}

		initialized = true
		if err := m.ContainerIsInWhiteList("hello-world@sha256:" + hashHelloWorld); err != nil {
			t.Error(unexpectedFail)
		}

		if err := m.ContainerIsInWhiteList("hello-world@ha256:" + hashHelloWorld); err == nil {
			t.Error(unexpectedSuccess)
		}

		if err := m.ContainerIsInWhiteList("hello-world@sha256:" + fakehashHelloWorld); err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestInitContainerWhiteList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakecwlPath)

		os.RemoveAll(fakecwlPath)
		Init(fakecwlPath)
		if err := containerHashIsInWhiteList(hashHelloWorld); err == nil {
			t.Error(unexpectedSuccess)
		}

		if err := initContainerWhiteList(); err != nil {
			t.Error(unexpectedFail)
		}
	})
	t.Run("Error", func(t *testing.T) {
		if err := initContainerWhiteList(); err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestAddHashToContainerWhiteList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakecwlPath)

		Init(fakecwlPath)
		if err := addHashToContainerWhiteList(fakehashExtraContainer); err != nil {
			t.Error(unexpectedFail)
		}
		if err := containerHashIsInWhiteList(fakehashExtraContainer); err != nil {
			t.Error(unexpectedFail)
		}
		os.RemoveAll(fakecwlPath + "/" + cwlFileName)
		if err := addHashToContainerWhiteList(fakehashExtraContainer); err != nil {
			t.Error(unexpectedFail)
		}
		if err := containerHashIsInWhiteList(fakehashExtraContainer); err != nil {
			t.Error(unexpectedFail)
		}
		if err := addHashToContainerWhiteList(fakehashExtraContainer); err != nil {
			t.Error(unexpectedFail)
		}
	})
	t.Run("Error", func(t *testing.T) {
		if err := initContainerWhiteList(); err == nil {
			t.Error(unexpectedSuccess)
		}
		if err := addHashToContainerWhiteList(fakehashExtraContainer); err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestDelHashFromContainerWhiteList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakecwlPath)

		Init(fakecwlPath)
		if err := addHashToContainerWhiteList(fakehashExtraContainer); err != nil {
			t.Error(unexpectedFail)
		}
		if err := containerHashIsInWhiteList(fakehashExtraContainer); err != nil {
			t.Error(unexpectedFail)
		}

		if err := delHashFromContainerWhiteList(fakehashExtraContainer); err != nil {
			log.Println(logPrefix, "Can't create "+cwlFileName+": ", err)
		}

		if err := containerHashIsInWhiteList(fakehashExtraContainer); err == nil {
			t.Error(unexpectedSuccess)
		}

		if err := delHashFromContainerWhiteList(fakehashExtraContainer); err != nil {
			t.Error(unexpectedFail)
		}

	})
	t.Run("Error", func(t *testing.T) {
		os.RemoveAll(fakecwlPath + "/" + cwlFileName)
		if err := delHashFromContainerWhiteList(fakehashExtraContainer); err == nil {
			t.Error(unexpectedSuccess)
		}

	})
}

func TestDelAllHashFromContainerWhiteList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakecwlPath)

		Init(fakecwlPath)
		if err := addHashToContainerWhiteList(fakehashExtraContainer); err != nil {
			t.Error(unexpectedFail)
		}
		if err := delAllHashFromContainerWhiteList(); err != nil {
			t.Error(unexpectedFail)
		}

		if err := containerHashIsInWhiteList(fakehashExtraContainer); err == nil {
			t.Error(unexpectedSuccess)
		}

		if containerWhiteList != nil {
			t.Error(unexpectedSuccess)
		}
	})
	t.Run("Error", func(t *testing.T) {
		os.RemoveAll(fakecwlPath + "/" + cwlFileName)
		if err := delAllHashFromContainerWhiteList(); err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func RequestUpdateCWL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakecwlPath)

		Init(fakecwlPath)
		if err := addHashToContainerWhiteList(fakehashExtraContainer); err != nil {
			t.Error(unexpectedFail)
		}
		if err := containerHashIsInWhiteList(fakehashExtraContainer); err != nil {
			t.Error(unexpectedFail)
		}

		if err := delHashFromContainerWhiteList(fakehashExtraContainer); err != nil {
			t.Error(unexpectedSuccess)
		}

		if err := containerHashIsInWhiteList(fakehashExtraContainer); err == nil {
			t.Error(unexpectedSuccess)
		}
	})
	t.Run("Error", func(t *testing.T) {
		os.RemoveAll(fakecwlPath + "/" + cwlFileName)

		if err := delHashFromContainerWhiteList(fakehashExtraContainer); err != nil {
			t.Error(unexpectedSuccess)
		}

	})
}

func TestInit(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakecwlPath)

		Init(fakecwlPath)
		if _, err := os.Stat(fakecwlPath); err != nil {
			t.Error(err.Error())
		}

		if _, err := os.Stat(fakecwlFilePath); os.IsNotExist(err) {
			t.Error(err.Error())
		}
	})
	t.Run("Error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error(r)
			}
			os.RemoveAll("/fakecwlPath")
		}()
		Init("/fakecwlPath")
		if _, err := os.Stat("/fakecwlPath"); err != nil {
			t.Error(err.Error())
		}
	})

}

func TestRequestVerifierConf(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakecwlPath)

		Init(fakecwlPath)
		containerInfo := RequestVerifierConf{
			SecureInsName: "verifier",
			CmdType:       "addHashCWL",
			Desc: []RequestDescInfo{
				{
					//ContainerName: "hello_world_",
					ContainerHash: fakehashHelloWorld,
				},
				{
					//ContainerName: "hello_world",
					ContainerHash: hashHelloWorld,
				},
			},
		}
		m := GetInstance()

		resp := m.RequestVerifierConf(containerInfo)
		if resp.Message != ErrorNone {
			t.Error(unexpectedFail)
		}

		containerInfo.CmdType = "delHashCWL"
		resp = m.RequestVerifierConf(containerInfo)
		if resp.Message != ErrorNone {
			t.Error(unexpectedFail)
		}

		containerInfo.CmdType = "printAllHashCWL"
		resp = m.RequestVerifierConf(containerInfo)
		if resp.Message != ErrorNone {
			t.Error(unexpectedFail)
		}

		containerInfo.CmdType = "delAllHashCWL"
		resp = m.RequestVerifierConf(containerInfo)
		if resp.Message != ErrorNone {
			t.Error(unexpectedFail)
		}

		containerInfo.CmdType = "CWL"
		resp = m.RequestVerifierConf(containerInfo)
		if resp.Message == ErrorNone {
			t.Error(unexpectedSuccess)
		}
	})
}
