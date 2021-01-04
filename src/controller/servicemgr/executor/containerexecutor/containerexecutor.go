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

// Package containerexecutor provides functions to execute service application in container enviroment
package containerexecutor

import (
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
	"os"
	"runtime"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/spf13/pflag"

	verifier "github.com/lf-edge/edge-home-orchestration-go/src/controller/securemgr/verifier"
	servicemgr "github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/executor"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/notification"
)

var (
	logPrefix         = "[containerexecutor]"
	log               = logmgr.GetInstance()
	containerExecutor *ContainerExecutor
)

// ContainerExecutor struct
type ContainerExecutor struct {
	executor.ServiceExecutionInfo

	ceImplIns CEImpl
	executor.HasClientNotification
}

func init() {
	containerExecutor = &ContainerExecutor{}

	// @Note : Set Container Executor to docker.io implmentation
	containerExecutor.SetCEImpl(newCEDocker())
	// @Note : Set Notification implementation
	containerExecutor.SetNotiImpl(notification.GetInstance())
}

// GetInstance returns the singletone ContainerExecutor instance
func GetInstance() *ContainerExecutor {
	return containerExecutor
}

// Execute executes container service application
func (c ContainerExecutor) Execute(s executor.ServiceExecutionInfo) error {
	c.ServiceExecutionInfo = s

	log.Println(logPrefix, c.ServiceName, c.ParamStr)
	log.Println(logPrefix, "parameter length :", len(c.ParamStr))
	paramLen := len(c.ParamStr)

	err := verifier.GetInstance().ContainerIsInWhiteList(s.ParamStr[paramLen-1])
	if err != nil {
		log.Println(logPrefix, err.Error())
		return err
	}

	// @Note : Pull docker image
	err = c.ceImplIns.ImagePull(s.ParamStr[paramLen-1])
	if err != nil {
		log.Println(logPrefix, err.Error())
	}

	// @Note : Create containers with converting configuration
	resp, err := c.ceImplIns.Create(convertConfig(s.ParamStr))
	if err != nil {
		log.Println(logPrefix, err.Error())
	} else {
		log.Println(logPrefix, "create container :", resp.ID[:10])
	}

	// @Note : Start container
	err = c.ceImplIns.Start(resp.ID)
	if err != nil {
		log.Println("err :", err)
		return err
	}

	// @Note : get log of container
	out, err := c.ceImplIns.Logs(resp.ID)
	if err != nil {
		log.Println(logPrefix, err.Error())
	} else {
		stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	}

	// @Note : Waiting Container execution status
	var executionStatus string
	statusCh, errCh := c.ceImplIns.Wait(resp.ID, container.WaitConditionNotRunning)
	select {
	case err = <-errCh:
		log.Println(logPrefix, err.Error())
		executionStatus = servicemgr.ConstServiceStatusFailed
	case status := <-statusCh:
		log.Println(logPrefix, "container execution status :", status.StatusCode)
		if status.StatusCode == 0 {
			executionStatus = servicemgr.ConstServiceStatusFinished
		}
	}

	// @Note : make notification
	c.NotiImplIns.InvokeNotification(c.NotificationTargetURL, float64(c.ServiceID), executionStatus)

	// @Note : Remove container after execution
	err = c.ceImplIns.Remove(resp.ID)
	if err != nil {
		log.Println(logPrefix, err.Error())
	}

	return nil
}

// SetCEImpl sets executor implementation
func (c *ContainerExecutor) SetCEImpl(ce CEImpl) {
	c.ceImplIns = ce
}

func convertConfig(paramStr []string) (
	containerConf *container.Config, hostConf *container.HostConfig, networkConf *network.NetworkingConfig) {

	// @Note : initialize getting docker run configurations
	flags := pflag.NewFlagSet(" ", pflag.ContinueOnError)
	copts := addFlags(flags)

	paramLen := len(paramStr)
	param := paramStr[2 : paramLen-1]
	flags.Parse(param)

	conf, _ := parse(flags, copts, runtime.GOOS)
	conf.Config.Image = paramStr[paramLen-1]

	return conf.Config, conf.HostConfig, conf.NetworkingConfig
}
