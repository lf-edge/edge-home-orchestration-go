/*******************************************************************************
 * Copyright 2022 Samsung Electronics All Rights Reserved.
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
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/securemgr/authenticator"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/securemgr/authorizer"
	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/securemgr/verifier"
)

// Handle Platform Dependencies
const (
	logPrefix              = "[securemgr]"
	certificateFilePath    = "/data/cert"
	containerWhiteListPath = "/data/cwl"
	passPhraseJWTPath      = "/data/jwt"
	rbacRulePath           = "/data/rbac"
)

// SecuremgrImpl structure
type SecuremgrImpl struct {
	IsSecured bool
}

var (
	securemgrIns *SecuremgrImpl
	log          = logmgr.GetInstance()
)

func init() {
	securemgrIns = new(SecuremgrImpl)
}

// GetInstance gives the SecuremgrImpl singletone instance
func GetInstance() *SecuremgrImpl {
	return securemgrIns
}

// Start initializes the securemgr components
func Start(edgeDir string) {
	verifier.Init(edgeDir + containerWhiteListPath)
	authenticator.Init(edgeDir + passPhraseJWTPath)
	authorizer.Init(edgeDir + rbacRulePath)
	securemgrIns.IsSecured = true

	log.Info("Start securemgr")
}
