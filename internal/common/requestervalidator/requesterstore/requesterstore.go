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

package requesterstore

import (
	"errors"
	"sync"
)

type RequesterStore interface {
	GetRequester(serviceName string) ([]string, error)
	StoreRequesterInfo(serviceName string, requesters []string)
}

type requesters struct {
	requesterInfos map[string][]string
	mutex          *sync.Mutex
}

const notFoundRegisteredService = "not found registered service"

var requesterList requesters

func GetInstance() RequesterStore {
	return &requesterList
}

func init() {
	requesterList.mutex = &sync.Mutex{}
	requesterList.requesterInfos = make(map[string][]string)
}

func (r requesters) GetRequester(serviceName string) ([]string, error) {
	val, ok := r.requesterInfos[serviceName]
	if !ok {
		return nil, errors.New(notFoundRegisteredService)
	}

	return val, nil
}

func (r *requesters) StoreRequesterInfo(serviceName string, requesters []string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.requesterInfos[serviceName] = make([]string, len(requesters))
	for idx, req := range requesters {
		r.requesterInfos[serviceName][idx] = req
	}
}
