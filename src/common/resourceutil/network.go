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
	"strings"
	"time"

	resourceDB "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/resource"

	netutil "github.com/vishvananda/netlink"
)

type netUtil struct {
	linkList func() ([]netutil.Link, error)
}

var (
	net = netUtil{}
)

func init() {
	net.linkList = netutil.LinkList
}

func processNetInfo() {
	go func() {
		for {
			checkNetworkMBps()
			checkNetworkBandwidth()

			time.Sleep(time.Duration(defaultProcessingTime) * time.Second)
		}
	}()
}

func checkNetworkMBps() {
	linklist, err := net.linkList()
	if err != nil {
		log.Println(logPrefix, "network link getting fail : ", err.Error())
		return
	}

	var prevTotalBytes, nextTotalBytes uint64

	for _, link := range linklist {
		prevTotalBytes += link.Attrs().Statistics.RxBytes
		prevTotalBytes += link.Attrs().Statistics.TxBytes
	}

	time.Sleep(1 * time.Second)

	linklist, err = net.linkList()
	if err != nil {
		log.Println(logPrefix, "network link getting fail : ", err.Error())
		return
	}

	for _, link := range linklist {
		nextTotalBytes += link.Attrs().Statistics.RxBytes
		nextTotalBytes += link.Attrs().Statistics.TxBytes
	}

	info := resourceDB.ResourceInfo{}
	info.Name = NetMBps
	info.Value = float64((nextTotalBytes - prevTotalBytes)) / 1024 / 1024

	err = resourceDBExecutor.Set(info)
	if err != nil {
		log.Println(logPrefix, "DB error : ", err.Error())
	}

	return
}

func checkNetworkBandwidth() {
	linklist, err := net.linkList()
	if err != nil {
		log.Println(logPrefix, "network link getting fail : ", err.Error())
		return
	}

	var count, total int

	for _, link := range linklist {
		if strings.Contains(link.Attrs().Name, "eth") ||
			strings.Contains(link.Attrs().Name, "enp") ||
			strings.Contains(link.Attrs().Name, "wl") {
			count++
			total += link.Attrs().TxQLen
		}
	}

	if count == 0 {
		log.Println("Not matched network interface")
		return
	}

	info := resourceDB.ResourceInfo{}
	info.Name = NetBandwidth
	info.Value = float64(total / count)

	err = resourceDBExecutor.Set(info)
	if err != nil {
		log.Println(logPrefix, "DB error : ", err.Error())
	}

	return
}
