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
	"io/ioutil"
	"os"
	"testing"
)

const (
	fakejwtPath     = "fakejwt"
	fakejwtFilePath = fakejwtPath + "/" + passPhraseJWTFileName
)

const fakejwtpubKey = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0qFs/QpUuZx8mjysmY7M
rt9PL6GBwgyl+yh5NrBM5P+SmTC0epfF+o/2vQcRvx0ensi37uFAYPZhVa712Bte
a/ID5c7usRlqPkIHpHd4DImdZd8EvGBEJolutHDv62qa4cqVOOCpkFXWDMSewxoY
ND0zGk7OFI7D8ttkUsvectjXW2tHRkHTcaj65NLBHA4HNGR9p8MHV1FIgWF7aZU6
SQBiuaHzBMhZlWL+OEa4kVTl//kpaMFwvGB7Jis4og9LUPZx0bZXxVGVz0Wqe5L5
D05zqiDCvS6Tkoc9gtNrFMmzuysnrXLrewWhqVewLatN/wBHndc8MsvugV1phqvL
4QIDAQAB
-----END PUBLIC KEY-----`

func TestInit(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakejwtPath)

		if _, err := os.Stat(fakejwtPath); err != nil {
			err := os.MkdirAll(fakejwtPath, 0444)
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

		err := ioutil.WriteFile(fakejwtPath+"/"+pubKeyPath, []byte(fakejwtpubKey), 0666)
		if err != nil {
			t.Error(err.Error())
		}
		Init(fakejwtPath)
		os.RemoveAll(fakejwtPath + "/" + pubKeyPath)

		err = ioutil.WriteFile(fakejwtPath+"/"+pubKeyPath, []byte(""), 0666)
		if err != nil {
			t.Error(err.Error())
		}
		Init(fakejwtPath)
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
