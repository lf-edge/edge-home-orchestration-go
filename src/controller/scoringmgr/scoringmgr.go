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

// Package scoringmgr provides the way to apply specific scoring method for each service application
package scoringmgr

import (
	"github.com/lf-edge/edge-home-orchestration-go/src/common/resourceutil"
	"errors"
	"math"
)

const (
	logPrefix     = "scoringmgr"
	INVALID_SCORE = 0.0
)

// Scoring is the interface to apply application specific scoring functions
type Scoring interface {
	GetScore(ID string) (scoreValue float64, err error)
	GetScoreWithResource(resource map[string]interface{}) (scoreValue float64, err error)
	GetResource(ID string) (resource map[string]interface{}, err error)
}

// ScoringImpl structure
type ScoringImpl struct{}

var (
	constLibStatusInit = 1
	constLibStatusRun  = 2
	constLibStatusDone = true

	scoringIns *ScoringImpl

	resourceIns resourceutil.GetResource
)

func init() {
	scoringIns = new(ScoringImpl)
	resourceIns = &resourceutil.ResourceImpl{}
}

// GetInstance gives the ScoringImpl singletone instance
func GetInstance() *ScoringImpl {
	return scoringIns
}

// GetScore provides score value for specific application on local device
func (ScoringImpl) GetScore(ID string) (scoreValue float64, err error) {
	scoreValue = calculateScore(ID)
	return
}

// GetResource provides resource value for running applications on local device
func (ScoringImpl) GetResource(ID string) (resource map[string]interface{}, err error) {
	resource = make(map[string]interface{})
	cpuUsage, err := resourceIns.GetResource(resourceutil.CPUUsage)
	if err != nil {
		resource["error"] = INVALID_SCORE
		return
	} else {
		resource["cpuUsage"] = cpuUsage
	}

	cpuCount, err := resourceIns.GetResource(resourceutil.CPUCount)
	if err != nil {
		resource["error"] = INVALID_SCORE
		return
	} else {
		resource["cpuCount"] = cpuCount
	}

	cpuFreq, err := resourceIns.GetResource(resourceutil.CPUFreq)
	if err != nil {
		resource["error"] = INVALID_SCORE
		return
	} else {
		resource["cpuFreq"] = cpuFreq
	}

	netBandwidth, err := resourceIns.GetResource(resourceutil.NetBandwidth)
	if err != nil {
		resource["error"] = INVALID_SCORE
		return
	} else {
		resource["netBandwidth"] = netBandwidth
	}

	resourceIns.SetDeviceID(ID)
	rtt, err := resourceIns.GetResource(resourceutil.NetRTT)
	if err != nil {
		resource["error"] = INVALID_SCORE
		return
	} else {
		resource["rtt"] = rtt
	}

	return
}

// GetScoreWithResource provides score value of an edge device
func (ScoringImpl) GetScoreWithResource(resource map[string]interface{}) (scoreValue float64, err error) {
	if _, found := resource["error"]; found {
		return INVALID_SCORE, errors.New("Resource Not Found")
	}

	cpuScore := cpuScore(resource["cpuUsage"].(float64), resource["cpuCount"].(float64), resource["cpuFreq"].(float64))
	netScore := netScore(resource["netBandwidth"].(float64))
	renderingScore := renderingScore(resource["rtt"].(float64))
	return float64(netScore + (cpuScore / 2) + renderingScore), nil
}

func calculateScore(ID string) float64 {
	cpuUsage, err := resourceIns.GetResource(resourceutil.CPUUsage)
	if err != nil {
		return INVALID_SCORE
	}
	cpuCount, err := resourceIns.GetResource(resourceutil.CPUCount)
	if err != nil {
		return INVALID_SCORE
	}
	cpuFreq, err := resourceIns.GetResource(resourceutil.CPUFreq)
	if err != nil {
		return INVALID_SCORE
	}
	cpuScore := cpuScore(cpuUsage, cpuCount, cpuFreq)

	netBandwidth, err := resourceIns.GetResource(resourceutil.NetBandwidth)
	if err != nil {
		return INVALID_SCORE
	}
	netScore := netScore(netBandwidth)

	resourceIns.SetDeviceID(ID)
	rtt, err := resourceIns.GetResource(resourceutil.NetRTT)
	if err != nil {
		return INVALID_SCORE
	}
	renderingScore := renderingScore(rtt)

	return float64(netScore + (cpuScore / 2) + renderingScore)
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
