/*******************************************************************************
 * Copyright 2019-2020 Samsung Electronics All Rights Reserved.
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

//Package discoverymgr provides functions to register local device to network and find other orchestration devices
package discoverymgr

import (
	wrapper "github.com/lf-edge/edge-home-orchestration-go/internal/controller/discoverymgr/wrapper"
	configurationdb "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/configuration"
	networkdb "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/network"
	servicedb "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/service"
	systemdb "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/system"
	"sync"
)

const (
	logPrefix       = "[discoverymgr]"
	edgeDirect      = "/var/edge-orchestration/"
	configPath      = edgeDirect + "mnedc/client.config"
)

// OrchestrationInformation is the struct to handle orchestration
type OrchestrationInformation struct {
	Platform      string `json:"Platform"`
	ExecutionType string `json:"ExecutionType"`

	// List of IP and Services
	IPv4          []string `json:"IPv4"`
	// IPv6       []string   `json:"IPv6"`
	ServiceList   []string `json:"ServiceList"`
}

// ExportDeviceMap gives device info map for discoverymgr user
type ExportDeviceMap map[string]OrchestrationInformation

type requestData struct {
	DeviceID  string
	PrivateIP string
	VirtualIP string
}

var (
	mapMTX           sync.Mutex
	wrapperIns       wrapper.ZeroconfInterface
	shutdownChan     chan struct{}
	isMNEDCConnected bool

	sysQuery     systemdb.DBInterface
	confQuery    configurationdb.DBInterface
	netQuery     networkdb.DBInterface
	serviceQuery servicedb.DBInterface

	configAlternate = "/storage/emulated/0/client.config"
)
