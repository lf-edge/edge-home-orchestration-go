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

package application

import (
	"encoding/json"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/errors"
	bolt "github.com/lf-edge/edge-home-orchestration-go/internal/db/bolt/wrapper"
)

const bucketName = "application"

// Info struct
type Info struct {
	ServiceName        string   `json:"serviceName"`
	ExecutableFileName string   `json:"executableFileName"`
	AllowedRequester   []string `json:"allowedRequester"`
	ExecType           string   `json:"execType"`
	ExecCmd            []string `json:"execCmd"`
}

// DBInterface interface
type DBInterface interface {
	Get(name string) (Info, error)
	GetList() ([]Info, error)
	Set(info Info) error
	Delete(name string) error
}

// Query struct
type Query struct {
}

var db bolt.Database

func init() {
	db = bolt.NewBoltDB(bucketName)
}

// Get returns application info that matches service name
func (Query) Get(name string) (Info, error) {
	var info Info

	value, err := db.Get([]byte(name))
	if err != nil {
		return info, err
	}

	info, err = decode(value)
	if err != nil {
		return info, err
	}

	return info, nil
}

// GetList returns the list of application info
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

// Set sets the application info
func (Query) Set(info Info) error {
	encoded, err := info.encode()
	if err != nil {
		return err
	}

	err = db.Put([]byte(info.ServiceName), encoded)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the application info that matches service name
func (Query) Delete(id string) error {
	return db.Delete([]byte(id))
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
