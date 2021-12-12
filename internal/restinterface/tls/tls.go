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
	"sync/atomic"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
)

// File names of certificate and private key
const (
	CertificateFileName = "edge-orchestration.crt"
	KeyFileName         = "edge-orchestration.key"
)

var (
	certFilePath atomic.Value
	handler      Handler
	log          = logmgr.GetInstance()
)

func init() {
	certFilePath.Store("")
	handler = defaultIdentifier{}
}

// SetCertFilePath stores a certificate file path
func SetCertFilePath(path string) {
	log.Println("SetCertFilePath: ", path)
	certFilePath.Store(path)
}

// GetCertFilePath gets a certificate file path
func GetCertFilePath() string {
	return certFilePath.Load().(string)
}

// Handler provides interfaces for the tls
type Handler interface{}

// CertificateSetter interface provides setting the certificate file path
type CertificateSetter interface {
	SetCertificateFilePath(path string)
}

type certificateGetter interface {
	GetCertificateFilePath() string
}

// HasCertificate indicates presence of the certificate
type HasCertificate struct {
	IsSetCert bool
}

// SetHandler sets handler
func SetHandler(h Handler) {
	handler = h
}

// SetCertificateFilePath stores a certificate via handler
func (h *HasCertificate) SetCertificateFilePath(path string) {
	SetCertFilePath(path)
	h.IsSetCert = true
}

// GetCertificateFilePath gets a certificate via handler
func (h *HasCertificate) GetCertificateFilePath() string {
	if h.IsSetCert {
		return GetCertFilePath()
	}
	return ""
}

type defaultIdentifier struct{}
