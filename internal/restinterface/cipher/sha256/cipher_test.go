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
package sha256

import (
	"encoding/json"
	"strings"
	"testing"
)

var ec *Cipher

func TestGetCipher(t *testing.T) {
	c := GetCipher("")
	data := []byte("I love you !!")
	_, err := c.EncryptByte(data)
	log.Println(err)
	if err == nil {
		t.Error()
	}
}

func TestEncryptDecryptByte(t *testing.T) {
	ec := Cipher{}
	ec.passphrase = []byte("edge-orchestration")

	data := []byte("I love you !!")
	encryptedByte, err := ec.EncryptByte(data)

	errCheck(t, err)
	if len(encryptedByte) == 0 {
		t.Error(err)
	} else {
		log.Println("passphrase: ", string(ec.passphrase))
		log.Println(encryptedByte)
	}

	decryptedByte, err := ec.DecryptByte(encryptedByte)

	errCheck(t, err)
	if len(decryptedByte) == 0 {
		t.Error(err)
	} else {
		log.Println(decryptedByte)
		log.Println(string(decryptedByte))
	}

	assertEqualByteSlice(t, data, decryptedByte)

	ec.passphrase = []byte("")
	encryptedByte, err = ec.EncryptByte(data)
	noErrCheck(t, err)

	decryptedByte, err = ec.DecryptByte(data)
	noErrCheck(t, err)

	var nulByte []byte
	encryptedNulByte, err := ec.EncryptByte(nulByte)
	if err == nil || len(encryptedNulByte) != 0 {
		t.Error(encryptedNulByte)
	}

	decryptedNulByte, err := ec.DecryptByte(nulByte)
	if err == nil || len(encryptedNulByte) != 0 {
		t.Error(decryptedNulByte)
	}
}

func TestEncryptDecryptByteJsonMap(t *testing.T) {
	ec := Cipher{}
	ec.passphrase = []byte("edge-orchestration")

	doc := `{"member":7,"project":"edge-orchestration"}`

	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(doc), &jsonMap)

	encryptedByte, err := ec.EncryptJSONToByte(jsonMap)

	errCheck(t, err)
	if len(encryptedByte) == 0 {
		t.Error(err)
	} else {
		log.Println(encryptedByte)
	}

	decryptedJSONMap, err := ec.DecryptByteToJSON(encryptedByte)
	jsonByte, err := json.Marshal(decryptedJSONMap)

	errCheck(t, err)
	if len(jsonByte) == 0 {
		t.Error(err)
	} else {
		log.Println(jsonByte)
		log.Println(string(jsonByte))
	}

	assertEqualStr(t, doc, string(jsonByte))

	ec.passphrase = []byte("")
	encryptedByte, err = ec.EncryptJSONToByte(jsonMap)
	noErrCheck(t, err)

	decryptedJSONMap, err = ec.DecryptByteToJSON(encryptedByte)
	noErrCheck(t, err)

	var nulByte []byte
	encryptedNulByte, err := ec.EncryptByte(nulByte)
	if err == nil || len(encryptedNulByte) != 0 {
		t.Error(encryptedNulByte)
	}

	decryptedNulByte, err := ec.DecryptByte(nulByte)
	if err == nil || len(encryptedNulByte) != 0 {
		t.Error(decryptedNulByte)
	}
}

func assertEqualByteSlice(t *testing.T, a, b []byte) {
	t.Helper()
	if len(a) != len(b) {
		t.Error("byte array not equal", a, b)
	}
	for i, v := range a {
		if v != b[i] {
			t.Error("byte array not equal", a, b)
		}
	}
}

func assertEqualStr(t *testing.T, a, b string) {
	t.Helper()
	if strings.Compare(a, b) != 0 {
		t.Errorf("%s != %s", a, b)
	}
}

func errCheck(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Error(err)
	}
}

func noErrCheck(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error(err)
	}
}
