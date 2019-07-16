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

package resourceutil

import (
	"errors"
	"testing"
	"time"

	commoncpu "common/resourceutil/cpu"
	memutil "github.com/shirou/gopsutil/mem"
	netutil "github.com/vishvananda/netlink"
)

type dummpyLink struct {
	Link
	attrs   netutil.LinkAttrs
	netType string
}

func (d dummpyLink) Attrs() *netutil.LinkAttrs {
	return &d.attrs
}

func (d dummpyLink) Type() string {
	return d.netType
}

type Link interface {
	Attrs() *netutil.LinkAttrs
	Type() string
}

var (
	dummyCPUPercent = []float64{
		10.0,
		10.0,
		10.0,
	}

	dummyCPUPercentResult = 10.0

	dummpyCPUInfoStat1 = commoncpu.InfoStat{
		Mhz: 3800.0,
	}

	dummpyCPUInfoStat2 = commoncpu.InfoStat{
		Mhz: 3800.0,
	}

	dummyCPUFreqResult  = 3800.0
	dummyCPUCountResult = 2.0

	dummpyCPUInfoStatSlice = []commoncpu.InfoStat{
		dummpyCPUInfoStat1,
		dummpyCPUInfoStat2,
	}

	dummyMemAvailableResult = 1.0
	dummyMemFreeResult      = 1.0

	dummyVirtualMemoryStat = memutil.VirtualMemoryStat{
		Available: 1024,
		Free:      1024,
	}

	dummyNetMBpsResult      = 0.0
	dummyNetBandwidthResult = 1000.0

	dummyStatistics = netutil.LinkStatistics{
		RxBytes: 1024,
		TxBytes: 1024,
	}

	dummyNetAttrs = netutil.LinkAttrs{
		Statistics: &dummyStatistics,
		Name:       "eth",
		TxQLen:     1000,
	}

	dummyNetAttrsNotEthernet = netutil.LinkAttrs{
		Statistics: &dummyStatistics,
		Name:       "wifi",
		TxQLen:     1000,
	}

	dummyNetLink = dummpyLink{
		attrs:   dummyNetAttrs,
		netType: "dummyType",
	}

	dummyNetLinkNotEthernet = dummpyLink{
		attrs:   dummyNetAttrsNotEthernet,
		netType: "dummyType",
	}
)

var resourceIns ResourceImpl

func fakeCPUPercent(interval time.Duration, percpu bool) ([]float64, error) {
	return dummyCPUPercent, nil
}

func fakeCPUPercentWithError(interval time.Duration, percpu bool) ([]float64, error) {
	return dummyCPUPercent, errors.New("fakeCPUPercentWithError")
}

func fakeCPUInfo() ([]commoncpu.InfoStat, error) {
	return dummpyCPUInfoStatSlice, nil
}

func fakeCPUInfoWithError() ([]commoncpu.InfoStat, error) {
	return dummpyCPUInfoStatSlice, errors.New("fakeCPUInfoWithError")
}

func fakeVirtualMemory() (*memutil.VirtualMemoryStat, error) {
	return &dummyVirtualMemoryStat, nil
}

func fakeVirtualMemoryWithError() (*memutil.VirtualMemoryStat, error) {
	return &dummyVirtualMemoryStat, errors.New("fakeCPUInfoWithError")
}

func fakeLinkList() ([]netutil.Link, error) {
	linkList := make([]netutil.Link, 0)
	linkList = append(linkList, dummyNetLink)
	return linkList, nil
}

func fakeLinkListExcludeEthernet() ([]netutil.Link, error) {
	linkList := make([]netutil.Link, 0)
	linkList = append(linkList, dummyNetLinkNotEthernet)
	return linkList, nil
}

func fakeLinkListWithError() ([]netutil.Link, error) {
	linkList := make([]netutil.Link, 0)
	linkList = append(linkList, dummyNetLink)
	return linkList, errors.New("fakeLinkListPrevWithError")
}

func TestGetCPUUsage_ExpectedSuccess(t *testing.T) {
	cpu.percent = fakeCPUPercent

	cpuUsage, err := resourceIns.GetResource(CPUUsage)
	if err != nil {
		t.Errorf(err.Error())
	}

	if cpuUsage != dummyCPUPercentResult {
		t.Errorf("%f != %f", cpuUsage, dummyCPUPercentResult)
	}
}

func TestGetCPUUsage_CPUUtilReturnError_ExpectedErrorReturn(t *testing.T) {
	cpu.percent = fakeCPUPercentWithError

	_, err := resourceIns.GetResource(CPUUsage)
	if err == nil {
		t.Errorf("Not working error handling logic")
	}
}

func TestGetCPUFreq_ExpectedSuccess(t *testing.T) {
	cpu.info = fakeCPUInfo

	cpuFreq, err := resourceIns.GetResource(CPUFreq)
	if err != nil {
		t.Errorf(err.Error())
	}

	if cpuFreq != dummyCPUFreqResult {
		t.Errorf("%f != %f", cpuFreq, dummyCPUFreqResult)
	}
}

func TestGetCPUFreq_CPUUtilReturnError_ExpectedErrorReturn(t *testing.T) {
	cpu.info = fakeCPUInfoWithError

	_, err := resourceIns.GetResource(CPUFreq)
	if err == nil {
		t.Errorf("Not working error handling logic")
	}
}

func TestGetCPUCount_ExpectedSuccess(t *testing.T) {
	cpu.info = fakeCPUInfo

	cpuCount, err := resourceIns.GetResource(CPUCount)
	if err != nil {
		t.Errorf(err.Error())
	}

	if cpuCount != dummyCPUCountResult {
		t.Errorf("%f != %f", cpuCount, dummyCPUCountResult)
	}
}

func TestGetCPUCount_CPUUtilReturnError_ExpectedErrorReturn(t *testing.T) {
	cpu.info = fakeCPUInfoWithError

	_, err := resourceIns.GetResource(CPUCount)
	if err == nil {
		t.Errorf("Not working error handling logic")
	}
}

func TestGetMemAvailable_ExpectedSuccess(t *testing.T) {
	mem.virtualMemory = fakeVirtualMemory

	memAvailable, err := resourceIns.GetResource(MemAvailable)
	if err != nil {
		t.Errorf(err.Error())
	}

	if memAvailable != dummyMemAvailableResult {
		t.Errorf("%f != %f", memAvailable, dummyMemAvailableResult)
	}
}

func TestGetMemAvailable_MemUtilReturnError_ExpectedErrorReturn(t *testing.T) {
	mem.virtualMemory = fakeVirtualMemoryWithError

	_, err := resourceIns.GetResource(MemAvailable)
	if err == nil {
		t.Errorf("Not working error handling logic")
	}
}

func TestGetMemFree_ExpectedSuccess(t *testing.T) {
	mem.virtualMemory = fakeVirtualMemory

	memFree, err := resourceIns.GetResource(MemFree)
	if err != nil {
		t.Errorf(err.Error())
	}

	if memFree != dummyMemFreeResult {
		t.Errorf("%f != %f", memFree, dummyMemFreeResult)
	}
}

func TestGetMemFree_MemUtilReturnError_ExpectedErrorReturn(t *testing.T) {
	mem.virtualMemory = fakeVirtualMemoryWithError

	_, err := resourceIns.GetResource(MemFree)
	if err == nil {
		t.Errorf("Not working error handling logic")
	}
}

func TestGetNetMBps_ExpectedSuccess(t *testing.T) {
	net.linkList = fakeLinkList

	netMBps, err := resourceIns.GetResource(NetMBps)
	if err != nil {
		t.Errorf(err.Error())
	}

	if netMBps != dummyNetMBpsResult {
		t.Errorf("%f != %f", netMBps, dummyNetMBpsResult)
	}
}

func TestGetNetMBps_NetUtilReturnError_ExpectedErrorReturn(t *testing.T) {
	net.linkList = fakeLinkListWithError

	_, err := resourceIns.GetResource(NetMBps)
	if err == nil {
		t.Errorf("Not working error handling logic")
	}
}

func TestGetNetBandwidth_ExpectedSuccess(t *testing.T) {
	net.linkList = fakeLinkList

	netBandwidth, err := resourceIns.GetResource(NetBandwidth)
	if err != nil {
		t.Errorf(err.Error())
	}

	if netBandwidth != dummyNetBandwidthResult {
		t.Errorf("%f != %f", netBandwidth, dummyNetBandwidthResult)
	}
}

func TestGetNetBandwidth_NetUtilReturnExcludeEthernetInterface_ExpectedErrorReturn(t *testing.T) {
	net.linkList = fakeLinkListExcludeEthernet

	_, err := resourceIns.GetResource(NetBandwidth)
	if err == nil {
		t.Errorf("Not working error handling logic")
	}
}

func TestGetNetBandwidth_NetUtilReturnError_ExpectedErrorReturn(t *testing.T) {
	net.linkList = fakeLinkListWithError

	_, err := resourceIns.GetResource(NetBandwidth)
	if err == nil {
		t.Errorf("Not working error handling logic")
	}
}
