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
	"bufio"
	"errors"
	"log"
	"os"
	"os/exec"

	"controller/servicemgr"
	"controller/servicemgr/executor"
	"controller/servicemgr/notification"
)

var (
	logPrefix       = "[androidexecutor]"
	androidexecutor = &AndroidExecutor{}
	adbPath         = "/system/bin/am"
)

// AndroidExecutor struct
type AndroidExecutor struct {
	executor.ServiceExecutionInfo
	executor.HasClientNotification
}

func init() {
	androidexecutor.SetNotiImpl(notification.GetInstance())
}

// GetInstance returns the single tone AndroidExecutor instance
func GetInstance() *AndroidExecutor {
	return androidexecutor
}

// Execute executes android service application
func (t AndroidExecutor) Execute(s executor.ServiceExecutionInfo) (err error) {
	t.ServiceExecutionInfo = s

	log.Println(logPrefix, t.ServiceName, t.ParamStr)
	log.Println(logPrefix, "parameter length :", len(t.ParamStr))

	cmd, pid, err := t.setService()
	if err != nil {
		return
	}

	log.Println(logPrefix, "Just ran subprocess ", pid)

	executeCh := make(chan error)
	go func() {
		executeCh <- cmd.Wait()
	}()

	status, err := t.waitService(executeCh)
	t.notifyServiceStatus(status)

	return
}

func (t AndroidExecutor) setService() (cmd *exec.Cmd, pid int, err error) {
	if len(t.ParamStr) < 1 {
		err = errors.New("error: empty parameter")
		return
	}

	adbStart := []string{"start", "-n"}
	adbStart = append(adbStart, t.ParamStr...)
	log.Println(logPrefix, "Adb start cmd: ", adbStart)
	cmd = exec.Command(adbPath, adbStart[0:]...)

	stdout, _ := cmd.StdoutPipe()
	err = cmd.Start()
	if err != nil {
		log.Println(logPrefix, err.Error())
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := scanner.Text()
		log.Println(m)
	}

	pid = cmd.Process.Pid

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
