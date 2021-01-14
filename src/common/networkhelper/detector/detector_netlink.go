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

// Package detector implements wrapper of netlink package
package detector

import (
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"

	"github.com/vishvananda/netlink"
)

const logPrefix = "detector_netlink"

type detectorImpl struct{}

var detectorIns detectorImpl
var (
	subChan chan netlink.AddrUpdate
	done    chan struct{}
	log     = logmgr.GetInstance()
)

var addrSubscribe func(ch chan<- netlink.AddrUpdate, done <-chan struct{}) error

// Detector adds subscribers that want to detect network interface changes
type Detector interface {
	AddrSubscribe(chan<- bool)
}

func init() {
	subChan = make(chan netlink.AddrUpdate, 1)
	done = make(chan struct{}, 1)

	addrSubscribe = netlink.AddrSubscribe
}

// GetInstance returns a detectorImpl struct instance
func GetInstance() Detector {
	return detectorIns
}

// AddrSubscribe is needed to get notification of NW change
func (dt detectorImpl) AddrSubscribe(isTrue chan<- bool) {
	for {
		err := addrSubscribe(subChan, done)
		if err == nil {
			break
		}
		log.Println(err)
	}

	for {
		select {
		case detect := <-subChan:
			isTrue <- detectionHandler(detect)
		}
	}
}

func detectionHandler(detect netlink.AddrUpdate) bool {
	updatedAddr := detect
	if updatedAddr.LinkAddress.IP.To4() == nil {
		return false
	}

	if updatedAddr.NewAddr {
		log.Println(logPrefix, "[DetectionHandler]", "New Connection : ", updatedAddr.LinkAddress.IP)
		return true
	}

	log.Println(logPrefix, "[DetectionHandler]", "Disconnected : ", updatedAddr.LinkAddress.IP)
	return false
}
