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

// Package restinterface implements internal/external REST API
package restinterface

import (
	"net/http"

	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/cipher"
)

// Route type
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes type
type Routes []Route

//IRestRoutes is the interface implemented by Get/Set for REST API
type IRestRoutes interface {
	cipher.Setter
	GetRoutes() Routes
}

// HasRoutes wraps Routes
type HasRoutes struct {
	Routes Routes
}

// GetRoutes returns Routes
func (h HasRoutes) GetRoutes() Routes {
	return h.Routes
}
