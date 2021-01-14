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

package system

import (
	"encoding/json"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/errors"
	bolt "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/wrapper"
)

const (
	bucketName = "system"

	ID       = "id"
	Platform = "platform"
	ExecType = "execType"
)

type SystemInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type DBInterface interface {
	Get(name string) (SystemInfo, error)
	Set(info SystemInfo) error
	Delete(name string) error
}

type Query struct {
}

var db bolt.Database

func init() {
	db = bolt.NewBoltDB(bucketName)
}

func (Query) Get(name string) (SystemInfo, error) {
	var info SystemInfo

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

func (Query) Set(info SystemInfo) error {
	encoded, err := info.encode()
	if err != nil {
		return err
	}

	err = db.Put([]byte(info.Name), encoded)
	if err != nil {
		return err
	}
	return nil
}

func (Query) Delete(name string) error {
	return db.Delete([]byte(name))
}

func (info SystemInfo) encode() ([]byte, error) {
	encoded, err := json.Marshal(info)
	if err != nil {
		return nil, errors.InvalidJSON{Message: err.Error()}
	}
	return encoded, nil
}

func decode(data []byte) (SystemInfo, error) {
	var info SystemInfo
	err := json.Unmarshal(data, &info)
	if err != nil {
		return info, errors.InvalidJSON{Message: err.Error()}
	}
	return info, nil
}
