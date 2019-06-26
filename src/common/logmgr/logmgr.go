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

// Package logmgr implements to handle log files and log format
package logmgr

import (
	"log"

	"github.com/leemcloughlin/logfile"
)

var logFileName = "logmgr.log"

// Init sets the enviroments for logging
func Init(logFilePath string) {
	logFile, err := logfile.New(
		&logfile.LogFile{
			FileName:    logFilePath + "/" + logFileName,
			FileMode:    0644,
			MaxSize:     500 * 1024, // 500K
			OldVersions: 3,
			Flags:       logfile.OverWriteOnStart | logfile.RotateOnStart})
	if err != nil {
		log.Panicf("Failed to create logFile %s: %s\n", logFileName, err)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(logFile)
}
