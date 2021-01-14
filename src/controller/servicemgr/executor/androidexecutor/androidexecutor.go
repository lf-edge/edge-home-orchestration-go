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

// Package androidexecutor provides functions to execute service application in android native
package androidexecutor

import (
	"errors"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
	"os"
	"strings"
	"sync"

	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/executor"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/notification"
)

var (
	logPrefix       = "[androidexecutor]"
	log             = logmgr.GetInstance()
	androidexecutor = &AndroidExecutor{}
)

// Execute callback
type ExecuteCallback interface {
	Execute(packageName string, args string) int
}

// AndroidExecutor struct
type AndroidExecutor struct {
	executor.ServiceExecutionInfo
	executor.HasClientNotification
	executeCB ExecuteCallback
}

func init() {
	androidexecutor.SetNotiImpl(notification.GetInstance())
}

// GetInstance returns the single tone AndroidExecutor instance
func GetInstance() *AndroidExecutor {
	return androidexecutor
}

func (t *AndroidExecutor) SetExecuteCallback(executeCallback ExecuteCallback) {
	t.executeCB = executeCallback
}

// Execute executes android service application
func (t *AndroidExecutor) Execute(s executor.ServiceExecutionInfo) (err error) {
	t.ServiceExecutionInfo = s

	log.Println(logPrefix, t.ServiceName, t.ParamStr)
	log.Println(logPrefix, "parameter length :", len(t.ParamStr))

	result, err := t.setService()
	if err != nil {
		return
	}

	log.Println(logPrefix, "Just ran subprocess [Result] ", result)

	var wait sync.WaitGroup
	wait.Add(1)

	executeCh := make(chan error)
	go func() {
		status, _ := t.waitService(executeCh)
		t.notifyServiceStatus(status)
		wait.Done()
	}()

	switch {
	case result >= 0:
		executeCh <- nil
	default:
		executeCh <- err
	}

	wait.Wait()

	return
}

func (t AndroidExecutor) setService() (result int, err error) {
	if len(t.ParamStr) < 1 {
		err = errors.New("error: empty parameter")
		return
	}

	if nil == t.executeCB {
		log.Println(logPrefix, "Java callback is nil")
		err = errors.New("Failed to execute: Java Callback is nil")
		return
	}
	log.Println(logPrefix, "Invoke java callback with packageName: ", t.ParamStr[0])

	switch len(t.ParamStr) {
	case 1:
		result = t.executeCB.Execute(t.ParamStr[0], "")
	default:
		args := strings.Join(t.ParamStr[1:], " ")
		result = t.executeCB.Execute(t.ParamStr[0], args)
	}

	if result < 0 {
		log.Println(logPrefix, "Failed to execute in java layer")
		err = errors.New("Failed to execute in java layer")
		return
	}
	log.Println(logPrefix, "Successfully executed in java layer")
	return
}

func (t AndroidExecutor) waitService(executeCh <-chan error) (status string, e error) {
	e = <-executeCh

	status = servicemgr.ConstServiceStatusFinished

	if e != nil {
		if e.Error() == os.Kill.String() {
			log.Println(logPrefix, "Success to delete service")
		} else {
			status = servicemgr.ConstServiceStatusFailed
			log.Println(logPrefix, t.ServiceName, "exited with error : ", e)
		}
	} else {
		log.Println(logPrefix, t.ServiceName, "is exited with no error")
	}

	return
}

func (t AndroidExecutor) notifyServiceStatus(status string) {
	t.NotiImplIns.InvokeNotification(t.NotificationTargetURL, float64(t.ServiceID), status)
}
