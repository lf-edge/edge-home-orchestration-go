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

	resourceDB "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/resource"

	memutil "github.com/shirou/gopsutil/mem"
)

type memUtil struct {
	virtualMemory func() (*memutil.VirtualMemoryStat, error)
}

var (
	mem = memUtil{}
)

func init() {
	mem.virtualMemory = memutil.VirtualMemory
}

func processMEMInfo() {
	go func() {
		for {
			checkMemoryAvailable()
			checkMemoryFree()

			time.Sleep(time.Duration(defaultProcessingTime) * time.Second)
		}
	}()
}

func checkMemoryAvailable() {
	memStat, err := mem.virtualMemory()
	if err != nil {
		log.Println(logPrefix, "mem info getting fail : ", err.Error())
		return
	}

	info := resourceDB.ResourceInfo{}
	info.Name = MemAvailable
	info.Value = float64(memStat.Available) / 1024

	err = resourceDBExecutor.Set(info)
	if err != nil {
		log.Println(logPrefix, "DB error : ", err.Error())
	}

	return
}

func checkMemoryFree() {
	memStat, err := mem.virtualMemory()
	if err != nil {
		log.Println(logPrefix, "mem info getting fail : ", err.Error())
		return
	}

	info := resourceDB.ResourceInfo{}
	info.Name = MemFree
	info.Value = float64(memStat.Free) / 1024

	err = resourceDBExecutor.Set(info)
	if err != nil {
		log.Println(logPrefix, "DB error : ", err.Error())
	}

	return
}
