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

package senderresolver

import (
	"testing"

	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	remoteAddr string
	wait       sync.WaitGroup
)

func init() {
	wait.Add(1)
}

type handler struct{}

func TestGetNameByPort(t *testing.T) {
	originProc := processInfoPath
	processInfoPath = "/proc"
	defer func() {
		processInfoPath = originProc
	}()

	go func() {
		err := http.ListenAndServe("0.0.0.0:56001", handler{})
		if err != nil {
			println("unexpected error: ", err.Error())
			t.Error("unexpected error: ", err.Error())
		}
	}()

	time.Sleep(time.Microsecond * 100)

	_, err := http.Get("http://0.0.0.0:56001")
	if err != nil {
		println("unexpected error: ", err.Error())
		t.Error("unexpected error: ", err.Error())
	}

	wait.Wait()

	port, err := strconv.Atoi(strings.Split(remoteAddr, ":")[1])
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}

	println("port : ", strconv.Itoa(port))

	pName, err := GetNameByPort(int64(port))
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	println(pName)
}

func (handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	println(r.RemoteAddr)
	remoteAddr = r.RemoteAddr
	wait.Done()
}
