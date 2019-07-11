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
	db "db/bolt/network"
	"fmt"
	"restinterface/resthelper"

	"strconv"
	"strings"
	"time"
)

const (
	pingAPI            = "/api/v1/ping"
	defaultRttDuration = 5
)

var (
	helper     resthelper.RestHelper
	dbExecutor db.DBInterface
)

func init() {
	helper = resthelper.GetHelper()
	dbExecutor = db.Query{}

	processRTT()
}

func processRTT() {
	go func() {
		for {
			netInfos, err := dbExecutor.GetList()
			if err != nil {
				return
			}

			for _, netInfo := range netInfos {
				totalCount := len(netInfo.IPv4.Wired) + len(netInfo.IPv4.Wireless)
				ch := make(chan float64, totalCount)
				for _, ip := range netInfo.IPv4.Wired {
					go func(targetIP string) {
						ch <- checkRTT(targetIP)
					}(ip)
				}
				for _, ip := range netInfo.IPv4.Wireless {
					go func(targetIP string) {
						ch <- checkRTT(targetIP)
					}(ip)
				}
				go func(info db.NetworkInfo) {
					result := selectMinRTT(ch, totalCount)
					fmt.Println("devid : ", info.Id, ", result : ", result)
					info.RTT = result
					dbExecutor.Update(info)
				}(netInfo)
			}
			time.Sleep(time.Duration(defaultRttDuration) * time.Second)
		}
	}()
}

func checkRTT(ip string) (rtt float64) {
	s := strings.Split(ip, ":")
	if len(s) != 2 {
		return
	}

	address := s[0]
	port, err := strconv.Atoi(s[1])
	if err != nil {
		return
	}

	targetURL := helper.MakeTargetURL(address, port, pingAPI)

	reqTime := time.Now()
	_, _, err = helper.DoGet(targetURL)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	return time.Now().Sub(reqTime).Seconds()
}

func selectMinRTT(ch chan float64, totalCount int) (minRTT float64) {
	for i := 0; i < totalCount; i++ {
		select {
		case rtt := <-ch:
			if (rtt != 0 && rtt < minRTT) || minRTT == 0 {
				minRTT = rtt
			}
		}
	}
	return
}
