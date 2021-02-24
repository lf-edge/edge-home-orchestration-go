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

package containerexecutor

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

// CEImpl is the interface implemented by container execution functions
type CEImpl interface {
	Create(conf *container.Config, hostConf *container.HostConfig, networkConf *network.NetworkingConfig) (container.ContainerCreateCreatedBody, error)
	Remove(id string) error
	Start(id string) error
	Wait(id string, condition container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error)
	Logs(id string) (io.ReadCloser, error)
	ImagePull(image string) error

	// @Note : When below api is need to implements, it will be opened
	// PS() ([]types.Container, error)
	// Stop(id string, timeout *time.Duration) error
	// Events() (<-chan events.Message, <-chan error)
	// ImageTag(source string, target string) error
}

// CEDocker structure
type CEDocker struct {
	ctx    context.Context
	cli    *client.Client
}

func newCEDocker() (ceDocker *CEDocker) {
	cli, err := client.NewEnvClient()
	if err == nil {
		ceDocker = &CEDocker{context.Background(), cli}
	}
	return
}

// Create is to create container
func (ce CEDocker) Create(conf *container.Config, hostConf *container.HostConfig, networkConf *network.NetworkingConfig) (resp container.ContainerCreateCreatedBody, err error) {
	resp, err = ce.cli.ContainerCreate(ce.ctx, conf, hostConf, networkConf, "")
	return
}

// Remove is to remove container
func (ce CEDocker) Remove(id string) (err error) {
	return ce.cli.ContainerRemove(ce.ctx, id, types.ContainerRemoveOptions{})
}

// Start is to start container
func (ce CEDocker) Start(id string) (err error) {
	return ce.cli.ContainerStart(ce.ctx, id, types.ContainerStartOptions{})
}

// Wait is to wait container
func (ce CEDocker) Wait(id string, condition container.WaitCondition) (statusCh <-chan container.ContainerWaitOKBody, errCh <-chan error) {
	return ce.cli.ContainerWait(ce.ctx, id, container.WaitConditionNotRunning)
}

// Logs is to logs container
func (ce CEDocker) Logs(id string) (io.ReadCloser, error) {
	opts := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow: true,
	}
	return ce.cli.ContainerLogs(ce.ctx, id, opts)
}

// ImagePull is to pull container images
func (ce CEDocker) ImagePull(image string) (err error) {
	reader, err := ce.cli.ImagePull(ce.ctx, image, types.ImagePullOptions{})
	if err == nil {
		io.Copy(os.Stdout, reader)
	}

	return
}

// PS function
// func (ce CEDocker) PS() ([]types.Container, error) {
// 	return ce.cli.ContainerList(ce.ctx, types.ContainerListOptions{})
// }

// Stop function
// func (ce CEDocker) Stop(id string, timeout *time.Duration) (err error) {
// 	return ce.cli.ContainerStop(ce.ctx, id, timeout)
// }

// Events function
// func (ce CEDocker) Events() (<-chan events.Message, <-chan error) {
// 	return ce.cli.Events(ce.ctx, types.EventsOptions{})
// }// Events function
// func (ce CEDocker) Events() (<-chan events.Message, <-chan error) {
// 	return ce.cli.Events(ce.ctx, types.EventsOptions{})
// }

// ImageTag function
// func (ce CEDocker) ImageTag(source string, target string) (err error) {
// 	return ce.cli.ImageTag(ce.ctx, source, target)
// }
