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
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/securemgr/authorizer"
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
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
	pubKeyPath            = "app_rsa.pub"
)

var (
	logPrefix             = "[securemgr: authenticator]"
	log                   = logmgr.GetInstance()
	authenticatorIns      *AuthenticatorImpl
	passphrase            = []byte{}
	passPhraseJWTFilePath = ""
	initialized           = false
	rsaKeyInitialized     = false
	verifyKey             *rsa.PublicKey
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

// Init sets the environments for securemgr
func Init(passPhraseJWTPath string) {
	if _, err := os.Stat(passPhraseJWTPath); err != nil {
		err := os.MkdirAll(passPhraseJWTPath, os.ModePerm)
		if err != nil {
			log.Panicf("%s Failed to create passPhraseJWTPath: %s\n", logPrefix, err)
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
			log.Println(logPrefix, "Cannot create passPhraseJWT.txt:", err)
		}
	}

	verifyBytes, err := ioutil.ReadFile(passPhraseJWTPath + "/" + pubKeyPath)
	if err != nil {
		log.Println(logPrefix, err)
	} else {
		verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
		if err != nil {
			log.Fatal(err)
		} else {
			rsaKeyInitialized = true
		}
	}

	initialized = true
}

var IsAuthorizedRequest = func(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if initialized == false {
			next.ServeHTTP(w, r) // pass control to the next handler
			return
		}
		notReqAuth := []string{
			"/api/v1/ping",
			"/api/v1/servicemgr/services",
			"/api/v1/servicemgr/services/notification/{serviceid}",
			"/api/v1/scoringmgr/score",
		}
		for _, url := range notReqAuth {

			if url == r.URL.Path {
				next.ServeHTTP(w, r)
				return
			}
		}

		if r.Header["Authorization"] != nil {

			token, err := jwt.Parse(r.Header["Authorization"][0], func(token *jwt.Token) (interface{}, error) {
				// log.Println(token.Claims)
				// log.Printf("%s Signing method: %v\n", logPrefix, jwt.GetSigningMethod(fmt.Sprintf("%v", token.Header["alg"])))

				switch token.Header["alg"] {
				case "HS256":
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
					}
					return passphrase, nil
				case "RS256":
					if rsaKeyInitialized {
						if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
							return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
						}
						return verifyKey, nil
					} else {
						return nil, errors.New("RSA keys are not initialized")
					}
				}
				return nil, errors.New("Unsupported algo")
			})

			if err != nil {
				log.Println(logPrefix, err.Error())
			}

			if token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					name, _ := claims["aud"].(string)
					if err = authorizer.Authorizer(name, r); err == nil {
						next.ServeHTTP(w, r) // pass control to the next handler
					}
				}
			}
		} else {
			log.Println(logPrefix, "Request doesn't contain an Authorization token")
		}
	})
}
