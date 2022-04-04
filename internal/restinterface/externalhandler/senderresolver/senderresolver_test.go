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
	"io/ioutil"
	"os"
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

var (
	test = `0: 00000000:5012 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 258820 1 ffff9290f095cec0 100 0 0 10 0
1: 00000000:BBD2 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 257808 1 ffff929189365780 100 0 0 10 0
2: 0100007F:C3-2 0100007F:DAC1 01 00000000:00000000 02:00001770 00000000  1000        0 9305104 3 ffff92922f9ad780 20 0 0 10 -1
`
	test1 = `0: 00000000:5012 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 258820 1 ffff9290f095cec0 100 0 0 10 0
1: 00000000:BBD2 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 257808 1 ffff929189365780 100 0 0 10 0
2: 0100007F:C3A2 0100007F:D-01 01 00000000:00000000 02:00001770 00000000  1000        0 9305104 3 ffff92922f9ad780 20 0 0 10 -1
`
)

const (
	unexpectedSuccess = "unexpected success"
	unexpectedFail    = "unexpected fail"
	unexpectedError   = "unexpected error: "
	fakeProcNetTCP    = "fakeprocnettcp"
)

func init() {
	wait.Add(1)
}

type handler struct{}

func TestGetPid(t *testing.T) {
	t.Run("Fail", func(t *testing.T) {
		originProc := processInfoPath
		processInfoPath = "\\"
		defer func() {
			processInfoPath = originProc
		}()

		_, err := getPid("fd")
		if err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestGetProcess(t *testing.T) {
	t.Run("Fail", func(t *testing.T) {
		_, err := getProcess("")
		if err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestGetData(t *testing.T) {
	t.Run("Fail", func(t *testing.T) {
		originProcNetTCP := procNetTCP
		defer func() {
			procNetTCP = originProcNetTCP
		}()

		procNetTCP = "\\"

		_, err := getData()
		if err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestGetNameByPort(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		originProc := processInfoPath
		processInfoPath = "/proc"
		defer func() {
			processInfoPath = originProc

		}()
		go func() {
			err := http.ListenAndServe("0.0.0.0:56001", handler{})
			if err != nil {
				t.Error(unexpectedError, err.Error())
			}
		}()

		time.Sleep(time.Microsecond * 100)

		_, err := http.Get("http://0.0.0.0:56001")
		if err != nil {
			t.Error(unexpectedError, err.Error())
		}

		wait.Wait()

		port, err := strconv.Atoi(strings.Split(remoteAddr, ":")[1])
		if err != nil {
			t.Error(unexpectedError, err.Error())
		}

		_, err = GetNameByPort(int64(port))
		if err != nil {
			t.Error(unexpectedError, err.Error())
		}
	})
	t.Run("Fail", func(t *testing.T) {
		originProc := processInfoPath
		originProcNetTCP := procNetTCP
		processInfoPath = "/proc"
		defer func() {
			_ = os.Remove(fakeProcNetTCP)
			procNetTCP = originProcNetTCP
			processInfoPath = originProc
		}()

		port, err := strconv.Atoi(strings.Split(remoteAddr, ":")[1])
		if err != nil {
			t.Error(unexpectedError, err.Error())
		}

		processInfoPath = "\\"
		_, err = GetNameByPort(int64(port))
		if err == nil {
			t.Error(unexpectedSuccess)
		}

		processInfoPath = originProc
		_, err = GetNameByPort(100000)
		if err == nil {
			t.Error(unexpectedSuccess)
		}

		procNetTCP = "\\"
		_, err = GetNameByPort(int64(port))
		if err == nil {
			t.Error(unexpectedSuccess)
		}

		procNetTCP = fakeProcNetTCP
		err = ioutil.WriteFile(procNetTCP, []byte(test), 0644)
		if err != nil {
			t.Error(unexpectedError, err.Error())
		}

		_, err = GetNameByPort(int64(port))
		if err == nil {
			t.Error(unexpectedSuccess)
		}

		err = ioutil.WriteFile(procNetTCP, []byte(test1), 0644)
		if err != nil {
			t.Error(unexpectedError, err.Error())
		}

		_, err = GetNameByPort(int64(port))
		if err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func (handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	println(r.RemoteAddr)
	remoteAddr = r.RemoteAddr
	wait.Done()
}
