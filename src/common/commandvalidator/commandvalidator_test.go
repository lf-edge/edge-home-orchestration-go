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
package commandvalidator

import (
	"testing"

	"strconv"
	"sync"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/types/configuremgrtypes"
)

func TestCheckCommand(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		t.Run("HasInjectionOperator", func(t *testing.T) {
			serviceName := "test"
			command := []string{"ls", ";", "ps"}
			err := CommandValidator{}.CheckCommand(serviceName, command)
			if err == nil {
				t.Error("unexpected succeed")
			}
		})
		t.Run("NotSotredService", func(t *testing.T) {
			serviceName := "test"
			command := []string{"ls"}
			err := CommandValidator{}.CheckCommand(serviceName, command)
			if err == nil {
				t.Error("unexpected succeed")
			}
		})
		t.Run("NotMatchedService", func(t *testing.T) {
			serviceName := "test"
			command := []string{"ls"}
			validator := CommandValidator{}
			info := configuremgrtypes.ServiceInfo{ServiceName: serviceName, ExecutableFileName: "dir"}
			err := validator.AddWhiteCommand(info)
			if err != nil {
				t.Error("unexpected error: ", err.Error())
			}

			err = validator.CheckCommand(serviceName, command)
			if err == nil {
				t.Error("unexpected succeed")
			}
		})
		t.Run("InvalidCommand", func(t *testing.T) {
			serviceName := "test"
			command := []string{""}
			err := CommandValidator{}.CheckCommand(serviceName, command)
			if err == nil {
				t.Error("unexpected succeed")
			}
		})
	})
	t.Run("Success", func(t *testing.T) {
		serviceName := "TestCheckCommand/Success"
		command := []string{"ls"}
		validator := CommandValidator{}
		info := configuremgrtypes.ServiceInfo{ServiceName: serviceName, ExecutableFileName: command[0]}
		err := validator.AddWhiteCommand(info)
		if err != nil {
			t.Error("unexpected error: ", err.Error())
		}

		err = validator.CheckCommand(serviceName, command)
		if err != nil {
			t.Error("unexpected succeed")
		}
	})
}

func TestStoreServiceInfo(t *testing.T) {
	serviceName := "TestStoreServiceInfo"
	executableName := "testExecutable"

	validator := CommandValidator{}

	t.Run("Success", func(t *testing.T) {
		t.Run("StoreOnce", func(t *testing.T) {
			info := configuremgrtypes.ServiceInfo{ServiceName: serviceName, ExecutableFileName: executableName}
			if err := validator.AddWhiteCommand(info); err != nil {
				t.Error("unexpected error: " + err.Error())
			}

			expected, err := validator.GetCommand(serviceName)
			if err != nil {
				t.Error("unexpected error")
			} else if expected != executableName {
				t.Error("returning value does not same with store value")
			}
		})
		t.Run("StoreMultipleAsRacing", func(t *testing.T) {
			iter := 10000
			tests := make(map[string]string, iter)

			for idx := 0; idx < iter; idx++ {
				tests[serviceName+strconv.Itoa(idx)] = executableName + strconv.Itoa(idx)
			}

			var wait sync.WaitGroup
			wait.Add(iter)

			for key, value := range tests {
				go func(k, v string) {
					defer wait.Done()
					if err := validator.AddWhiteCommand(configuremgrtypes.ServiceInfo{
						ServiceName:        k,
						ExecutableFileName: v,
					}); err != nil {
						t.Error("unexpected error: " + err.Error())
					}
				}(key, value)
			}

			wait.Wait()

			for key, value := range tests {
				expected, err := validator.GetCommand(key)
				if err != nil {
					t.Error("unexpected error")
				} else if expected != value {
					t.Error("stored value not exist")
				}
			}
		})
	})
	t.Run("Error", func(t *testing.T) {
		t.Run("StoreBlackList", func(t *testing.T) {
			for _, black := range blackList {
				info := configuremgrtypes.ServiceInfo{ServiceName: serviceName, ExecutableFileName: black}
				if err := validator.AddWhiteCommand(info); err == nil {
					t.Error("unexpected succeed, " + black)
				}
			}
		})
	})
}

func TestGetServiceFileName(t *testing.T) {
	serviceName := "TestGetServiceFileName"
	executableName := "testExecutable"

	validator := CommandValidator{}

	t.Run("Success", func(t *testing.T) {
		info := configuremgrtypes.ServiceInfo{ServiceName: serviceName, ExecutableFileName: executableName}
		validator.AddWhiteCommand(info)

		expected, err := validator.GetCommand(serviceName)
		if err != nil {
			t.Error("unexpected error")
		} else if expected != executableName {
			t.Error("returning value does not same with store value")
		}
	})

	t.Run("Error", func(t *testing.T) {
		_, err := validator.GetCommand(serviceName + "error")
		if err == nil {
			t.Error("unexpected succeed")
		}
	})
}

func TestAddWhiteCommand(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		serviceName := "test"
		executableName := ""
		info := configuremgrtypes.ServiceInfo{ServiceName: serviceName, ExecutableFileName: executableName}
		err := CommandValidator{}.AddWhiteCommand(info)
		if err == nil {
			t.Error("unexpected success")
		}
	})
}

func TestGetExecutableName(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		command, err := getExecutableName("")
		if err == nil {
			t.Error("unexpected success")
		} else if command != "" {
			t.Error("unexpected command")
		}
	})
	t.Run("Success", func(t *testing.T) {
		t.Run("CommandInPath", func(t *testing.T) {
			command, err := getExecutableName("ls")
			if err != nil {
				t.Error("unexpected error: ", err.Error())
			} else if command != "ls" {
				t.Error("unexpected command")
			}
		})
		t.Run("DirectPath", func(t *testing.T) {
			command, err := getExecutableName("/usr/bin/ls")
			if err != nil {
				t.Error("unexpected error: ", err.Error())
			} else if command != "ls" {
				t.Error("unexpected command")
			}
		})
	})
}

var blackList = []string{
	"sudo",
	"su",
	"bash",
	"bsh",
	"csh",
	"adb",
	"sh",
	"ssh",
	"scp",
	"cat",
	"chage",
	"chpasswd",
	"dmidecode",
	"dmsetup",
	"fcinfo",
	"fdisk",
	"iscsiadm",
	"lsof",
	"multipath",
	"oratab",
	"prtvtoc",
	"ps",
	"pburn",
	"pfexec",
	"dzdo",
}
