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

package senderresolver

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
)

const (
	PROC_TCP      = "/proc/net/tcp"
	WELLKNOWNPORT = 56001
)

var (
	PROCESS_INFO_PATH = "/process"
	log               = logmgr.GetInstance()
)

func GetNameByPort(port int64) (string, error) {
	lines, err := getData()
	if err != nil {
		return "", err
	}
	for _, str := range lines {
		line_array := removeEmpty(strings.Split(strings.TrimSpace(str), " "))

		src, err := strconv.ParseInt(strings.Split(line_array[1], ":")[1], 16, 32)
		if err != nil {
			return "", err
		}

		dst, err := strconv.ParseInt(strings.Split(line_array[2], ":")[1], 16, 32)
		if err != nil {
			return "", err
		}

		if dst != WELLKNOWNPORT || src != port {
			continue
		}

		pid, err := getPid(line_array[9])
		if err != nil {
			return "", err
		}

		process, err := getProcess(pid)
		if err != nil {
			return "", err
		}

		log.Println("returning: ", process)
		return process, nil

	}

	return "", errors.New("not found port")
}

func getData() ([]string, error) {
	data, err := ioutil.ReadFile(PROC_TCP)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")

	return lines[1 : len(lines)-1], nil
}

func removeEmpty(array []string) []string {
	new_array := make([]string, 0)
	for _, str := range array {
		if str != "" {
			new_array = append(new_array, str)
		}
	}
	return new_array
}

func getProcess(pid string) (string, error) {
	fp := PROCESS_INFO_PATH + "/" + pid + "/comm"

	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return "", err
	}

	if data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	return string(data), nil
}

func getPid(inode string) (string, error) {
	pid := "-"

	d, err := filepath.Glob(PROCESS_INFO_PATH + "/[0-9]*/fd/[0-9]*")
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(inode)
	for _, item := range d {
		path, _ := os.Readlink(item)
		out := re.FindString(path)
		if len(out) != 0 {
			pid = strings.Split(item, "/")[2]
		}
	}
	return pid, nil
}
