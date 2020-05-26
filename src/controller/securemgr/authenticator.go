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
package securemgr

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// AuthenticatorImpl structure
type AuthenticatorImpl struct{}

const (
	passPhraseJWTFileName = "passPhraseJWT.txt"
)

var (
	logPrefixA            = "[securemgr: authenticator]"
	authenticatorIns      *AuthenticatorImpl
	passphrase            = []byte{}
	passPhraseJWTFilePath = ""
	initializedA          = false
)

func init() {
	authenticatorIns = new(AuthenticatorImpl)
}

var alphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(b)
}

// InitA sets the environments for securemgr
func InitA(passPhraseJWTPath string) {
	if _, err := os.Stat(passPhraseJWTPath); err != nil {
		err := os.MkdirAll(passPhraseJWTPath, os.ModePerm)
		if err != nil {
			log.Panicf("Failed to create passPhraseJWTPath %s: %s\n", passPhraseJWTPath, err)
			return
		}
	}

	passPhraseJWTFilePath = passPhraseJWTPath + "/" + passPhraseJWTFileName

	var err error
	passphrase, err = ioutil.ReadFile(passPhraseJWTFilePath)
	if err != nil {
		rand.Seed(time.Now().UnixNano())
		passphrase = []byte(randString(16))
		err = ioutil.WriteFile(passPhraseJWTFilePath, passphrase, 0666)
		if err != nil {
			log.Println(logPrefixA, "cannot create "+passPhraseJWTFilePath+": ", err)
		}
	}
	initializedA = true
}

var IsAuthorizedRequest = func(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authenticatorIns == nil {
			log.Println(logPrefixA, "authenticatorIns not initialized")
			return
		}

		if r.Header["Authorization"] != nil {

			token, err := jwt.Parse(r.Header["Authorization"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					log.Println(logPrefixA, "authenticatorIns not initialized")
					return nil, errors.New("Token has an error")
				}
				if !initializedA {
					passphrase = []byte("")
				}
				return passphrase, nil
			})

			if err != nil {
				fmt.Fprintf(w, err.Error())
			}

			if token.Valid {
				next.ServeHTTP(w, r) // pass control to the next handler
			}
		} else {
			fmt.Fprintf(w, "Request doesn't contain an Authorization token\n")
		}
	})
}
