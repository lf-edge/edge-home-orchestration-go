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

// Package nativeexecutor provides functions to execute service application in Linux native
package nativeexecutor

import (
	"bufio"
	"errors"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
	"os"
	"os/exec"

	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/executor"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/notification"
)

var (
	logPrefix     = "[nativeexecutor]"
	log           = logmgr.GetInstance()
	nativeexecutor = &NativeExecutor{}
)

// NativeExecutor struct
type NativeExecutor struct {
	executor.ServiceExecutionInfo
	executor.HasClientNotification
}

func init() {
	nativeexecutor.SetNotiImpl(notification.GetInstance())
}

// GetInstance returns the single tone NativeExecutor instance
func GetInstance() *NativeExecutor {
	return nativeexecutor
}

// Execute executes native service application
func (t NativeExecutor) Execute(s executor.ServiceExecutionInfo) (err error) {
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

func (t NativeExecutor) setService() (cmd *exec.Cmd, pid int, err error) {
	if len(t.ParamStr) < 1 {
		err = errors.New("error: empty parameter")
		return
	}
	cmd = exec.Command(t.ParamStr[0], t.ParamStr[1:]...)

	// set "owner" account: need to execute user app
	/*
		execUser, _ := user.Lookup("owner")

		uid, _ := strconv.Atoi(execUser.Uid)
		gid, _ := strconv.Atoi(execUser.Gid)
		groups, _ := execUser.GroupIds()

		var gids = []uint32{}
		for _, i := range groups {
			id, _ := strconv.Atoi(i)
			gids = append(gids, uint32(id))
		}

		log.Printf("uid(%d), gid(%d)", uid, gid)
		log.Printf("groupIds: %v", gids)

		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{
			Uid:    uint32(uid),
			Gid:    uint32(gid),
			Groups: gids,
		}
	*/

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

func (t NativeExecutor) waitService(executeCh <-chan error) (status string, e error) {
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

func (t NativeExecutor) notifyServiceStatus(status string) {
	t.NotiImplIns.InvokeNotification(t.NotificationTargetURL, float64(t.ServiceID), status)
}
