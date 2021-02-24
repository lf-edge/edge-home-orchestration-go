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

// Package errormsg implements error messages for orchestration
package errormsg

import (
	"errors"
	"strconv"
)

// Define Orchestration Error
const (
	ErrorNotReadyOrchestrationInit = ((1 + iota) ^ -1) + 1
	ErrorNoDeviceReturn
	ErrorNoNetworkInterface
)

var orchestrationErrorString = [...]string{
	"",
	"Please wait until Orchestration init function has been completed",
	"No Device Is Return",
	"No Network Interface",
}

// ToString converts error const to string
func ToString(args interface{}) string {

	switch args.(type) {
	case int:
		v := args.(int)
		return orchestrationErrorString[v*-1]
	case error:
		err := args.(error)
		v, _ := strconv.Atoi(err.Error())
		return orchestrationErrorString[v*-1]
	}

	return "NOT SUPPORT TYPE"
}

// ToError converts int to error string
func ToError(orcheError int) error {
	return errors.New(strconv.Itoa(orcheError))
}

// ToInt converts error to int
func ToInt(err error) int {
	v, _ := strconv.Atoi(err.Error())
	return v
}
