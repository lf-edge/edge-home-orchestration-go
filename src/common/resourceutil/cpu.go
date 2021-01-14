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

package resourceutil

import (
	"time"

	commoncpu "github.com/lf-edge/edge-home-orchestration-go/src/common/resourceutil/cpu"
	resourceDB "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/resource"
)

type cpuUtil struct {
	percent func(interval time.Duration, percpu bool) ([]float64, error)
	info    func() ([]commoncpu.InfoStat, error)
}

var (
	cpu = cpuUtil{}
)

func init() {
	cpu.info = commoncpu.Info
	cpu.percent = commoncpu.Percent
}

func processCPUInfo() {
	go func() {
		for {
			checkCPUUsage()
			checkCPUFreq()
			checkCPUCount()

			time.Sleep(time.Duration(defaultProcessingTime) * time.Second)
		}
	}()
}

func checkCPUUsage() {
	var usage float64

	cpus, err := cpu.percent(time.Second, true)
	if err != nil {
		log.Println(logPrefix, "usage of cpu is fail : ", err.Error())
		return
	}

	for _, cpu := range cpus {
		usage += float64(cpu)
	}
	usage /= float64(len(cpus))

	info := resourceDB.ResourceInfo{}
	info.Name = CPUUsage
	info.Value = usage

	err = resourceDBExecutor.Set(info)
	if err != nil {
		log.Println(logPrefix, "DB error : ", err.Error())
	}
	return
}

func checkCPUFreq() (out float64, err error) {
	infos, err := cpu.info()
	if err != nil {
		log.Println(logPrefix, "cpu.Info() fail : ", err.Error())
		return
	}

	info := resourceDB.ResourceInfo{}
	info.Name = CPUFreq
	info.Value = infos[0].Mhz

	err = resourceDBExecutor.Set(info)
	if err != nil {
		log.Println(logPrefix, "DB error : ", err.Error())
	}
	return
}

func checkCPUCount() {
	infos, err := cpu.info()
	if err != nil {
		log.Println(logPrefix, "cpu info getting fail : ", err.Error())
		return
	}

	info := resourceDB.ResourceInfo{}
	info.Name = CPUCount
	info.Value = float64(len(infos))

	err = resourceDBExecutor.Set(info)
	if err != nil {
		log.Println(logPrefix, "DB error : ", err.Error())
	}
	return
}
