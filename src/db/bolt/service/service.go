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

package service

import (
	"encoding/json"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/errors"
	"github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/common"
	bolt "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/wrapper"
)

const bucketName = "service"

type ServiceInfo struct {
	ID       string   `json:"id"`
	Services []string `json:"services"`
}

type DBInterface interface {
	Get(id string) (ServiceInfo, error)
	GetList() ([]ServiceInfo, error)
	Set(info ServiceInfo) error
	Update(info ServiceInfo) error
	Delete(name string) error
}

type Query struct {
}

var db bolt.Database

func init() {
	db = bolt.NewBoltDB(bucketName)
}

func (Query) Get(id string) (ServiceInfo, error) {
	var info ServiceInfo

	value, err := db.Get([]byte(id))
	if err != nil {
		return info, err
	}

	info, err = decode(value)
	if err != nil {
		return info, err
	}

	return info, nil
}

func (Query) GetList() ([]ServiceInfo, error) {
	infos, err := db.List()
	if err != nil {
		return nil, err
	}

	list := make([]ServiceInfo, 0)
	for _, data := range infos {
		info, err := decode([]byte(data.(string)))
		if err != nil {
			continue
		}
		list = append(list, info)
	}
	return list, nil
}

func (Query) Set(info ServiceInfo) error {
	encoded, err := info.encode()
	if err != nil {
		return err
	}

	err = db.Put([]byte(info.ID), encoded)
	if err != nil {
		return err
	}
	return nil
}

func (Query) Update(info ServiceInfo) error {
	data, err := db.Get([]byte(info.ID))
	if err != nil {
		return errors.DBOperationError{Message: err.Error()}
	}

	stored, err := decode(data)
	if err != nil {
		return err
	}

	for _, service := range info.Services {
		if !common.HasElem(stored.Services, service) {
			stored.Services = append(stored.Services, service)
		}
	}

	encoded, err := stored.encode()
	if err != nil {
		return err
	}

	return db.Put([]byte(info.ID), encoded)
}

func (Query) Delete(id string) error {
	return db.Delete([]byte(id))
}

func (info ServiceInfo) convertToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":       info.ID,
		"services": info.Services,
	}
}

func (info ServiceInfo) encode() ([]byte, error) {
	encoded, err := json.Marshal(info)
	if err != nil {
		return nil, errors.InvalidJSON{Message: err.Error()}
	}
	return encoded, nil
}

func decode(data []byte) (ServiceInfo, error) {
	var info ServiceInfo
	err := json.Unmarshal(data, &info)
	if err != nil {
		return info, errors.InvalidJSON{Message: err.Error()}
	}
	return info, nil
}
