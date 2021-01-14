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

// Package resourceutil provides the information of resource usage of local device
package resourceutil

import (
	"time"

	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/resthelper"

	netDB "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/network"
)

const (
	pingAPI            = "/api/v1/ping"
	internalPort       = 56002
	defaultRttDuration = 5
	tryLimit           = 12
)

var (
	helper        resthelper.RestHelper
	netDBExecutor netDB.DBInterface
)

func init() {
	helper = resthelper.GetHelper()
	netDBExecutor = netDB.Query{}
}

func processRTT() {
	go func() {
		for {
			netInfos, err := netDBExecutor.GetList()
			if err != nil {
				return
			}

			for _, netInfo := range netInfos {
				totalCount := len(netInfo.IPv4)
				ch := make(chan float64, totalCount)
				for _, ip := range netInfo.IPv4 {
					go func(targetIP string) {
						ch <- checkRTT(targetIP)
					}(ip)
				}
				go func(info netDB.NetworkInfo) {
					result := selectMinRTT(ch, totalCount)
					if info.RTT < 0 && result < 0 {
						if info.RTT == -1 * tryLimit {
							_, err := netDBExecutor.Get(info.ID)
							if err == nil {
								log.Println(logPrefix, "Delete", info.ID, "from netDB")
								netDBExecutor.Delete(info.ID)
							}
						} else {
							info.RTT += result
							netDBExecutor.Update(info)
						}
					} else {
						info.RTT = result
						netDBExecutor.Update(info)
					}
				}(netInfo)
			}
			time.Sleep(time.Duration(defaultRttDuration) * time.Second)
		}
	}()
}

func checkRTT(ip string) (rtt float64) {
	targetURL := helper.MakeTargetURL(ip, internalPort, pingAPI)

	reqTime := time.Now()
	_, _, err := helper.DoGet(targetURL)
	if err != nil {
		log.Println(err.Error())
		return -1
	}

	return time.Now().Sub(reqTime).Seconds()
}

func selectMinRTT(ch chan float64, totalCount int) (minRTT float64) {
	for i := 0; i < totalCount; i++ {
		select {
		case rtt := <-ch:
			if (minRTT < 0 && rtt > 0) || (rtt > 0 && rtt < minRTT) || minRTT == 0 {
				minRTT = rtt
			}
		}
	}
	return
}
