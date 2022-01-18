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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
	types "github.com/lf-edge/edge-home-orchestration-go/internal/common/types/configuremgrtypes"
	appDB "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/application"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/ini.v1"
)

const logPrefix = "[configuremgr]"

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
	appQuery        appDB.DBInterface
	configuremgrObj *ConfigureMgr
	log             = logmgr.GetInstance()
)

func init() {
	appQuery = appDB.Query{}
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
		log.Warn(logPrefix, " No config file path")
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
				log.Debug(logPrefix, " Event:", event)
				switch event.Op {
				case fsnotify.Create, fsnotify.Write:
					_, dirName := filepath.Split(event.Name)
					confFileName := fmt.Sprint(event.Name, "/", dirName, ".conf")
					log.Debug(logPrefix, " IsConfExist:", confFileName)

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
						log.Warn(logPrefix, " ", confFileName, "does not exist")
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
					log.Warn(logPrefix, " error:", err)
				}
			} //select end
		} //for end
	}()

	err = watcher.Add(cfgMgr.confpath)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(logPrefix, " Start watching for ", cfgMgr.confpath)
	log.Debug(logPrefix, " Configuremgr watcher register end")
}

func getServiceInfo(path string) (types.ServiceInfo, error) {
	confPath, err := getdirname(path)
	if err != nil {
		log.Warn(logPrefix, " Wrong libPath or confPath")
		return types.ServiceInfo{}, err
	}

	cfg, err := ini.Load(confPath)
	if err != nil {
		log.Debug(logPrefix, " Fail to read ", confPath, "file, err = ", err)
		return types.ServiceInfo{}, err
	}

	serviceName := cfg.Section("ServiceInfo").Key("ServiceName").String()
	executableName := cfg.Section("ServiceInfo").Key("ExecutableFileName").String()
	allowedRequesterName := cfg.Section("ServiceInfo").Key("AllowedRequester").Strings(",")
	execType := cfg.Section("ServiceInfo").Key("ExecType").String()
	execCmd := cfg.Section("ServiceInfo").Key("ExecCmd").Strings(" ")

	log.Debug(logPrefix, " ServiceName:", serviceName)
	log.Debug(logPrefix, " ExecutableFileName:", executableName)
	log.Debug(logPrefix, " AllowedRequester:", allowedRequesterName)
	log.Debug(logPrefix, " ExecType:", execType)
	log.Debug(logPrefix, " ExecCmd:", execCmd)

	if execType != configuremgrObj.execType {
		log.Warn(logPrefix, " Type of ", serviceName, " is not ", configuremgrObj.execType)
		return types.ServiceInfo{}, errors.New("execution type mismatch")
	}

	ret := types.ServiceInfo{
		ServiceName:        serviceName,
		ExecutableFileName: executableName,
		AllowedRequester:   allowedRequesterName,
		ExecType:           execType,
		ExecCmd:            execCmd,
	}

	appInfo := appDB.Info{
		ServiceName:        serviceName,
		ExecutableFileName: executableName,
		AllowedRequester:   allowedRequesterName,
		ExecType:           execType,
		ExecCmd:            execCmd,
	}

	setAppDB(appInfo)

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

func setAppDB(appInfo appDB.Info) {
	err := appQuery.Set(appInfo)
	if err != nil {
		log.Warn(err.Error())
	}
}

// GetAppDB returns information corresponding to the service application name
func GetAppDB(name string) (appDB.Info, error) {
	info, err := appQuery.Get(name)
	if err != nil {
		log.Warn(err.Error())
	}

	return info, err
}
