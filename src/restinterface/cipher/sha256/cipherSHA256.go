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

// Package sha256 implements encryption/decryption functions by sha256
package sha256

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	sHA256Cipher := new(Cipher)
	passphrase, err := ioutil.ReadFile(cipherKeyFilePath)
	if err != nil {
		sHA256Cipher.passphrase = []byte{}
		log.Println("len :", len(sHA256Cipher.passphrase))
		log.Println("can't read passphrase key from keyFilePath - ", err)
	} else {
		sHA256Cipher.passphrase = passphrase
	}
	return sHA256Cipher
}

// EncryptByte encrypts from []byte to []byte
func (ec *Cipher) EncryptByte(byteData []byte) (encryptedByte []byte, err error) {
	if len(byteData) == 0 {
		return nil, errors.New("input of encryptbyte is empty")
	}

	hashByte, err := ec.createHashByte()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	block, _ := aes.NewCipher(hashByte)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	encryptedByte = gcm.Seal(nonce, nonce, byteData, nil)
	return
}

// EncryptJSONToByte encrypts from map[string]interface{} to []byte
func (ec *Cipher) EncryptJSONToByte(jsonMap map[string]interface{}) (encryptedByte []byte, err error) {
	jsonByte, err := json.Marshal(jsonMap)
	if err != nil {
		return
	}
	encryptedByte, err = ec.EncryptByte(jsonByte)
	if err != nil {
		return
	}
	return
}

// DecryptByte decrypts from []byte to []byte
func (ec *Cipher) DecryptByte(byteData []byte) (decryptedByte []byte, err error) {
	if len(byteData) == 0 {
		return nil, fmt.Errorf("input of DecryptByte is empty")
	}

	hashByte, err := ec.createHashByte()
	if err != nil {
		log.Println(err)
		return
	}

	block, _ := aes.NewCipher(hashByte)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := byteData[:nonceSize], byteData[nonceSize:]
	decryptedByte, err = gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return
	}
	return
}

// DecryptByteToJSON decrypts from []byte to map[string]interface{}
func (ec *Cipher) DecryptByteToJSON(data []byte) (jsonMap map[string]interface{}, err error) {
	decrpytedByte, err := ec.DecryptByte(data)
	if err != nil {
		log.Println("descryption fail ", err.Error())
		return
	}

	err = json.Unmarshal(decrpytedByte, &jsonMap)
	if err != nil {
		log.Println("unmarshal fail ", err.Error())
		return
	}
	return
}

func (ec *Cipher) createHashByte() ([]byte, error) {
	if len(ec.passphrase) == 0 {
		log.Println("make error null hash byte ")
		return nil, errors.New("null hash byte")
	}

	hash := sha256.Sum256(ec.passphrase)
	return hash[:], nil
}
