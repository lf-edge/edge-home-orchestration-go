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
	"io/ioutil"
	"sync/atomic"

	"github.com/satori/go.uuid"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
)

const (
	CertificateFileName = "edge-orchestration.crt"
	KeyFileName         = "edge-orchestration.key"
)

var (
	certFilePath atomic.Value
	handler      PSKHandler
	log          = logmgr.GetInstance()
)

func init() {
	certFilePath.Store("")
	handler = defaultIdentifier{}
}

func SetCertFilePath(path string) {
	log.Println("SetCertFilePath: ", path)
	certFilePath.Store(path)
}

func GetCertFilePath() string {
	return certFilePath.Load().(string)
}

type PSKHandler interface {
	GetIdentity() string
	GetKey(identity string) ([]byte, error)
}

type CertificateSetter interface {
	SetCertificateFilePath(path string)
}

type certificateGetter interface {
	GetCertificateFilePath() string
}

type HasCertificate struct {
	IsSetCert bool
}

func SetPSKHandler(h PSKHandler) {
	handler = h
}

func (h *HasCertificate) SetCertificateFilePath(path string) {
	SetCertFilePath(path)
	h.IsSetCert = true
}

func (h *HasCertificate) GetCertificateFilePath() string {
	if h.IsSetCert {
		return GetCertFilePath()
	}
	return ""
}

func GetIdentity() string {
	return handler.GetIdentity()
}

func GetKey(id string) ([]byte, error) {
	return handler.GetKey(id)
}

type defaultIdentifier struct{}

func (defaultIdentifier) GetIdentity() string {
	return uuid.Must(uuid.NewV4(), nil).String()
}

func (defaultIdentifier) GetKey(id string) ([]byte, error) {
	key, err := ioutil.ReadFile(GetCertFilePath() + "/" + KeyFileName)
	if err != nil {
		return nil, err
	}
	return key, nil
}
