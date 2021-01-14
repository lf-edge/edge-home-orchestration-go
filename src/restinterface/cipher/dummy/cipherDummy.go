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

// Package dummy provides the mocking functions
package dummy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"

	c "github.com/lf-edge/edge-home-orchestration-go/src/restinterface/cipher"
)

// Cipher has passphrase for ciphering
type Cipher struct {
	passphrase []byte
}

var (
	log = logmgr.GetInstance()
)

// GetCipher set passphrase for ciphering
func GetCipher(cipherKeyFilePath string) c.IEdgeCipherer {
	dummyCipher := new(Cipher)
	passphrase, err := ioutil.ReadFile(cipherKeyFilePath)
	if err != nil {
		dummyCipher.passphrase = []byte{}
		log.Println("can't read passphrase key from keyFilePath - ", err)
	} else {
		dummyCipher.passphrase = passphrase
	}

	return dummyCipher
}

// EncryptByte is mocking function for EncryptByte
func (ec *Cipher) EncryptByte(byteData []byte) ([]byte, error) {
	return byteData, nil
}

// EncryptJSONToByte is mocking function for EncryptJSONToByte
func (ec *Cipher) EncryptJSONToByte(jsonMap map[string]interface{}) ([]byte, error) {
	jsonByte, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonByte, nil
}

// DecryptByte is mocking function for DecryptByte
func (ec *Cipher) DecryptByte(byteData []byte) ([]byte, error) {

	if len(byteData) == 0 {
		return nil, fmt.Errorf("input of DecryptByte is empty")
	}

	return byteData, nil
}

// DecryptByteToJSON is mocking function for DecryptByteToJSON
func (ec *Cipher) DecryptByteToJSON(data []byte) (jsonMap map[string]interface{}, err error) {

	err = json.Unmarshal(data, &jsonMap)
	if err != nil {
		return nil, err
	}
	return
}
