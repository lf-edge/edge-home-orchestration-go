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

package tls

import (
	"os"
	"testing"
)

func TestGetSetCertFilePath(t *testing.T) {
	testPath := "/test/testcert/path"

	SetCertFilePath(testPath)

	ret := GetCertFilePath()
	if ret != testPath {
		t.Error("expect: ", testPath, " returned: ", ret)
	}
}

func TestHasCertificate(t *testing.T) {
	testHasCertificate := HasCertificate{}

	testPath := "/test/testcert/path"

	testHasCertificate.SetCertificateFilePath(testPath)

	ret := testHasCertificate.GetCertificateFilePath()
	if ret != testPath {
		t.Error("expect: ", testPath, " returned: ", ret)
	} else if testHasCertificate.IsSetCert != true {
		t.Error("expect certificate key is set, but not set")
	}

	testHasCertificate.IsSetCert = false
	ret = testHasCertificate.GetCertificateFilePath()
	if ret != "" {
		t.Error("unexpected value: ")
	}
}

func TestGetIdentity(t *testing.T) {
	if GetIdentity() == GetIdentity() {
		t.Error("expect different value, but it is same")
	}
}

func TestGetKey(t *testing.T) {
	testKeyName := "edge-orchestration.key"
	testPath := "./"
	testStr := "this is test"
	SetCertFilePath(testPath)

	t.Run("Error", func(t *testing.T) {
		_, err := GetKey("")
		if err == nil {
			t.Error("unexpected succeed")
		}
	})
	t.Run("Success", func(t *testing.T) {
		f, err := os.Create(testPath + testKeyName)
		if err != nil {
			t.Error(err.Error())
		}
		defer func() {
			f.Close()
			os.Remove(testPath + testKeyName)
		}()

		f.Write([]byte(testStr))

		ret, err := GetKey("")
		if err != nil {
			t.Error(err.Error())
		}
		if string(ret) != testStr {
			t.Error("unexpeted value: ", string(ret))
		}
	})

}
