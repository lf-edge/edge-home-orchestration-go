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

package restinterface

import (
	"net/http"
	"testing"
)

const (
	unexpectedSuccess = "unexpected success"
	unexpectedFail    = "unexpected fail"
)

func TestGetRoutes(t *testing.T) {
	hasRouter := new(HasRoutes)
	route1handler := func(w http.ResponseWriter, r *http.Request) {}

	hasRouter.Routes = Routes{
		Route{Name: "route1", Method: "GET", Pattern: "/api/v1/route1", HandlerFunc: route1handler},
	}
	if hasRouter1 := hasRouter.GetRoutes(); hasRouter1 == nil {
		t.Error(unexpectedFail)
	}
}
