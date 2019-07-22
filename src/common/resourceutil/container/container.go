// Package container implements specific resourceutil functions for docker container
package container

import (
	"common/resourceutil"
	"math"
)

// Getter is the implementation for container
type Getter struct{}

var resourceIns resourceutil.GetResource

func init() {
	resourceIns = &resourceutil.ResourceImpl{}
}

// Run gets the infomations of device resource usage
func (Getter) Run(ID string) float64 {
	cpuUsage, err := resourceIns.GetResource(resourceutil.CPUUsage)
	if err != nil {
		return 0.0
	}
	cpuCount, err := resourceIns.GetResource(resourceutil.CPUCount)
	if err != nil {
		return 0.0
	}
	cpuFreq, err := resourceIns.GetResource(resourceutil.CPUFreq)
	if err != nil {
		return 0.0
	}
	cpuScore := cpuScore(cpuUsage, cpuCount, cpuFreq)

	netBandwidth, err := resourceIns.GetResource(resourceutil.NetBandwidth)
	if err != nil {
		return 0.0
	}
	netScore := netScore(netBandwidth)

	resourceIns.SetDeviceID(ID)
	rtt, err := resourceIns.GetResource(resourceutil.NetRTT)
	if err != nil {
		return 0.0
	}
	renderingScore := renderingScore(rtt)

	return float64(netScore + (cpuScore / 2) + renderingScore)
}

// Close stops device resource monitoring
func (Getter) Close() {
}

func netScore(bandWidth float64) (score float64) {
	return 1 / (8770 * math.Pow(bandWidth, -0.9))
}

func cpuScore(usage float64, count float64, freq float64) (score float64) {
	return ((1 / (5.66 * math.Pow(freq, -0.66))) +
		(1 / (3.22 * math.Pow(usage, -0.241))) +
		(1 / (4 * math.Pow(count, -0.3)))) / 3
}

func renderingScore(rtt float64) (score float64) {
	if rtt <= 0 {
		score = 0
	} else {
		score = 0.77 * math.Pow(rtt, -0.43)
	}
	return
}
