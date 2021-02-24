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
package cpu

import (
	"testing"

	"errors"
	"os"
	"time"
)

var originFileOpen func(name string) (*os.File, error)

func TestPercent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		originFileOpen = fileOpen
		fileOpen = fakeFileOpenCPUStat
		defer func() {
			fileOpen = originFileOpen
			removeFakeCPUStat()
		}()

		_, err := Percent(time.Second, true)
		if err != nil {
			t.Error("unexpected error")
		}
	})
	t.Run("Error", func(t *testing.T) {
		originFileOpen = fileOpen
		fileOpen = fakeFileOpenError
		defer func() {
			fileOpen = originFileOpen
		}()

		_, err := Percent(time.Second, true)
		if err == nil {
			t.Error("unexpected success")
		}
	})
}

func TestReadCPUUsage(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		originFileOpen = fileOpen
		fileOpen = fakeFileOpenCPUStat
		defer func() {
			fileOpen = originFileOpen
			removeFakeCPUStat()
		}()

		ret, err := readCPUUsage()
		if err != nil {
			t.Error("unexpected error")
		} else if ret[0] != 32034449.0 ||
			ret[1] != 1051857.0 ||
			ret[2] != 33086306.0 {
			t.Error("unexpected return value")
		}
	})
	t.Run("Error", func(t *testing.T) {
		originFileOpen = fileOpen
		fileOpen = fakeFileOpenError
		defer func() {
			fileOpen = originFileOpen
		}()

		_, err := readCPUUsage()
		if err == nil {
			t.Error("unexpected success")
		}
	})
}

func fakeFileOpenCPUStat(name string) (*os.File, error) {
	f, _ := os.Create("./fakecpustat")

	fakeFile := []byte("cpu  841350 3843 192386 32017606 16843 0 14278 0 0 0\ncpu0 101493 864 24208 4006865 573 0 3105 0 0 0\ncpu1 107576 38 24750 4002233 529 0 2152 0 0 0\ncpu2 116448 434 24907 3993428 975 0 1802 0 0 0\ncpu3 113891 1359 24687 3983124 11621 0 1772 0 0 0\n")
	f.Write(fakeFile)
	f.Close()

	return os.Open("./fakecpustat")
}

func removeFakeCPUStat() {
	os.Remove("./fakecpustat")
}

func TestGetCPUMaxFreq(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		originFileOpen = fileOpen
		fileOpen = fakeFileOpenCPUMaxFreq
		defer func() {
			fileOpen = originFileOpen
			removeFakeCPUMaxFreq()
		}()

		ret, err := getCPUMaxFreq()
		if err != nil {
			t.Error("unexpected error")
		} else if ret != 1000.0 {
			t.Error("unexpected result")
		}
	})
	t.Run("Error", func(t *testing.T) {
		originFileOpen = fileOpen
		fileOpen = fakeFileOpenError
		defer func() {
			fileOpen = originFileOpen
		}()

		_, err := getCPUMaxFreq()
		if err == nil {
			t.Error("unexpected success")
		}
	})
}

func fakeFileOpenCPUMaxFreq(name string) (*os.File, error) {
	f, _ := os.Create("./fakecpumaxfreq")

	fakeFile := []byte("1000000\n")
	f.Write(fakeFile)
	f.Close()

	return os.Open("./fakecpumaxfreq")
}

func removeFakeCPUMaxFreq() {
	os.Remove("./fakecpumaxfreq")
}

func TestGetCPUFreqCpuInfo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		originFileOpen = fileOpen
		fileOpen = fakeFileOpenCPUInfo
		defer func() {
			fileOpen = originFileOpen
			removeFakeCPUInfo()
		}()

		ret, err := getCPUFreqCpuInfo()
		if err != nil {
			t.Error("unexpected error")
		} else if ret != 3300.0 {
			t.Error("unexpected result")
		}
	})
	t.Run("Error", func(t *testing.T) {
		originFileOpen = fileOpen
		fileOpen = fakeFileOpenError
		defer func() {
			fileOpen = originFileOpen
		}()

		_, err := getCPUFreqCpuInfo()
		if err == nil {
			t.Error("unexpected success")
		}
	})
}

func TestGetCPUs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		originFileOpen = fileOpen
		fileOpen = fakeFileOpenCPUInfo
		defer func() {
			fileOpen = originFileOpen
			removeFakeCPUInfo()
		}()

		ret, err := getCPUs()
		if err != nil {
			t.Error("unexpected error")
		} else if len(ret) != 4 {
			t.Error("unexpected result")
		}
	})
	t.Run("Error", func(t *testing.T) {
		originFileOpen = fileOpen
		fileOpen = fakeFileOpenError
		defer func() {
			fileOpen = originFileOpen
		}()

		_, err := getCPUs()
		if err == nil {
			t.Error("unexpected success")
		}
	})
}

func fakeFileOpenError(name string) (*os.File, error) {
	return nil, errors.New("")
}

func fakeFileOpenCPUInfo(name string) (*os.File, error) {
	f, _ := os.Create("./fakecpuinfo")

	fakeFile := []byte("processor	: 0\nprocessor	: 1\nprocessor	: 2\n" +
		"processor	: 3\nprocessor\nprocessor	: asd\n" +
		"model name : Intel(R) Core(TM) i3-2120 CPU @ 3.30GHz\n")
	f.Write(fakeFile)
	f.Close()

	return os.Open("./fakecpuinfo")
}

func removeFakeCPUInfo() {
	os.Remove("./fakecpuinfo")
}
