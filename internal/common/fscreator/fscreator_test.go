/*******************************************************************************
* Copyright 2021 Samsung Electronics All Rights Reserved.
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

package fscreator

import (
	"os"
	"testing"
)

const (
	fakefsPath        = "fakefs"
	unexpectedSuccess = "unexpected success"
	unexpectedFail    = "unexpected fail"
)

func TestCreateFileSystem(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakefsPath)

		CreateFileSystem(fakefsPath)
		for _, dir := range edgeDirs {
			if _, err := os.Stat(fakefsPath+dir); err != nil {
				t.Error(err.Error())
			}
		}
	})

	t.Run("Error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error(r)
			}
			for _, dir := range edgeDirs {
				if _, err := os.Stat("/"+fakefsPath+dir); err == nil {
					t.Error(unexpectedSuccess)
				}
			}
			os.RemoveAll("/"+fakefsPath)
		}()
		CreateFileSystem("/"+fakefsPath)
	})
}
