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
	"log"
	"strings"
	"time"

	"common/errors"
	commoncpu "common/resourceutil/cpu"
	memutil "github.com/shirou/gopsutil/mem"
	netutil "github.com/vishvananda/netlink"
)

// ResourceImpl is implementation for resourceutil interface
type ResourceImpl struct{}

type cpuUtil struct {
	percent func(interval time.Duration, percpu bool) ([]float64, error)
	info    func() ([]commoncpu.InfoStat, error)
}

type memUtil struct {
	virtualMemory func() (*memutil.VirtualMemoryStat, error)
}

type netUtil struct {
	linkList func() ([]netutil.Link, error)
}

var (
	cpu = cpuUtil{}
	mem = memUtil{}
	net = netUtil{}
)

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

	logPrefix = "resourceutil"
)

// Command is an interface to control resource monitoring operations
type Command interface {
	Run() float64
	Close()
}

// GetResource is an interface to get reource
type GetResource interface {
	GetResource(string) (float64, error)
}

func init() {
	cpu.info = commoncpu.Info
	cpu.percent = commoncpu.Percent
	mem.virtualMemory = memutil.VirtualMemory
	net.linkList = netutil.LinkList
}

// GetResource returns a resource value that matches resourceName
func (ResourceImpl) GetResource(resourceName string) (float64, error) {
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
		return getNetworkRTT()
	default:
		return 0.0, errors.NotSupport{Message: "Not suppoted resource name"}
	}
}

func getCPUUsage() (out float64, err error) {
	cpus, err := cpu.percent(time.Second, true)
	if err != nil {
		log.Println(logPrefix, "usage of cpu is fail : ", err.Error())
		err = errors.SystemError{Message: err.Error()}
		return
	}

	for _, cpu := range cpus {
		out += float64(cpu)
	}
	out /= float64(len(cpus))

	return
}

func getCPUFreq() (out float64, err error) {
	infos, err := cpu.info()
	if err != nil {
		log.Println(logPrefix, "cpu.Info() fail : ", err.Error())
		err = errors.SystemError{Message: err.Error()}
		return
	}

	out = infos[0].Mhz
	return
}

func getCPUCount() (out float64, err error) {
	infos, err := cpu.info()
	if err != nil {
		log.Println(logPrefix, "cpu info getting fail : ", err.Error())
		err = errors.SystemError{Message: err.Error()}
		return
	}

	out = float64(len(infos))
	return
}

func getMemoryAvailable() (out float64, err error) {
	memStat, err := mem.virtualMemory()
	if err != nil {
		log.Println(logPrefix, "mem info getting fail : ", err.Error())
		err = errors.SystemError{Message: err.Error()}
		return
	}

	out = float64(memStat.Available) / 1024
	return
}

func getMemoryFree() (out float64, err error) {
	memStat, err := mem.virtualMemory()
	if err != nil {
		log.Println(logPrefix, "mem info getting fail : ", err.Error())
		err = errors.SystemError{Message: err.Error()}
		return
	}

	out = float64(memStat.Free) / 1024
	return
}

func getNetworkMBps() (out float64, err error) {
	linklist, err := net.linkList()
	if err != nil {
		log.Println(logPrefix, "network link getting fail : ", err.Error())
		err = errors.SystemError{Message: err.Error()}
		return
	}

	var prevTotalBytes, nextTotalBytes uint64

	for _, link := range linklist {
		prevTotalBytes += link.Attrs().Statistics.RxBytes
		prevTotalBytes += link.Attrs().Statistics.TxBytes
	}

	time.Sleep(1 * time.Second)

	linklist, err = net.linkList()
	if err != nil {
		log.Println(logPrefix, "network link getting fail : ", err.Error())
		err = errors.SystemError{Message: err.Error()}
		return
	}

	for _, link := range linklist {
		nextTotalBytes += link.Attrs().Statistics.RxBytes
		nextTotalBytes += link.Attrs().Statistics.TxBytes
	}

	out = float64((nextTotalBytes - prevTotalBytes)) / 1024 / 1024
	return
}

func getNetworkBandwidth() (out float64, err error) {
	linklist, err := net.linkList()
	if err != nil {
		log.Println(logPrefix, "network link getting fail : ", err.Error())
		err = errors.SystemError{Message: err.Error()}
		return
	}

	var count, total int

	for _, link := range linklist {
		if strings.Contains(link.Attrs().Name, "eth") ||
			strings.Contains(link.Attrs().Name, "enp") {
			count++
			total += link.Attrs().TxQLen
		}
	}

	if count == 0 {
		err = errors.SystemError{Message: "Not matched network interface"}
		return
	}

	out = float64(total / count)
	return
}

// TODO RTT with candidate target
func getNetworkRTT() (out float64, err error) {
	return
}
