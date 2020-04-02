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
package securemgr

import (
	"os"
	"testing"
)

const (
	fakecwlPath			= "fakecwl"
	fakecwlFilePath		= fakecwlPath + "/" + cwlFileName
	fakehashHelloWorld	= "1834bdb494c6150a9861cf32432df7c5d93fe2bc99e008da83a57a318dc207d7"
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
			t.Error("unexpected success")
		}
	})
}

func TestContainerHashIsInWhiteList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		containerWhiteList = append(containerWhiteList, hashHelloWorld) // hello-world container image

		if !containerHashIsInWhiteList(hashHelloWorld) {
			t.Error("unexpected fail")
		}

		if containerHashIsInWhiteList("121212") {
			t.Error("unexpected success")
		}
	})
}

func TestContainerIsInWhiteList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		containerWhiteList = append(containerWhiteList, hashHelloWorld) // hello-world container image

		m := GetInstance()
		if err := m.ContainerIsInWhiteList("hello-world@sha256:" + hashHelloWorld); err != nil {
			t.Error("unexpected fail")
		}

		initialized = true
		if err := m.ContainerIsInWhiteList("hello-world@sha256:" + hashHelloWorld); err != nil {
			t.Error("unexpected fail")
		}

		if err := m.ContainerIsInWhiteList("hello-world@ha256:" + hashHelloWorld); err == nil {
			t.Error("unexpected success")
		}

		if err := m.ContainerIsInWhiteList("hello-world@sha256:" + fakehashHelloWorld); err == nil {
			t.Error("unexpected success")
		}
	})
}

func TestInitContainerWhiteList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakecwlPath)

		Init(fakecwlPath)
		if !containerHashIsInWhiteList(hashHelloWorld) {
			t.Error("unexpected fail")
		}

		if err := initContainerWhiteList(fakecwlFilePath); err != nil {
			t.Error("unexpected fail")
		}
		if !containerHashIsInWhiteList(hashHelloWorld) {
			t.Error("unexpected fail")
		}
	})
	t.Run("Error", func(t *testing.T) {
		err := initContainerWhiteList("")
		if err == nil {
			t.Error("unexpected success")
		}
	})
}

func TestInit(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakecwlPath)

		Init("./")
		if _, err := os.Stat(cwlFileName); err != nil {
			t.Error("unexpected success")
		}
		if err := os.Remove(cwlFileName); err != nil {
			t.Error(err.Error())
		}

		Init(fakecwlPath)
		if _, err := os.Stat(fakecwlPath); err != nil {
			t.Error(err.Error())
		}

		if _, err := os.Stat(fakecwlFilePath); os.IsNotExist(err) {
			t.Error(err.Error())
		}
	})
//	t.Run("Error", func(t *testing.T) {
//	})
}

