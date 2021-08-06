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

// Package configuremgr provides interfaces between orchestrationapi and configuremgr
package configuremgr

import (
	"errors"
	"fmt"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	types "github.com/lf-edge/edge-home-orchestration-go/internal/common/types/configuremgrtypes"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/ini.v1"
)

const logPrefix = "Configuremgr"

// Notifier is the interface to get scoring information for each service application
type Notifier interface {
	Notify(serviceinfo types.ServiceInfo)
}

// Watcher is the interface to check if service application is installed/updated/deleted
type Watcher interface {
	Watch(notifier Notifier)
}

// ConfigureMgr has config folder path and execution type information
type ConfigureMgr struct {
	confpath string
	execType string
}

var (
	configuremgrObj *ConfigureMgr
	log             = logmgr.GetInstance()
)

func init() {
	configuremgrObj = new(ConfigureMgr)
}

// GetInstance set configpath, execution type and gives ConfigureMgrs Singletone instance
func GetInstance(configPath, execType string) *ConfigureMgr {
	configuremgrObj.confpath = configPath
	configuremgrObj.execType = execType
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
func (cfgMgr ConfigureMgr) Watch(notifier Notifier) {
	// logic for already installed configuration
	files, err := ioutil.ReadDir(cfgMgr.confpath)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		info, err := getServiceInfo(cfgMgr.confpath + "/" + f.Name())
		if err == nil {
			notifier.Notify(info)
		}
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
					if !isConfExist {
						log.Println(confFileName, "does not exist")
						continue
					}
					info, err := getServiceInfo(event.Name)
					if err == nil {
						notifier.Notify(info)
					}
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

func getServiceInfo(path string) (types.ServiceInfo, error) {
	confPath, err := getdirname(path)
	if err != nil {
		log.Warn("wrong libPath or confPath")
		return types.ServiceInfo{}, err
	}

	cfg, err := ini.Load(confPath)
	if err != nil {
		log.Debug("Fail to read ", confPath, "file, err = ", err)
		return types.ServiceInfo{}, err
	}

	serviceName := cfg.Section("ServiceInfo").Key("ServiceName").String()
	executableName := cfg.Section("ServiceInfo").Key("ExecutableFileName").String()
	allowedRequesterName := cfg.Section("ServiceInfo").Key("AllowedRequester").Strings(",")
	execType := cfg.Section("ServiceInfo").Key("ExecType").String()

	log.Println("[configuremgr] ServiceName:", serviceName)
	log.Println("[configuremgr] ExecutableFileName:", executableName)
	log.Println("[configuremgr] AllowedRequester:", allowedRequesterName)
	log.Println("[configuremgr] ExecType:", execType)

	if execType != configuremgrObj.execType {
		log.Warn("Type of ", serviceName, " is not ", configuremgrObj.execType)
		return types.ServiceInfo{}, errors.New("execution type mismatch")
	}

	ret := types.ServiceInfo{
		ServiceName:        serviceName,
		ExecutableFileName: executableName,
		AllowedRequester:   allowedRequesterName,
		ExecType:           execType,
	}

	return ret, nil
}

func getdirname(path string) (confPath string, err error) {

	idx := strings.LastIndex(path, "/")
	if idx == (len(path) - 1) {
		path = path[:len(path)-1]
	}

	dirname := path[strings.LastIndex(path, "/")+1:]

	confPath = path + "/" + dirname + ".conf"

	//NOTE : copy but really copy, it can be not existed.
	for i := 0; i < 5; i++ {
		if _, err = os.Stat(confPath); err == nil {
			return
		}
		time.Sleep(time.Second * 1)
	}
	return "", err
}
