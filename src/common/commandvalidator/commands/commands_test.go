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

package commands

import (
	"testing"

	"strconv"
	"sync"
)

func TestStoreServiceInfo(t *testing.T) {
	serviceName := "test"
	executableName := "testExecutable"

	t.Run("Success", func(t *testing.T) {
		t.Run("StoreOnce", func(t *testing.T) {
			GetInstance().StoreServiceInfo(serviceName, executableName)

			expected, err := GetInstance().GetServiceFileName(serviceName)
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
					GetInstance().StoreServiceInfo(k, v)
				}(key, value)
			}

			wait.Wait()

			for key, value := range tests {
				expected, err := GetInstance().GetServiceFileName(key)
				if err != nil {
					t.Error("unexpected error")
				} else if expected != value {
					t.Error("stored value not exist")
				}
			}
		})
	})
}

func TestGetServiceFileName(t *testing.T) {
	serviceName := "test"
	executableName := "testExecutable"
	t.Run("Success", func(t *testing.T) {
		GetInstance().StoreServiceInfo(serviceName, executableName)

		expected, err := GetInstance().GetServiceFileName(serviceName)
		if err != nil {
			t.Error("unexpected error")
		} else if expected != executableName {
			t.Error("returning value does not same with store value")
		}
	})

	t.Run("Error", func(t *testing.T) {
		_, err := GetInstance().GetServiceFileName(serviceName + "error")
		if err == nil {
			t.Error("unexpected succeed")
		}
	})
}
