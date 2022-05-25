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

const (
	fakecpustat       = "./fakecpustat"
	fakecpumaxfreq    = "./fakecpumaxfreq"
	unexpectedSuccess = "unexpected success"
	unexpectedFail    = "unexpected fail"
	unexpectedResult  = "unexpected result"
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
			t.Error(unexpectedFail)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		originFileOpen = fileOpen
		fileOpen = fakeFileOpenError
		defer func() {
			fileOpen = originFileOpen
		}()

		_, err := Percent(time.Second, true)
		if err == nil {
			t.Error(unexpectedSuccess)
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
			t.Error(unexpectedFail)
		} else if ret[0] != 32034449.0 ||
			ret[1] != 1051857.0 ||
			ret[2] != 33086306.0 {
			t.Error("unexpected return value")
		}
	})
	t.Run("Fail", func(t *testing.T) {
		t.Run("FileOpenError", func(t *testing.T) {
			originFileOpen = fileOpen
			fileOpen = fakeFileOpenError
			defer func() {
				fileOpen = originFileOpen
			}()

			_, err := readCPUUsage()
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
		t.Run("AbsentDelim", func(t *testing.T) {
			originFileOpen = fileOpen
			fileOpen = fakeFileOpenCPUStatDelim
			defer func() {
				fileOpen = originFileOpen
				removeFakeCPUStat()
			}()

			_, err := readCPUUsage()
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
		t.Run("WrongFormat", func(t *testing.T) {
			originFileOpen = fileOpen
			fileOpen = fakeFileOpenCPUStatFormat
			defer func() {
				fileOpen = originFileOpen
				removeFakeCPUStat()
			}()

			_, err := readCPUUsage()
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
	})
}

func fakeFileOpenCPUStat(name string) (*os.File, error) {
	f, _ := os.Create(fakecpustat)

	fakeFile := []byte("cpu  841350 3843 192386 32017606 16843 0 14278 0 0 0\ncpu0 101493 864 24208 4006865 573 0 3105 0 0 0\ncpu1 107576 38 24750 4002233 529 0 2152 0 0 0\ncpu2 116448 434 24907 3993428 975 0 1802 0 0 0\ncpu3 113891 1359 24687 3983124 11621 0 1772 0 0 0\n")
	f.Write(fakeFile)
	f.Close()

	return os.Open(fakecpustat)
}

func fakeFileOpenCPUStatDelim(name string) (*os.File, error) {
	f, _ := os.Create(fakecpustat)

	fakeFile := []byte("cpu  841350 3843 192386 32017606 16843 0 14278 0 0 0")
	f.Write(fakeFile)
	f.Close()

	return os.Open(fakecpustat)
}

func fakeFileOpenCPUStatFormat(name string) (*os.File, error) {
	f, _ := os.Create(fakecpustat)

	fakeFile := []byte("cpu  841350 3843 192386 32017606 16843 0 14278 0 0 0 cpu0 101493 864 24208 4006865 573 0 3105 0 0 0\n")
	f.Write(fakeFile)
	f.Close()

	return os.Open(fakecpustat)
}

func removeFakeCPUStat() {
	os.Remove(fakecpustat)
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
			t.Error(unexpectedFail)
		} else if ret != 1000.0 {
			t.Error(unexpectedResult)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		t.Run("FileOpenError", func(t *testing.T) {
			originFileOpen = fileOpen
			fileOpen = fakeFileOpenError
			defer func() {
				fileOpen = originFileOpen
				removeFakeCPUMaxFreq()
			}()

			_, err := getCPUMaxFreq()
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
		t.Run("FileOpenCPUMaxFreqDelim", func(t *testing.T) {
			originFileOpen = fileOpen
			fileOpen = fakeFileOpenCPUMaxFreqDelim
			defer func() {
				fileOpen = originFileOpen
				removeFakeCPUMaxFreq()
			}()

			_, err := getCPUMaxFreq()
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
		t.Run("FileOpenCPUMaxFreqFormat", func(t *testing.T) {
			originFileOpen = fileOpen
			fileOpen = fakeFileOpenCPUMaxFreqFormat
			defer func() {
				fileOpen = originFileOpen
				removeFakeCPUMaxFreq()
			}()

			_, err := getCPUMaxFreq()
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
	})
}

func fakeFileOpenCPUMaxFreq(name string) (*os.File, error) {
	f, _ := os.Create(fakecpumaxfreq)

	fakeFile := []byte("1000000\n")
	f.Write(fakeFile)
	f.Close()

	return os.Open(fakecpumaxfreq)
}

func fakeFileOpenCPUMaxFreqDelim(name string) (*os.File, error) {
	f, _ := os.Create(fakecpumaxfreq)

	fakeFile := []byte("1000000")
	f.Write(fakeFile)
	f.Close()

	return os.Open(fakecpumaxfreq)
}

func fakeFileOpenCPUMaxFreqFormat(name string) (*os.File, error) {
	f, _ := os.Create(fakecpumaxfreq)

	fakeFile := []byte("10000-00\n")
	f.Write(fakeFile)
	f.Close()

	return os.Open(fakecpumaxfreq)
}

func removeFakeCPUMaxFreq() {
	os.Remove(fakecpumaxfreq)
}

func TestGetCPUFreqCPUInfo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		originFileOpen = fileOpen
		fileOpen = fakeFileOpenCPUInfo
		defer func() {
			fileOpen = originFileOpen
			removeFakeCPUInfo()
		}()

		ret, err := getCPUFreqCPUInfo()
		if err != nil {
			t.Error(unexpectedFail)
		} else if ret != 3300.0 {
			t.Error(unexpectedResult)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		t.Run("FileOpenError", func(t *testing.T) {
			originFileOpen = fileOpen
			fileOpen = fakeFileOpenError
			defer func() {
				fileOpen = originFileOpen
				removeFakeCPUInfo()
			}()

			_, err := getCPUFreqCPUInfo()
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
		t.Run("FileOpenCPUInfoDelim", func(t *testing.T) {
			originFileOpen = fileOpen
			fileOpen = fakeFileOpenCPUInfoDelim
			defer func() {
				fileOpen = originFileOpen
				removeFakeCPUInfo()
			}()

			_, err := getCPUFreqCPUInfo()
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
		t.Run("FileOpenCPUInfoFormat", func(t *testing.T) {
			originFileOpen = fileOpen
			fileOpen = fakeFileOpenCPUInfoFormat
			defer func() {
				fileOpen = originFileOpen
				removeFakeCPUInfo()
			}()

			_, err := getCPUFreqCPUInfo()
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
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
			t.Error(unexpectedFail)
		} else if len(ret) != 4 {
			t.Error(unexpectedResult)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		originFileOpen = fileOpen
		fileOpen = fakeFileOpenError
		defer func() {
			fileOpen = originFileOpen
		}()

		_, err := getCPUs()
		if err == nil {
			t.Error(unexpectedSuccess)
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

func fakeFileOpenCPUInfoDelim(name string) (*os.File, error) {
	f, _ := os.Create("./fakecpuinfo")

	fakeFile := []byte("processor	: 0 processor	: 1 processor	: 2 " +
		"processor	: 3 processor processor	: asd " +
		"model name : Intel(R) Core(TM) i3-2120 CPU @ 3.30GHz")
	f.Write(fakeFile)
	f.Close()

	return os.Open("./fakecpuinfo")
}

func fakeFileOpenCPUInfoFormat(name string) (*os.File, error) {
	f, _ := os.Create("./fakecpuinfo")

	fakeFile := []byte("processor	: 0\nprocessor	: 1\nprocessor	: 2\n" +
		"processor	: 3\nprocessor\nprocessor	: asd\n" +
		"model name : Intel(R) Core(TM) i3-2120 CPU @ 3.-30GHz\n")
	f.Write(fakeFile)
	f.Close()

	return os.Open("./fakecpuinfo")
}

func removeFakeCPUInfo() {
	os.Remove("./fakecpuinfo")
}

func TestInfo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_, err := Info()
		if err != nil {
			t.Error(unexpectedFail)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		t.Run("FileOpenError", func(t *testing.T) {
			originFileOpen = fileOpen
			fileOpen = fakeFileOpenError
			defer func() {
				fileOpen = originFileOpen
			}()

			_, err := Info()
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
		t.Run("FileOpenCPUInfoDelim", func(t *testing.T) {
			originFileOpen = fileOpen
			fileOpen = fakeFileOpenCPUInfoDelim
			defer func() {
				fileOpen = originFileOpen
				removeFakeCPUInfo()
			}()

			_, err := Info()
			if err == nil {
				t.Error(unexpectedSuccess)
			}
		})
	})
}
