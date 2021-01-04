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

// Package native provides native specific functions for configuremgr
package native

import (
	"fmt"
	"io/ioutil"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
	"os"
	"path/filepath"
	"strings"
	"time"

	types "github.com/lf-edge/edge-home-orchestration-go/src/common/types/configuremgrtypes"
	"github.com/lf-edge/edge-home-orchestration-go/src/controller/configuremgr"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/ini.v1"
)

const logPrefix = "nativeconfiguremgr"

// ConfigureMgr has config folder path
type ConfigureMgr struct {
	confpath string
}

var (
	configuremgrObj *ConfigureMgr
	log             = logmgr.GetInstance()
)

func init() {
	configuremgrObj = new(ConfigureMgr)
}

// GetInstance set configpath and gives ConfigureMgrs Singletone instance
func GetInstance(configPath string) *ConfigureMgr {
	configuremgrObj.confpath = configPath
	return configuremgrObj
}

// SetConfigPath update config folder path
func (cfgMgr ConfigureMgr) SetConfigPath(configPath string) error {
	_, err := os.Stat(configPath)
	if err == nil {
		configuremgrObj.confpath = configPath
	} else {
		log.Println("no config file path")
	}
	return err
}

// Watch implements Watcher interface with ConfigureMgr struct
func (cfgMgr ConfigureMgr) Watch(notifier configuremgr.Notifier) {
	// logic for already installed configuration
	files, err := ioutil.ReadDir(cfgMgr.confpath)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		notifier.Notify(getServiceInfo(cfgMgr.confpath + "/" + f.Name()))
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				switch event.Op {
				case fsnotify.Create, fsnotify.Write:
					_, dirName := filepath.Split(event.Name)
					confFileName := fmt.Sprint(event.Name, "/", dirName, ".conf")
					log.Println("IsConfExist:", confFileName)

					// Should check file is exist on file system really,
					// even though CREATE event of directory received
					isConfExist := false
					for i := 0; i < 5; i++ {
						if _, err := os.Stat(confFileName); !os.IsNotExist(err) {
							isConfExist = true
							break
						}
						time.Sleep(time.Second * 1)
					}
					if isConfExist != true {
						log.Println(confFileName, "does not exist")
						continue
					}
					notifier.Notify(getServiceInfo(event.Name))
				case fsnotify.Remove:
					// TODO remove scoring
				}
			case err := <-watcher.Errors:
				if err != nil {
					log.Println("error:", err)
				}
			} //select end
		} //for end
	}()

	err = watcher.Add(cfgMgr.confpath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("start watching for", cfgMgr.confpath)
	log.Println("configuremgr watcher register end")
}

func getServiceInfo(path string) types.ServiceInfo {
	confPath, err := getdirname(path)
	if err != nil {
		log.Println("wrong libPath or confPath")
	}

	cfg, err := ini.Load(confPath)
	if err != nil {
		log.Printf("Fail to read %v file, err = %v", confPath, err)
	}

	serviceName := cfg.Section("ServiceInfo").Key("ServiceName").String()
	executableName := cfg.Section("ServiceInfo").Key("ExecutableFileName").String()
	allowedRequesterName := cfg.Section("ServiceInfo").Key("AllowedRequester").Strings(",")

	log.Println("[configuremgr] ServiceName:", serviceName)
	log.Println("[configuremgr] ExecutableFileName:", executableName)
	log.Println("[configuremgr] AllowedRequester:", allowedRequesterName)

	ret := types.ServiceInfo{
		ServiceName:        serviceName,
		ExecutableFileName: executableName,
		AllowedRequester:   allowedRequesterName,
	}

	return ret
}

func getdirname(path string) (confPath string, err error) {

	idx := strings.LastIndex(path, "/")
	if idx == (len(path) - 1) {
		path = path[:len(path)-1]
	}

	dirname := path[strings.LastIndex(path, "/")+1:]

	confPath = path + "/" + dirname + ".conf"

	//NOTE : copy but really copy, it can be not existed.
	for {
		if _, err := os.Stat(confPath); err == nil {
			break
		}
		time.Sleep(time.Second * 1)
	}

	log.Println("[configuremgr] confPath : " + confPath)

	return
}
