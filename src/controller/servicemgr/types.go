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

package servicemgr

import (
	"errors"
	"sync"
	"sync/atomic"
)

const logPrefix = "[servicemgr]"

const (
	// ConstKeyServiceID is key of service id
	ConstKeyServiceID = "ServiceID"

	// ConstKeyServiceName is key of service name
	ConstKeyServiceName = "ServiceName"

	// ConstKeyNotiChan is key of notification channel
	// ConstKeyNotiChan = "NotiChan"

	// ConstKeyStatus is key of status
	ConstKeyStatus = "Status"

	// ConstKeyUserArgs is key of user argument
	ConstKeyUserArgs = "UserArgs"

	ConstKeyRequester = "Requester"

	// ConstKeyNotiTargetURL is key of notification target URL
	ConstKeyNotiTargetURL = "NotificationTargetURL"

	// ConstServiceStatusFailed is service status is failed
	ConstServiceStatusFailed = "Failed"

	// ConstServiceStatusStarted is service status is started
	ConstServiceStatusStarted = "Started"

	// ConstServiceStatusFinished is service status is finished
	ConstServiceStatusFinished = "Finished"

	// ConstServiceFound is service status is found
	ConstServiceFound = "Found"

	// ConstServiceNotFound is service status is not found
	ConstServiceNotFound = "NotFound"

	//ConstServicePort is open port for rest server
	ConstServicePort = 56001
)

// ServiceExecutionResponse structure
type ServiceExecutionResponse struct {
	Status string `json:"Status"`
}

// ServiceDestroyResponse structure
type ServiceDestroyResponse struct {
	Return string `json:"Return"`
}

// StatusNotification structure
type StatusNotification struct {
	ServiceID uint64 `json:"ServiceID"`
	Status    string `json:"Status"`
}

// ConcurrentMap struct
type ConcurrentMap struct {
	sync.RWMutex
	items map[uint64]interface{}
}

// ConcurrentMapItem struct
type ConcurrentMapItem struct {
	Key   uint64
	Value interface{}
}

var (
	// ErrInvalidService is for error type of invalid service
	ErrInvalidService = errors.New("it is invalid service")

	// ServiceMap is service map
	ServiceMap ConcurrentMap

	// ServiceIdx is for unique service ID (process id)
	ServiceIdx uint64
)

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

// Iter is for iterating map item
func (cm *ConcurrentMap) Iter() <-chan ConcurrentMapItem {
	c := make(chan ConcurrentMapItem)

	go func() {
		cm.Lock()
		defer cm.Unlock()

		for k, v := range cm.items {
			c <- ConcurrentMapItem{k, v}
		}
		close(c)
	}()

	return c
}

func createServiceMap(name string) uint64 {
	serviceID := getServiceIdx()

	value := make(map[string]interface{})

	value[ConstKeyServiceName] = name

	ServiceMap.Set(serviceID, value)

	return serviceID
}

func deleteServiceMap(serviceID uint64) {
	ServiceMap.Remove(serviceID)
}

// getServiceIdx() is for getting global serviceID
func getServiceIdx() uint64 {
	atomic.AddUint64(&ServiceIdx, 1)
	return atomic.LoadUint64(&ServiceIdx)
}
