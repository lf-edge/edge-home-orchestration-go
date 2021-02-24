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
package authenticator

import (
	"os"
	"testing"
)

const (
	fakejwtPath     = "fakejwt"
	fakejwtFilePath = fakejwtPath + "/" + passPhraseJWTFileName
)

func TestInit(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakejwtPath)

		if _, err := os.Stat(fakejwtPath); err != nil {
			err := os.MkdirAll(fakejwtPath, 444)
			if err != nil {
				t.Error(err.Error())
				return
			}
		}
		Init(fakejwtPath)
		if _, err := os.Stat(fakejwtPath); err != nil {
			t.Error(err.Error())
		}
		if _, err := os.Stat(fakejwtFilePath); os.IsNotExist(err) {
			t.Error(err.Error())
		}
		os.RemoveAll(fakejwtPath)

		Init(fakejwtPath)
		if _, err := os.Stat(fakejwtPath); err != nil {
			t.Error(err.Error())
		}

		if _, err := os.Stat(fakejwtFilePath); os.IsNotExist(err) {
			t.Error(err.Error())
		}
	})
	t.Run("Error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error(r)
			}
			os.RemoveAll("/fakejwtPath")
		}()

		Init("/fakejwtPath")
		if _, err := os.Stat("/fakejwtPath"); os.IsExist(err) {
			t.Error("unexpected success")
		}
	})
}
