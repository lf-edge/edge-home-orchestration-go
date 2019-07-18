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
	"errors"

	types "common/types/configuremgrtypes"
)

const logPrefix = "scoringmgr"

// Scoring is the interface to apply application specific scoring functions
type Scoring interface {
	AddScoring(service types.ServiceInfo) (err error)
	RemoveScoring(appName string) (err error)
	RemoveAllScoring() (err error)

	GetScore(name string) (scoreValue float64, err error)
}

// ScoringImpl structure
type ScoringImpl struct{}

type rater struct {
	types.ServiceInfo

	scoreValue float64
}

var (
	constLibStatusInit = 1
	constLibStatusRun  = 2
	constLibStatusDone = true

	scoringIns *ScoringImpl
	raters     map[string]*rater
)

func init() {
	scoringIns = new(ScoringImpl)
	raters = make(map[string]*rater)
}

// GetInstance gives the ScoringImpl singletone instance
func GetInstance() *ScoringImpl {
	return scoringIns
}

// AddScoring for adding scoring object for resource scoring
func (ScoringImpl) AddScoring(service types.ServiceInfo) (err error) {
	rater := new(rater)

	if _, exist := raters[service.ServiceName]; exist {
		err = errors.New("duplicate service scoring objects have already been added")
		return
	}

	rater.ServiceInfo = service

	raters[rater.ServiceName] = rater
	return
}

//RemoveScoring for removal of scoring object with appName
func (ScoringImpl) RemoveScoring(appName string) (err error) {
	rater, exist := raters[appName]
	if !exist {
		err = errors.New("no scoring object matching appname")
		return
	}
	defer rater.ScoringFunc.Close()

	delete(raters, appName)
	return
}

//RemoveAllScoring for removal all of scoring object
func (s ScoringImpl) RemoveAllScoring() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("panic in RemoveAllScoring()")
		}
	}()

	for name := range raters {
		s.RemoveScoring(name)
	}
	return
}

// GetScore provides score value for specific application on local device
func (ScoringImpl) GetScore(name string) (scoreValue float64, err error) {
	rater, exist := raters[name]
	if !exist {
		err = errors.New("invalid library name")
		return
	}

	scoreValue = rater.ScoringFunc.Run()
	return
}
