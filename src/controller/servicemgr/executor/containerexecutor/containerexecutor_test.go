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
	"errors"
	"io/ioutil"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/mount"

	"github.com/docker/docker/api/types/blkiodev"

	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/executor"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/executor/containerexecutor/mocks"
	notificationMock "github.com/lf-edge/edge-home-orchestration-go/src/controller/servicemgr/notification/mocks"
	clientApiMock "github.com/lf-edge/edge-home-orchestration-go/src/restinterface/client/mocks"

	"github.com/docker/go-units"
	gomock "github.com/golang/mock/gomock"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

var (
	imageName   = "hello-wrold"
	containerID = "fakeimage1234"
	networkName = "fakenet"

	serviceInfo = executor.ServiceExecutionInfo{
		ServiceID: uint64(1), ServiceName: "alpine",
		ParamStr: []string{"docker", "run", "-v", "/var/run/:/var/run:rw", "-p", "8080:8080", "alpine"}}

	conf = &container.Config{
		Image: "alpine",
		// Cmd: s.ParamStr,
		Cmd: []string{"echo", "hello world"},
		Tty: true,
	}
	resp = container.ContainerCreateCreatedBody{
		ID: containerID,
	}

	containerList = []types.Container{
		{ID: containerID, Image: "alpine"},
	}

	statusChan = make(chan container.ContainerWaitOKBody)
	errCh      = make(chan error)

	reader     = strings.NewReader("Hello,World")
	readCloser = ioutil.NopCloser(reader)
)

func initializeMock(t *testing.T) (*mocks.MockCEImpl, *notificationMock.MockNotification, *clientApiMock.MockClienter) {
	t.Helper()

	ctrl := gomock.NewController(t)
	con := mocks.NewMockCEImpl(ctrl)

	client := clientApiMock.NewMockClienter(ctrl)
	noti := notificationMock.NewMockNotification(ctrl)

	return con, noti, client
}

func TestSetClient(t *testing.T) {
	cExecutor := GetInstance()

	_, noti, client := initializeMock(t)

	noti.EXPECT().SetClient(gomock.Any()).DoAndReturn(
		func(clientParam *clientApiMock.MockClienter) {
			if clientParam != client {
				t.Fail()
			}
		},
	)

	cExecutor.SetNotiImpl(noti)
	cExecutor.SetClient(client)
}
func TestExecute(t *testing.T) {
	cExecutor := GetInstance()
	con, noti, _ := initializeMock(t)

	gomock.InOrder(
		con.EXPECT().ImagePull(gomock.Any()).Return(nil),
		con.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil),
		con.EXPECT().Start(containerID).Return(nil),
		con.EXPECT().Logs(containerID).Return(readCloser, nil),
		con.EXPECT().Wait(containerID, container.WaitConditionNotRunning).Return(statusChan, errCh),
		noti.EXPECT().InvokeNotification(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes(),
		con.EXPECT().Remove(containerID),
	)

	cExecutor.SetCEImpl(con)
	cExecutor.SetNotiImpl(noti)

	var wait sync.WaitGroup
	wait.Add(1)

	go func() {
		err := cExecutor.Execute(serviceInfo)
		if err != nil {
			t.Fail()
		}
		wait.Done()
	}()

	statusChan <- container.ContainerWaitOKBody{StatusCode: 0}
	wait.Wait()
}
func TestExecuteFailedStartInvokedError(t *testing.T) {
	cExecutor := GetInstance()
	con, noti, _ := initializeMock(t)

	gomock.InOrder(
		con.EXPECT().ImagePull(gomock.Any()).Return(nil),
		con.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil),
		con.EXPECT().Start(containerID).Return(errors.New("invoked error")),
	)

	// cExecutor.SetClient(client)
	cExecutor.SetCEImpl(con)
	cExecutor.SetNotiImpl(noti)

	var wait sync.WaitGroup
	wait.Add(1)

	go func() {
		err := cExecutor.Execute(serviceInfo)
		if err == nil {
			t.Fail()
		}
		log.Println(err.Error())
		wait.Done()
	}()

	wait.Wait()
}

func TestExecuteWaitInvokedError(t *testing.T) {
	cExecutor := GetInstance()
	con, noti, _ := initializeMock(t)

	gomock.InOrder(
		con.EXPECT().ImagePull(gomock.Any()).Return(nil),
		con.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil),
		con.EXPECT().Start(containerID).Return(nil),
		con.EXPECT().Logs(containerID).Return(readCloser, nil),
		con.EXPECT().Wait(containerID, container.WaitConditionNotRunning).Return(statusChan, errCh),
		noti.EXPECT().InvokeNotification(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes(),
		con.EXPECT().Remove(containerID),
	)

	cExecutor.SetCEImpl(con)
	cExecutor.SetNotiImpl(noti)

	var wait sync.WaitGroup
	wait.Add(1)

	go func() {
		err := cExecutor.Execute(serviceInfo)
		if err != nil && strings.Compare(err.Error(), "wait error") != 0 {
			t.Error()
		}
		wait.Done()
	}()

	errCh <- errors.New("wait error")
	wait.Wait()
}

func TestSuccessConvertConfigWithAttach(t *testing.T) {
	validStr := []string{"docker", "run", "-a", "stdin", "-a", "stdout", "-a", "stderr", imageName}
	container, _, _ := convertConfig(validStr)
	if !(container.AttachStdin && container.AttachStdout && container.AttachStderr) {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithBlkIO(t *testing.T) {
	validStr := []string{"docker", "run", "--blkio-weight", "10", imageName}
	_, host, _ := convertConfig(validStr)
	if host.Resources.BlkioWeight != 10 {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithBlkIODev(t *testing.T) {
	validStr := []string{"docker", "run", "--blkio-weight-device", "/dev/sda:10", imageName}
	_, host, _ := convertConfig(validStr)
	for _, dev := range host.Resources.BlkioWeightDevice {
		if strings.Compare(dev.Path, "/dev/sda") != 0 || dev.Weight != 10 {
			t.Fail()
		}
	}
}

func TestSuccessConvertConfigWithCapAdd(t *testing.T) {
	validStr := []string{"docker", "run", "--cap-add", "ALL", imageName}
	_, host, _ := convertConfig(validStr)
	for _, capAddItem := range host.CapAdd {
		if strings.Compare(capAddItem, "ALL") != 0 {
			t.Fail()
		}
	}
}

func TestSuccessConvertConfigWithCapDrop(t *testing.T) {
	validStr := []string{"docker", "run", "--cap-drop", "SYS_ADMIN", imageName}
	_, host, _ := convertConfig(validStr)
	for _, capDropItem := range host.CapDrop {
		if strings.Compare(capDropItem, "SYS_ADMIN") != 0 {
			t.Fail()
		}
	}
}

func TestSuccessConvertConfigWithCgroupParent(t *testing.T) {
	validStr := []string{"docker", "run", "--cgroup-parent", "custom_cgroup", imageName}
	_, host, _ := convertConfig(validStr)
	if strings.Compare(host.CgroupParent, "custom_cgroup") != 0 {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithCIDFile(t *testing.T) {
	validStr := []string{"docker", "run", "--cidfile", containerID, imageName}
	_, host, _ := convertConfig(validStr)

	if strings.Compare(host.ContainerIDFile, containerID) != 0 {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithCPUOption(t *testing.T) {
	t.Run("CPUPeriod", func(t *testing.T) {
		validStr := []string{"docker", "run", "--cpu-period", "50000", imageName}
		_, host, _ := convertConfig(validStr)

		if host.Resources.CPUPeriod != int64(50000) {
			t.Fail()
		}
	})
	t.Run("CPUQuota", func(t *testing.T) {
		validStr := []string{"docker", "run", "--cpu-quota", "25000", imageName}
		_, host, _ := convertConfig(validStr)

		if host.Resources.CPUQuota != int64(25000) {
			t.Fail()
		}
	})
	t.Run("CPURealtimePeriod", func(t *testing.T) {
		validStr := []string{"docker", "run", "--cpu-rt-period", "50000", imageName}
		_, host, _ := convertConfig(validStr)

		if host.Resources.CPURealtimePeriod != int64(50000) {
			t.Fail()
		}
	})
	t.Run("CPURealtimeRunTime", func(t *testing.T) {
		validStr := []string{"docker", "run", "--cpu-rt-runtime", "50000", imageName}
		_, host, _ := convertConfig(validStr)

		if host.Resources.CPURealtimeRuntime != int64(50000) {
			t.Fail()
		}
	})
	t.Run("CPUShares", func(t *testing.T) {
		validStr := []string{"docker", "run", "--cpu-shares", "2", imageName}
		_, host, _ := convertConfig(validStr)

		if host.Resources.CPUShares != 2 {
			t.Fail()
		}
	})
	t.Run("CPUs", func(t *testing.T) {
		validStr := []string{"docker", "run", "--cpus", "0.5", imageName}
		_, host, _ := convertConfig(validStr)

		if host.Resources.NanoCPUs != int64(500000000) {
			t.Fail()
		}
	})
	t.Run("CPUSetCPUs", func(t *testing.T) {
		validStr := []string{"docker", "run", "--cpuset-cpus", "0-2", imageName}
		_, host, _ := convertConfig(validStr)

		if strings.Compare(host.Resources.CpusetCpus, "0-2") != 0 {
			t.Fail()
		}
	})
	t.Run("CPUSetMems", func(t *testing.T) {
		validStr := []string{"docker", "run", "--cpuset-mems", "1,3", imageName}
		_, host, _ := convertConfig(validStr)

		if strings.Compare(host.Resources.CpusetMems, "1,3") != 0 {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigDeviceOption(t *testing.T) {
	t.Run("Device", func(t *testing.T) {
		validStr := []string{"docker", "run", "--device", "/dev/sda1:/dev/sdb1", imageName}
		_, host, _ := convertConfig(validStr)

		expectDevice := []container.DeviceMapping{
			{PathOnHost: "/dev/sda1", PathInContainer: "/dev/sdb1", CgroupPermissions: "rwm"}}

		if reflect.DeepEqual(host.Devices, expectDevice) != true {
			t.Fail()
		}
	})
	t.Run("DeviceCGroupRule", func(t *testing.T) {
		validStr := []string{"docker", "run", "--device-cgroup-rule", "c 7:* rmw", imageName}
		_, host, _ := convertConfig(validStr)

		if reflect.DeepEqual(
			host.DeviceCgroupRules, []string{"c 7:* rmw"}) != true {
			t.Fail()
		}
	})
	t.Run("DeviceReadBps", func(t *testing.T) {
		validStr := []string{"docker", "run", "--device-read-bps", "/dev/sda:1mb", imageName}
		_, host, _ := convertConfig(validStr)

		expectDevice := []*blkiodev.ThrottleDevice{{Path: "/dev/sda", Rate: uint64(1048576)}}

		if reflect.DeepEqual(
			// @Note : 1mb = 1048576 bytes
			host.BlkioDeviceReadBps, expectDevice) != true {
			t.Fail()
		}
	})
	t.Run("DeviceWriteBps", func(t *testing.T) {
		validStr := []string{"docker", "run", "--device-write-bps", "/dev/sda:1mb", imageName}
		_, host, _ := convertConfig(validStr)

		expectDevice := []*blkiodev.ThrottleDevice{{Path: "/dev/sda", Rate: uint64(1048576)}}

		if reflect.DeepEqual(
			// @Note : 1mb = 1048576 bytes
			host.BlkioDeviceWriteBps, expectDevice) != true {
			t.Fail()
		}
	})
	t.Run("DeviceReadIOps", func(t *testing.T) {
		validStr := []string{"docker", "run", "--device-read-iops", "/dev/sda:100", imageName}
		_, host, _ := convertConfig(validStr)

		expectDevice := []*blkiodev.ThrottleDevice{{Path: "/dev/sda", Rate: uint64(100)}}

		if reflect.DeepEqual(
			host.BlkioDeviceReadIOps, expectDevice) != true {
			t.Fail()
		}
	})
	t.Run("DeviceWriteIOps", func(t *testing.T) {
		validStr := []string{"docker", "run", "--device-write-iops", "/dev/sda:100", imageName}
		_, host, _ := convertConfig(validStr)

		expectDevice := []*blkiodev.ThrottleDevice{{Path: "/dev/sda", Rate: uint64(100)}}

		if reflect.DeepEqual(
			host.BlkioDeviceWriteIOps, expectDevice) != true {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigDNSOption(t *testing.T) {
	t.Run("DNS", func(t *testing.T) {
		validStr := []string{"docker", "run", "--dns", "8.8.8.8", imageName}
		_, host, _ := convertConfig(validStr)

		if reflect.DeepEqual(host.DNS, []string{"8.8.8.8"}) != true {
			t.Fail()
		}
	})
	t.Run("DNSOption", func(t *testing.T) {
		// @Note : allow for both "--dns-opt" and "--dns-option", although the latter is the recommended way.
		validStr := []string{"docker", "run", "--dns-option", "nameserver 127.0.0.53", imageName}
		_, host, _ := convertConfig(validStr)

		if reflect.DeepEqual(host.DNSOptions, []string{"nameserver 127.0.0.53"}) != true {
			t.Fail()
		}
	})
	t.Run("DNSSearch", func(t *testing.T) {
		validStr := []string{"docker", "run", "--dns-search", "example.com", imageName}
		_, host, _ := convertConfig(validStr)

		if reflect.DeepEqual(host.DNSSearch, []string{"example.com"}) != true {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigWithEntryPoint(t *testing.T) {
	validStr := []string{"docker", "run", "--entrypoint", "/bin/bash", imageName}
	container, _, _ := convertConfig(validStr)
	log.Println(container.Entrypoint)
	for _, entrypoint := range container.Entrypoint {
		if strings.Compare(entrypoint, "/bin/bash") != 0 {
			t.Fail()
		}
	}
}

func TestSuccessConvertConfigEnvOption(t *testing.T) {
	t.Run("Env", func(t *testing.T) {
		validStr := []string{"docker", "run", "--env", "MYSQL_ROOT_PASSWORD=password", imageName}
		container, _, _ := convertConfig(validStr)
		envs := container.Env
		for _, env := range envs {
			if strings.Contains(env, "MYSQL_ROOT_PASSWORD") != true {
				t.Fail()
			}
		}
	})
	t.Run("EnvFile", func(t *testing.T) {
		validStr := []string{"docker", "run", "--env-file", "/etc/environment", imageName}
		container, _, _ := convertConfig(validStr)

		envs := container.Env
		testSuccessFlag := false

		log.Println(envs)
		for _, env := range envs {
			if strings.Contains(env, "PATH") == true {
				testSuccessFlag = true
				break
			} else {
				continue
			}
		}

		if testSuccessFlag != true {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigWithExpose(t *testing.T) {
	validStr := []string{"docker", "run", "--expose", "3306", imageName}
	container, _, _ := convertConfig(validStr)

	keys := reflect.ValueOf(container.ExposedPorts).MapKeys()

	if strings.Contains(keys[0].String(), "3306") != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithGroupAdd(t *testing.T) {
	validStr := []string{"docker", "run", "--group-add", "audio", "--group-add", "nogroup", imageName}
	_, host, _ := convertConfig(validStr)

	if reflect.DeepEqual(host.GroupAdd, []string{"audio", "nogroup"}) != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigHealthOption(t *testing.T) {
	t.Run("HealthCmd", func(t *testing.T) {
		validStr := []string{"docker", "run", "--health-cmd", "stat /etc/passwd || exit 1", imageName}
		container, _, _ := convertConfig(validStr)

		expectCmd := []string{"CMD-SHELL", "stat /etc/passwd || exit 1"}

		if reflect.DeepEqual(container.Healthcheck.Test, expectCmd) != true {
			t.Fail()
		}
	})
	t.Run("HealthInterval", func(t *testing.T) {
		validStr := []string{"docker", "run", "--health-interval", "2s", imageName}
		container, _, _ := convertConfig(validStr)

		if container.Healthcheck.Interval != time.Second*2 {
			t.Fail()
		}
	})
	t.Run("HealthRetries", func(t *testing.T) {
		validStr := []string{"docker", "run", "--health-retries", "2", imageName}
		container, _, _ := convertConfig(validStr)

		if container.Healthcheck.Retries != 2 {
			t.Fail()
		}

	})
	t.Run("HealthStartPeriod", func(t *testing.T) {
		validStr := []string{"docker", "run", "--health-start-period", "2m", imageName}
		container, _, _ := convertConfig(validStr)

		if container.Healthcheck.StartPeriod != time.Minute*2 {
			t.Fail()
		}
	})
	t.Run("HealthTimeout", func(t *testing.T) {
		validStr := []string{"docker", "run", "--health-timeout", "2ms", imageName}
		container, _, _ := convertConfig(validStr)

		if container.Healthcheck.Timeout != time.Millisecond*2 {
			t.Fail()
		}
	})
	t.Run("NoHealthCheck", func(t *testing.T) {
		// validStr := []string{"docker", "run", "--no-healthcheck", "true", imageName}
		validStr := []string{"docker", "run", "--no-healthcheck", imageName}
		container, _, _ := convertConfig(validStr)

		if reflect.DeepEqual(container.Healthcheck.Test, []string{"NONE"}) != true {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigWithHostname(t *testing.T) {
	validStr := []string{"docker", "run", "--hostname", "test", imageName}
	container, _, _ := convertConfig(validStr)

	if strings.Compare(container.Hostname, "test") != 0 {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithInit(t *testing.T) {
	validStr := []string{"docker", "run", "--init", imageName}
	_, host, _ := convertConfig(validStr)

	if *host.Init != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithInteractive(t *testing.T) {
	validStr := []string{"docker", "run", "-i", imageName}
	container, _, _ := convertConfig(validStr)

	if container.AttachStdin != true {
		t.Fail()
	}

	validStr = []string{"docker", "run", "--interactive", imageName}
	container, _, _ = convertConfig(validStr)

	if container.AttachStdin != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigIP(t *testing.T) {
	t.Run("IP", func(t *testing.T) {
		validStr := []string{"docker", "run", "--network", networkName, "--ip", "12.34.56.78", imageName}
		_, _, network := convertConfig(validStr)

		networkConf := network.EndpointsConfig[networkName]
		if strings.Compare(networkConf.IPAMConfig.IPv4Address, "12.34.56.78") != 0 {
			t.Fail()
		}
	})
	t.Run("IP6", func(t *testing.T) {
		validStr := []string{"docker", "run", "--network", networkName, "--ip6", "2001:db8::33", imageName}
		_, _, network := convertConfig(validStr)

		networkConf := network.EndpointsConfig[networkName]

		if strings.Compare(networkConf.IPAMConfig.IPv6Address, "2001:db8::33") != 0 {
			t.Fail()
		}
	})
	t.Run("LinkLocalIP", func(t *testing.T) {
		// TODO
		validStr := []string{"docker", "run", "--network", networkName, "--link-local-ip", "12.34.56.78", "--link-local-ip", "2001:db8::33", imageName}
		_, _, network := convertConfig(validStr)

		networkConf := network.EndpointsConfig[networkName]

		expectIP := []string{"12.34.56.78", "2001:db8::33"}

		if reflect.DeepEqual(networkConf.IPAMConfig.LinkLocalIPs, expectIP) != true {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigWithIPC(t *testing.T) {
	validStr := []string{"docker", "run", "--ipc", "host", imageName}
	_, host, _ := convertConfig(validStr)

	if host.IpcMode.IsHost() != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithIsolation(t *testing.T) {
	// @Note : --isolation has 1 of 3 options [default, process, hyperv]
	validStr := []string{"docker", "run", "--isolation", "process", imageName}
	_, host, _ := convertConfig(validStr)

	if reflect.DeepEqual(host.Isolation, container.Isolation("process")) != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigLabelOption(t *testing.T) {
	t.Run("Label", func(t *testing.T) {
		validStr := []string{"docker", "run", "--label", "com.example.foo=bar", imageName}
		container, _, _ := convertConfig(validStr)

		expectMap := make(map[string]string, 1)
		expectMap["com.example.foo"] = "bar"

		if reflect.DeepEqual(container.Labels, expectMap) != true {
			t.Fail()
		}
	})
	t.Run("LabelFile", func(t *testing.T) {
		validStr := []string{"docker", "run", "--label-file", "./test/labels", imageName}
		container, _, _ := convertConfig(validStr)

		expectMap := make(map[string]string, 2)
		expectMap["com.example.label1"] = "a label"
		expectMap["com.example.label2"] = "a label"

		if reflect.DeepEqual(container.Labels, expectMap) != true {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigWithLink(t *testing.T) {
	validStr := []string{"docker", "run", "--link", "edge-orchestration:edge", imageName}
	_, host, _ := convertConfig(validStr)

	links := []string{"edge-orchestration:edge"}

	if reflect.DeepEqual(host.Links, links) != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigLogOption(t *testing.T) {
	t.Run("LogDriver", func(t *testing.T) {
		// @Note : --log-driver has 1 of followed options [none, json-file, syslog, journald, gelf, fluentd, awslogs, splunk]
		validStr := []string{"docker", "run", "--log-driver", "none", imageName}
		_, host, _ := convertConfig(validStr)

		if strings.Compare(host.LogConfig.Type, "none") != 0 {
			t.Fail()
		}
	})

	t.Run("LogOpt", func(t *testing.T) {
		// @Note : --log-opt must be used with --log-driver (In below cases, use default log-driver(json-file))
		//         log-opt has different key:value followed log-driver
		validStr := []string{"docker", "run", "--log-opt", "max-size=10m", "--log-opt", "max-file=3", imageName}
		_, host, _ := convertConfig(validStr)

		expectLogOpt := make(map[string]string, 2)
		expectLogOpt["max-size"] = "10m"
		expectLogOpt["max-file"] = "3"

		if reflect.DeepEqual(host.LogConfig.Config, expectLogOpt) != true {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigWithMacAddr(t *testing.T) {
	validStr := []string{"docker", "run", "--mac-address", "12:34:56:78:9a:bc", imageName}
	container, _, _ := convertConfig(validStr)

	if strings.Compare(container.MacAddress, "12:34:56:78:9a:bc") != 0 {
		t.Fail()
	}
}

func TestSuccessConvertConfigMemoryOption(t *testing.T) {
	t.Run("KerenlMemory", func(t *testing.T) {
		validStr := []string{"docker", "run", "--kernel-memory", "2g", imageName}
		_, host, _ := convertConfig(validStr)

		// @Note : 2Gb = 2147483648 bytes
		if host.Resources.KernelMemory != int64(2147483648) {
			t.Fail()
		}
	})
	t.Run("Memory", func(t *testing.T) {
		validStr := []string{"docker", "run", "--memory", "2g", imageName}
		_, host, _ := convertConfig(validStr)

		// @Note : 2Gb = 2147483648 bytes
		if host.Memory != int64(2147483648) {
			t.Fail()
		}
	})
	t.Run("MemoryReservation", func(t *testing.T) {
		validStr := []string{"docker", "run", "--memory-reservation", "2b", imageName}
		_, host, _ := convertConfig(validStr)

		if host.MemoryReservation != int64(2) {
			t.Fail()
		}
	})
	t.Run("MemorySwap", func(t *testing.T) {
		validStr := []string{"docker", "run", "--memory-swap", "2k", imageName}
		_, host, _ := convertConfig(validStr)

		// @Note : 2Kb = 2048 bytes
		if host.MemorySwap != int64(2048) {
			t.Fail()
		}

	})
	t.Run("memroySwappiness", func(t *testing.T) {
		validStr := []string{"docker", "run", "--memory-swappiness", "50", imageName}
		_, host, _ := convertConfig(validStr)

		if *host.MemorySwappiness != 50 {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigWithMount(t *testing.T) {
	validStr := []string{"docker", "run", "--mount", "type=volume,target=/icanwrite", imageName}
	_, host, _ := convertConfig(validStr)
	expectMountedInfo := []mount.Mount{{Type: mount.TypeVolume, Target: "/icanwrite"}}

	if reflect.DeepEqual(host.Mounts, expectMountedInfo) != true {
		t.Fail()
	}

	validStr = []string{"docker", "run", "--mount", "type=bind,src=/data,dst=/data", imageName}
	_, host, _ = convertConfig(validStr)
	expectMountedInfo = []mount.Mount{{Type: mount.TypeBind, Source: "/data", Target: "/data"}}

	if reflect.DeepEqual(host.Mounts, expectMountedInfo) != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigNetworkOption(t *testing.T) {
	t.Run("Network", func(t *testing.T) {
		// @Note : allow for both "--net" and "--network", although the latter is the recommended way.
		// @Note : --network
		// Connect a container to a network
		//   'bridge': create a network stack on the default Docker bridge
		//   'none': no networking
		//   'container:<name|id>': reuse another container's network stack
		//   'host': use the Docker host network stack
		//   '<network-name>|<network-id>': connect to a user-defined network
		validStr := []string{"docker", "run", "--network", "host", imageName}
		_, host, _ := convertConfig(validStr)

		if host.NetworkMode.IsHost() != true {
			t.Fail()
		}
	})

	t.Run("NetworkAlias", func(t *testing.T) {
		validStr := []string{"docker", "run", "--network", networkName, "--network-alias", "alias", imageName}

		_, _, network := convertConfig(validStr)

		networkConf := network.EndpointsConfig[networkName]

		if reflect.DeepEqual(networkConf.Aliases, []string{"alias"}) != true {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigOOMOption(t *testing.T) {
	t.Run("OOMKilldisable", func(t *testing.T) {
		// validStr := []string{"docker", "run", "--oom-kill-disable", "true", imageName}
		validStr := []string{"docker", "run", "--oom-kill-disable", imageName}
		_, host, _ := convertConfig(validStr)

		if *host.OomKillDisable != true {
			t.Fail()
		}
	})
	t.Run("OOMScoreAdj", func(t *testing.T) {
		validStr := []string{"docker", "run", "--oom-score-adj", "10", imageName}
		_, host, _ := convertConfig(validStr)

		if host.OomScoreAdj != 10 {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigPIDOption(t *testing.T) {
	t.Run("PID", func(t *testing.T) {
		validStr := []string{"docker", "run", "--pid", "host", imageName}
		_, host, _ := convertConfig(validStr)

		if host.PidMode.IsHost() != true {
			t.Fail()
		}

		validStr = []string{"docker", "run", "--pid", "container:my-redis", imageName}
		_, host, _ = convertConfig(validStr)

		if host.PidMode.IsContainer() != true ||
			strings.Compare(host.PidMode.Container(), "my-redis") != 0 {
			t.Fail()
		}
	})
	t.Run("PIDsLimit", func(t *testing.T) {
		// @Note : --pids-limit=-1 means setting unlimited value
		validStr := []string{"docker", "run", "--pids-limit", "-1", imageName}
		_, host, _ := convertConfig(validStr)

		if *host.PidsLimit != -1 {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigWithPrivileged(t *testing.T) {
	// validStr := []string{"docker", "run", "--privileged", "true", imageName}
	validStr := []string{"docker", "run", "--privileged", imageName}
	_, host, _ := convertConfig(validStr)

	if host.Privileged != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigPublishOption(t *testing.T) {
	t.Run("PublishAll", func(t *testing.T) {
		// validStr := []string{"docker", "run", "--publish-all", "true", imageName}
		validStr := []string{"docker", "run", "--publish-all", imageName}
		_, host, _ := convertConfig(validStr)

		if host.PublishAllPorts != true {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigWithReadOnly(t *testing.T) {
	// validStr := []string{"docker", "run", "--read-only", "true", imageName}
	validStr := []string{"docker", "run", "--read-only", imageName}
	_, host, _ := convertConfig(validStr)

	if host.ReadonlyRootfs != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithRestart(t *testing.T) {
	// @Note : --isolation has 1 of 4 options [no, on-failure[:max-retries], unless-stopped, always]
	validStr := []string{"docker", "run", "--restart", "always", imageName}
	_, host, _ := convertConfig(validStr)

	if host.RestartPolicy.IsAlways() != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithRM(t *testing.T) {
	validStr := []string{"docker", "run", "--rm", imageName}
	_, host, _ := convertConfig(validStr)

	if host.AutoRemove != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithRuntime(t *testing.T) {
	validStr := []string{"docker", "run", "--runtime", "runtime", imageName}
	_, host, _ := convertConfig(validStr)

	if strings.Compare(host.Runtime, "runtime") != 0 {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithSecurityOpt(t *testing.T) {
	validStr := []string{
		"docker", "run", "--security-opt", "label=user:USER", "--security-opt", "no-new-privileges",
		"--security-opt", "label:level:TopSecret", imageName}
	_, host, _ := convertConfig(validStr)

	expectOption := []string{"label=user:USER", "no-new-privileges", "label:level:TopSecret"}
	if reflect.DeepEqual(host.SecurityOpt, expectOption) != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithShmSize(t *testing.T) {
	validStr := []string{"docker", "run", "--shm-size", "2g", imageName}
	_, host, _ := convertConfig(validStr)

	// @Note : 2Gb = 2147483648 bytes
	if host.ShmSize != int64(2147483648) {
		t.Fail()
	}
}

func TestSuccessConvertConfigStopOption(t *testing.T) {
	t.Run("StopSignal", func(t *testing.T) {
		validStr := []string{"docker", "run", "--stop-signal", "SIGINT", imageName}
		container, _, _ := convertConfig(validStr)

		if strings.Compare(container.StopSignal, "SIGINT") != 0 {
			t.Fail()
		}
	})
	t.Run("StopTimeout", func(t *testing.T) {
		validStr := []string{"docker", "run", "--stop-timeout", "10", imageName}
		container, _, _ := convertConfig(validStr)

		if *container.StopTimeout != 10 {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigWithStorageOpt(t *testing.T) {
	validStr := []string{"docker", "run", "--storage-opt", "size=120G", imageName}
	_, host, _ := convertConfig(validStr)

	storageOpt := make(map[string]string, 1)
	storageOpt["size"] = "120G"

	if reflect.DeepEqual(host.StorageOpt, storageOpt) != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithSysctl(t *testing.T) {
	validStr := []string{"docker", "run", "--sysctl", "net.ipv4.ip_forward=1", imageName}
	_, host, _ := convertConfig(validStr)

	sysctls := make(map[string]string, 1)
	sysctls["net.ipv4.ip_forward"] = "1"

	if reflect.DeepEqual(host.Sysctls, sysctls) != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithTmpfs(t *testing.T) {
	validStr := []string{"docker", "run", "--tmpfs", "/run:rw,noexec,nosuid,size=65536k", imageName}
	_, host, _ := convertConfig(validStr)

	tmpfs := make(map[string]string, 1)
	tmpfs["/run"] = "rw,noexec,nosuid,size=65536k"

	if reflect.DeepEqual(host.Tmpfs, tmpfs) != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithTTY(t *testing.T) {
	// validStr := []string{"docker", "run", "--tty", "true", imageName}
	validStr := []string{"docker", "run", "--tty", imageName}
	container, _, _ := convertConfig(validStr)

	if container.Tty != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigWithUlimit(t *testing.T) {
	// @Note : --ulimit <type>=<soft limit>[:<hard limit>]

	// <type> can be 1 of the following

	// core - limits the core file size (KB)
	// data - max data size (KB)
	// fsize - maximum filesize (KB)
	// memlock - max locked-in-memory address space (KB)
	// nofile - max number of open files
	// rss - max resident set size (KB)
	// stack - max stack size (KB)
	// cpu - max CPU time (MIN)
	// nproc - max number of processes
	// as - address space limit (KB)
	// maxlogins - max number of logins for this user
	// maxsyslogins - max number of logins on the system
	// priority - the priority to run user process with
	// locks - max number of file locks the user can hold
	// sigpending - max number of pending signals
	// msgqueue - max memory used by POSIX message queues (bytes)
	// nice - max nice priority allowed to raise to values: [-20, 19]
	// rtprio - max realtime priority

	validStr := []string{"docker", "run", "--ulimit", "nofile=1024:1024", imageName}
	_, host, _ := convertConfig(validStr)

	expectUlimit := []*units.Ulimit{{Name: "nofile", Soft: 1024, Hard: 1024}}

	if reflect.DeepEqual(host.Ulimits, expectUlimit) != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigUserOption(t *testing.T) {
	t.Run("User", func(t *testing.T) {
		validStr := []string{"docker", "run", "--user", "User", imageName}
		container, _, _ := convertConfig(validStr)

		if strings.Compare(container.User, "User") != 0 {
			t.Fail()
		}
	})
	t.Run("Usern", func(t *testing.T) {
		validStr := []string{"docker", "run", "--userns", "host", imageName}
		_, host, _ := convertConfig(validStr)

		if host.UsernsMode.IsHost() != true {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigWithUTS(t *testing.T) {
	validStr := []string{"docker", "run", "--uts", "host", imageName}
	_, host, _ := convertConfig(validStr)

	if host.UTSMode.IsHost() != true {
		t.Fail()
	}
}

func TestSuccessConvertConfigVolumeOption(t *testing.T) {
	t.Run("Volume", func(t *testing.T) {
		validStr := []string{"docker", "run", "--volume", "/var/run:/var/run:ro", imageName}
		_, host, _ := convertConfig(validStr)

		volume := []string{"/var/run:/var/run:ro"}
		if reflect.DeepEqual(host.Binds, volume) != true {
			t.Fail()
		}
	})
	t.Run("VolumeDriver", func(t *testing.T) {
		validStr := []string{"docker", "run", "--volume-driver", "vieux/sshfs", imageName}
		_, host, _ := convertConfig(validStr)

		if strings.Compare(host.VolumeDriver, "vieux/sshfs") != 0 {
			t.Fail()
		}
	})
	t.Run("VolumesFrom", func(t *testing.T) {
		validStr := []string{"docker", "run", "--volumes-from", "hello:rw", imageName}
		_, host, _ := convertConfig(validStr)

		volumesFrom := []string{"hello:rw"}
		if reflect.DeepEqual(host.VolumesFrom, volumesFrom) != true {
			t.Fail()
		}
	})
}

func TestSuccessConvertConfigWithWorkDir(t *testing.T) {
	validStr := []string{"docker", "run", "--workdir", "/var/www", imageName}
	container, _, _ := convertConfig(validStr)

	if strings.Compare(container.WorkingDir, "/var/www") != 0 {
		t.Fail()
	}
}
