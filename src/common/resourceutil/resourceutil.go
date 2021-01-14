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

// Package resourceutil provides the information of resource usage of local device
package resourceutil

import (
	"github.com/lf-edge/edge-home-orchestration-go/src/common/errors"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"

	resourceDB "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/resource"
)

// ResourceImpl is implementation for resourceutil interface
type ResourceImpl struct {
	targetDeviceID string
}

// MonitorImpl is implementation for Monitor interface
type MonitorImpl struct {
	netScoring func()
	cpuScoring func()
	memScoring func()
	rttScoring func()
}

var (
	resourceDBExecutor resourceDB.DBInterface
	monitoringExecutor MonitorImpl
	log                = logmgr.GetInstance()
)

func init() {
	resourceDBExecutor = resourceDB.Query{}
}

const (
	// CPUUsage defined cpu/usage
	CPUUsage = "cpu/usage"
	// CPUCount defined cpu/count
	CPUCount = "cpu/count"
	// CPUFreq defined cpu/freq
	CPUFreq = "cpu/freq"
	// MemFree defined memory/free
	MemFree = "memory/free"
	// MemAvailable defined memory/available
	MemAvailable = "memory/available"
	// NetMBps defined network/mbps
	NetMBps = "network/mbps"
	// NetBandwidth defined network/bandwidth
	NetBandwidth = "network/bandwidth"
	// NetRTT defined network/rtt
	NetRTT = "network/rtt"

	logPrefix             = "resourceutil"
	defaultProcessingTime = 5
)

// Command is an interface to control resource monitoring operations
type Command interface {
	Run(string) float64
	Close()
}

// Monitor is an interface to get device resource
type Monitor interface {
	StartMonitoringResource()
}

// GetResource is an interface to get resource
type GetResource interface {
	GetResource(string) (float64, error)
	SetDeviceID(string)
}

// GetMonitoringInstance return MonitorImpl instance
func GetMonitoringInstance() *MonitorImpl {
	monitoringExecutor.netScoring = processNetInfo
	monitoringExecutor.cpuScoring = processCPUInfo
	monitoringExecutor.memScoring = processMEMInfo
	monitoringExecutor.rttScoring = processRTT
	return &monitoringExecutor
}

// StartMonitoringResource to get device resources
func (m MonitorImpl) StartMonitoringResource() {
	m.netScoring()
	m.cpuScoring()
	m.memScoring()
	m.rttScoring()
}

// GetResource returns a resource value that matches resourceName
func (r *ResourceImpl) GetResource(resourceName string) (float64, error) {
	switch resourceName {
	case CPUUsage:
		return getCPUUsage()
	case CPUCount:
		return getCPUCount()
	case CPUFreq:
		return getCPUFreq()
	case MemFree:
		return getMemoryFree()
	case MemAvailable:
		return getMemoryAvailable()
	case NetMBps:
		return getNetworkMBps()
	case NetBandwidth:
		return getNetworkBandwidth()
	case NetRTT:
		return getNetworkRTT(r.targetDeviceID)
	default:
		return 0.0, errors.NotSupport{Message: "Not suppoted resource name"}
	}
}

// SetDeviceID set target device's id for RTT
func (r *ResourceImpl) SetDeviceID(ID string) {
	r.targetDeviceID = ID
}

func getCPUUsage() (out float64, err error) {
	info, err := resourceDBExecutor.Get(CPUUsage)
	if err != nil {
		return
	}
	out = info.Value
	return
}

func getCPUFreq() (out float64, err error) {
	info, err := resourceDBExecutor.Get(CPUFreq)
	if err != nil {
		return
	}
	out = info.Value
	return
}

func getCPUCount() (out float64, err error) {
	info, err := resourceDBExecutor.Get(CPUCount)
	if err != nil {
		return
	}
	out = info.Value
	return
}

func getMemoryAvailable() (out float64, err error) {
	info, err := resourceDBExecutor.Get(MemAvailable)
	if err != nil {
		return
	}
	out = info.Value
	return
}

func getMemoryFree() (out float64, err error) {
	info, err := resourceDBExecutor.Get(MemFree)
	if err != nil {
		return
	}
	out = info.Value
	return
}

func getNetworkMBps() (out float64, err error) {
	info, err := resourceDBExecutor.Get(NetMBps)
	if err != nil {
		return
	}
	out = info.Value
	return
}

func getNetworkBandwidth() (out float64, err error) {
	info, err := resourceDBExecutor.Get(NetBandwidth)
	if err != nil {
		return
	}
	out = info.Value
	return
}

func getNetworkRTT(ID string) (out float64, err error) {
	info, err := netDBExecutor.Get(ID)
	if err != nil {
		return
	}
	out = info.RTT
	return
}
