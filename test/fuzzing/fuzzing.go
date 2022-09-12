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

package fuzz

import (
	"net/http"
	"os"
	"testing"

	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/securemgr/authorizer"
)

const (
	fakerbacPath      = "fakerbac"
	unexpectedSuccess = "unexpected success"
	unexpectedFail    = "unexpected fail"
)

func FuzzTestAuthorizer(f *testing.F) {
	defer func() {
		os.RemoveAll(fakerbacPath)
	}()

	authorizer.Init(fakerbacPath)

	req, err := http.NewRequest("POST", "/api/v1/orchestration/securemgr", nil)
	if err != nil {
		return
	}

	f.Fuzz(func(t *testing.T, data []byte) int {
		orig := string(data)
		if err := authorizer.Authorizer(orig, req); err != nil {
			return 0
		}

		if orig == "Admin" || orig == "Member" {
			return 0
		}
		t.Error(unexpectedSuccess)
		return 1
	})
}
