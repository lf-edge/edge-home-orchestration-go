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

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/errors"
	"github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/common"
	bolt "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/wrapper"
)

const bucketName = "service"

// Info struct
type Info struct {
	ID       string   `json:"id"`
	Services []string `json:"services"`
}

// DBInterface interface
type DBInterface interface {
	Get(id string) (Info, error)
	GetList() ([]Info, error)
	Set(info Info) error
	Update(info Info) error
	Delete(name string) error
}

// Query struct
type Query struct {
}

var db bolt.Database

func init() {
	db = bolt.NewBoltDB(bucketName)
}

// Get returns service info that matches id
func (Query) Get(id string) (Info, error) {
	var info Info

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

// GetList returns the list of service info
func (Query) GetList() ([]Info, error) {
	infos, err := db.List()
	if err != nil {
		return nil, err
	}

	list := make([]Info, 0)
	for _, data := range infos {
		info, err := decode([]byte(data.(string)))
		if err != nil {
			continue
		}
		list = append(list, info)
	}
	return list, nil
}

// Set sets the service info
func (Query) Set(info Info) error {
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

// Update updates the service info that matches id
func (Query) Update(info Info) error {
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

// Delete deletes the service info that matches id
func (Query) Delete(id string) error {
	return db.Delete([]byte(id))
}

func (info Info) convertToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":       info.ID,
		"services": info.Services,
	}
}

func (info Info) encode() ([]byte, error) {
	encoded, err := json.Marshal(info)
	if err != nil {
		return nil, errors.InvalidJSON{Message: err.Error()}
	}
	return encoded, nil
}

func decode(data []byte) (Info, error) {
	var info Info
	err := json.Unmarshal(data, &info)
	if err != nil {
		return info, errors.InvalidJSON{Message: err.Error()}
	}
	return info, nil
}
