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

package logmgr

import (
	"fmt"
	"os"
	"testing"
)

func TestInit(t *testing.T) {

	logFilePath, _ := os.Getwd()
	log := GetInstance()
	log.Printf("FilePath = %s", logFilePath)
	logFileName = "logmgr.log"

	InitLogfile(logFilePath)

	TestFile := logFilePath + "/" + logFileName

	if _, err := os.Stat(TestFile); os.IsNotExist(err) {
		t.Error(err.Error())
	}

	err := os.Remove(TestFile)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestInitFolderFail(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("fail to create log :", r)
		} else {
			t.Error(r)
		}
	}()

	logFilePath := ""
	InitLogfile(logFilePath)
}

func TestInitFileFail(t *testing.T) {
	logFilePath, _ := os.Getwd()
	logFilePath += "/test"

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("fail to create log :", r)
			err := os.RemoveAll(logFilePath)
			if err != nil {
				t.Error(err.Error())
			}
		} else {
			t.Error(r)
		}
	}()

	if _, err := os.Stat(logFilePath); err != nil {
		err := os.MkdirAll(logFilePath, 0444)
		if err != nil {
			t.Error(err.Error())
		}
	}
	InitLogfile(logFilePath)

	err := os.RemoveAll(logFilePath)
	if err != nil {
		t.Error(err.Error())
	}
}
