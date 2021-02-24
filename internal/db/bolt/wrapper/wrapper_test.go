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
package wrapper

import (
	"testing"

	"os"
)

type data struct {
	key   []byte
	value []byte
}

const (
	testPath       = "./test"
	testBucketName = "test"
)

var (
	testData = []data{
		data{key: []byte("testkey1"), value: []byte("testvalue1")},
		data{key: []byte("testkey2"), value: []byte("testvalue2")},
		data{key: []byte("testkey3"), value: []byte("testvalue3")},
	}
)

func TestSetBoltDBPath(t *testing.T) {
	SetBoltDBPath(testPath)
	defer os.RemoveAll(testPath)
	t.Run("Success", func(t *testing.T) {

		if dbPath != testPath+"/data.db" {
			t.Error("unexpected value")
		}
	})
}

func TestList(t *testing.T) {
	SetBoltDBPath(testPath)
	defer os.RemoveAll(testPath)
	d := insertTestData()
	t.Run("Success", func(t *testing.T) {
		m, err := d.List()
		if err != nil {
			t.Error("unexpected error")
		}

		for _, data := range testData {
			val, exists := m[string(data.key)]
			if !exists {
				t.Error("expect exist value, but it is not exist")
			} else if val.(string) != string(data.value) {
				t.Error("expect same value, but it is not same")
			}
		}
	})
}

func TestGet(t *testing.T) {
	SetBoltDBPath(testPath)
	defer os.RemoveAll(testPath)
	d := insertTestData()
	t.Run("Success", func(t *testing.T) {

		for _, data := range testData {
			val, err := d.Get(data.key)
			if err != nil {
				t.Error("expect exist value, but it is not exist")
			} else if string(val) != string(data.value) {
				t.Error("expect same value, but it is not same")
			}
		}
	})
	t.Run("Error", func(t *testing.T) {
		t.Run("NotExistKey", func(t *testing.T) {
			_, err := d.Get([]byte("testkey4"))
			if err == nil {
				t.Error("unexpect success")
			}
		})
	})
}

func TestDelete(t *testing.T) {
	SetBoltDBPath(testPath)
	defer os.RemoveAll(testPath)
	d := insertTestData()
	t.Run("Success", func(t *testing.T) {
		err := d.Delete(testData[0].key)
		if err != nil {
			t.Error("unexpected error")
		}

		m, err := d.List()
		if err != nil {
			t.Error("unexpected error")
		}

		_, exists := m[string(testData[0].key)]
		if exists {
			t.Error("expect not exist value, but it exist")
		}

		for i := 1; i <= 2; i++ {
			val, exists := m[string(testData[i].key)]
			if !exists {
				t.Error("expect exist value is not exist")
			} else if val.(string) != string(testData[i].value) {
				t.Error("expect same value, but it is not same")
			}
		}
	})
}

func insertTestData() Database {
	d := NewBoltDB(testBucketName)
	for _, tData := range testData {
		d.Put(tData.key, tData.value)
	}
	return d
}
