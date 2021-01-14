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
package configuration

import (
	"encoding/json"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/errors"
	bolt "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/wrapper"
)

const bucketName = "configuration"

type Configuration struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	ExecType string `json:"executionType"`
}

type DBInterface interface {
	Get(id string) (Configuration, error)
	GetList() ([]Configuration, error)
	Set(conf Configuration) error
	Update(conf Configuration) error
	Delete(id string) error
}

type Query struct {
}

var db bolt.Database

func init() {
	db = bolt.NewBoltDB(bucketName)
}

func (Query) Get(id string) (Configuration, error) {
	var conf Configuration

	value, err := db.Get([]byte(id))
	if err != nil {
		return conf, err
	}

	conf, err = decode(value)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

func (Query) GetList() ([]Configuration, error) {
	infos, err := db.List()
	if err != nil {
		return nil, err
	}

	list := make([]Configuration, 0)
	for _, data := range infos {
		info, err := decode([]byte(data.(string)))
		if err != nil {
			continue
		}
		list = append(list, info)
	}
	return list, nil
}

func (Query) Set(conf Configuration) error {
	encoded, err := conf.encode()
	if err != nil {
		return err
	}

	err = db.Put([]byte(conf.ID), encoded)
	if err != nil {
		return err
	}
	return nil
}

func (Query) Update(conf Configuration) error {
	data, err := db.Get([]byte(conf.ID))
	if err != nil {
		return errors.DBOperationError{Message: err.Error()}
	}

	stored, err := decode(data)
	if err != nil {
		return err
	}

	stored.Platform = conf.Platform
	stored.ExecType = conf.ExecType

	encoded, err := stored.encode()
	if err != nil {
		return err
	}

	return db.Put([]byte(conf.ID), encoded)
}

func (Query) Delete(id string) error {
	return db.Delete([]byte(id))
}

func (conf Configuration) convertToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":            conf.ID,
		"platform":      conf.Platform,
		"executionType": conf.ExecType,
	}
}

func (conf Configuration) encode() ([]byte, error) {
	encoded, err := json.Marshal(conf)
	if err != nil {
		return nil, errors.InvalidJSON{Message: err.Error()}
	}
	return encoded, nil
}

func decode(data []byte) (Configuration, error) {
	var conf Configuration
	err := json.Unmarshal(data, &conf)
	if err != nil {
		return conf, errors.InvalidJSON{Message: err.Error()}
	}
	return conf, nil
}
