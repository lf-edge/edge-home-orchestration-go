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
	"testing"
)

const (
	testPath          = "/test/testcert/path"
	unexpectedSuccess = "unexpected success"
	unexpectedFail    = "unexpected fail"
)

func TestGetSetCertFilePath(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		SetCertFilePath(testPath)
		if ret := GetCertFilePath(); ret != testPath {
			t.Error(unexpectedFail)
		}
	})
}

func TestHasCertificate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		testHasCertificate := HasCertificate{}
		testHasCertificate.SetCertificateFilePath(testPath)

		if ret := testHasCertificate.GetCertificateFilePath(); ret != testPath {
			t.Error(unexpectedFail)
		} else if testHasCertificate.IsSetCert != true {
			t.Error(unexpectedFail)
		}

		testHasCertificate.IsSetCert = false

		if ret := testHasCertificate.GetCertificateFilePath(); ret != "" {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestSetHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		handlerTest := new(Handler)
		SetHandler(handlerTest)
		if handler != handlerTest {
			t.Error(unexpectedFail)
		}
	})
}
