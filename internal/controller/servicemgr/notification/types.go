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

package notification

import "sync"

const (
	// ConstKeyNotiChan is key of notification channel
	ConstKeyNotiChan = "NotiChan"
)

const logPrefix = "[notification]"

// ConcurrentMap type
type ConcurrentMap struct {
	sync.RWMutex
	items map[uint64]interface{}
}

// ConcurrentMapItem type
type ConcurrentMapItem struct {
	Key   uint64
	Value interface{}
}

// Set is for setting map item
func (cm *ConcurrentMap) Set(key uint64, value interface{}) {
	cm.Lock()
	defer cm.Unlock()

	cm.items[key] = value
}

// Get is for getting map item
func (cm *ConcurrentMap) Get(key uint64) (interface{}, bool) {
	cm.Lock()
	defer cm.Unlock()

	value, ok := cm.items[key]

	return value, ok
}

// Remove is for removing map item
func (cm *ConcurrentMap) Remove(key uint64) {
	cm.Lock()
	defer cm.Unlock()

	delete(cm.items, key)
}
