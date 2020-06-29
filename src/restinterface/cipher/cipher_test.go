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

package cipher

import (
	"testing"
)

type fakeCipher struct{}

func TestSetCipher(t *testing.T) {
	testHasCipher := HasCipher{}
	testHasCipher.SetCipher(fakeCipher{})

	if testHasCipher.IsSetKey != true {
		t.Error("expected key is set, but not set")
	}
}

func (fakeCipher) EncryptByte(byteData []byte) (encryptedByte []byte, err error) {
	return
}
func (fakeCipher) EncryptJSONToByte(jsonMap map[string]interface{}) (encryptedByte []byte, err error) {
	return
}
func (fakeCipher) DecryptByte(byteData []byte) (decryptedByte []byte, err error) {
	return
}
func (fakeCipher) DecryptByteToJSON(data []byte) (jsonMap map[string]interface{}, err error) {
	return
}
