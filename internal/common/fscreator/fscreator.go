/*******************************************************************************
* Copyright 2021 Samsung Electronics All Rights Reserved.
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

package fscreator

import (
	"os"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
)

const (
	logPrefix = "[fscreator]"
)

var edgeDirs = []string{
	"/apps",
	"/certs",
	"/data/db",
	"/datastorage",
	"/device",
	"/log",
	"/mnedc",
	"/user",
}

var (
	log = logmgr.GetInstance()
)

// CreateFileSystem creates a file system necessary for the work of the edge-orchestration
// ├── var
//
//	└── edge-orchestration
//	    ├── apps
//	    ├── certs
//	    ├── data
//	    │   └── db
//	    ├── datastorage
//	    ├── device
//	    ├── log
//	    ├── mnedc
//	    └── user
func CreateFileSystem(edgeDir string) error {
	for _, dir := range edgeDirs {
		if err := os.MkdirAll(edgeDir+dir, os.ModePerm); err != nil {
			log.Panicf("%s Failed to create %s: %s\n", logPrefix, edgeDir+dir, err)
			return err
		}
	}
	return nil
}
