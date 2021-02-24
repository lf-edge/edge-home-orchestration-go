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
	"os"
	"fmt"
	"runtime"
	"strings"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/leemcloughlin/logfile"
)

var (
	logFileName = "logmgr.log"
	logIns      *logrus.Logger
)

func init() {
	logIns = logrus.New()
	logIns.SetReportCaller(true)
	logIns.Formatter = &logrus.TextFormatter{
		FullTimestamp:true,
		ForceColors:true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			function := s[len(s)-1]
			_, filename := path.Split(f.File)
			filenline := fmt.Sprintf("%s:%d", filename, f.Line)
			return function, filenline
		},
	}
	logIns.Level = logrus.InfoLevel
	logIns.Out = os.Stdout
}

// Sets the environments for a log file
func InitLogfile(logFilePath string) {
	if _, err := os.Stat(logFilePath); err != nil {
		err := os.MkdirAll(logFilePath, os.ModePerm)
		if err != nil {
			logIns.Panicf("Failed to create logFilePath %s: %s\n", logFilePath, err)
		}
	}

	logFile, err := logfile.New(
		&logfile.LogFile{
			FileName:    logFilePath + "/" + logFileName,
			FileMode:    0644,
			MaxSize:     500 * 1024, // 500K
			OldVersions: 3,
			Flags:       logfile.OverWriteOnStart | logfile.RotateOnStart})
	if err != nil {
		logIns.Panicf("Failed to create logFile %s: %s\n", logFileName, err)
	}

	logIns.Out = logFile
}

func GetInstance() *logrus.Logger {
	return logIns
}
