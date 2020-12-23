//
// Copyright (c) 2020 Samsung Electronics Co., Ltd All Rights Reserved.
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package storagedriver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/edgexfoundry/device-sdk-go"
	"github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

var handler *StorageHandler

func TestMain(m *testing.M) {
	service := &sdk.Service{}
	logger := logger.NewClient("test", false, "./tests.log", "DEBUG")
	asyncValues := make(chan<- *models.AsyncValues)

	handler = NewStorageHandler(service, logger, asyncValues)
	os.Exit(m.Run())
}

func TestReadasBinary(t *testing.T) {

	req := httptest.NewRequest(http.MethodPost, "http://0.0.0.0:12345", nil)

	handler.readBodyAsString(nil, req)

	handler.readBodyAsBinary(nil, req)
}

func TestNewCommandValue(t *testing.T) {

	tests := []struct {
		Name          string
		Expected      interface{}
		Type          models.ValueType
		ErrorExpected bool
	}{
		{"Test A Binary", []byte{1, 0, 0, 1}, models.Binary, false},
		{"Test A String JSON", `{"name" : "My JSON"}`, models.String, false},
		{"Test A String Text", "Random Text", models.String, false},
		{"Test A Bool true", "true", models.Bool, false},
		{"Test A Bool false", "false", models.Bool, false},
		{"Test A Bool error", "bad", models.Bool, true},
		{"Test A Uint8", "255", models.Uint8, false},
		{"Test A Uint8 error", "FF", models.Uint8, true},
		{"Test A Uint16", "65535", models.Uint16, false},
		{"Test A Uint16 error", "FFFF", models.Uint16, true},
		{"Test A Uint32", "4294967295", models.Uint32, false},
		{"Test A Uint32 error", "FFFFFFFF", models.Uint32, true},
		{"Test A Uint64", "18446744073709551615", models.Uint64, false},
		{"Test A Uint64 error", "FFFFFFFFFFFFFFFF", models.Uint64, true},
		{"Test A Int8+", "101", models.Int8, false},
		{"Test A Int8-", "-101", models.Int8, false},
		{"Test A Int8 error", "-101.98", models.Int8, true},
		{"Test A Int16+", "2001", models.Int16, false},
		{"Test A Int16-", "-2001", models.Int16, false},
		{"Test A Int16 error", "-FF", models.Int16, true},
		{"Test A Int32+", "32000", models.Int32, false},
		{"Test A Int32-", "-32000", models.Int32, false},
		{"Test A Int32 error", "-32.456", models.Int32, true},
		{"Test A Int64+", "214748364800", models.Int64, false},
		{"Test A Int64-", "-214748364800", models.Int64, false},
		{"Test A Int64 error", "-21474.99", models.Int64, true},
	}

	for _, currentTest := range tests {
		t.Run(currentTest.Name, func(t *testing.T) {
			cmdVal, err := handler.newCommandValue("test", currentTest.Expected, currentTest.Type)
			if currentTest.ErrorExpected {
				if err == nil {
					t.Fatal("Expected an Error")
				}
				return
			}

			if !currentTest.ErrorExpected && err != nil {
				t.Fatal("Unexpected an Error")
			}

			var actual interface{}

			switch currentTest.Type {
			case models.Binary:
				actual, err = cmdVal.BinaryValue()
				if !assert.NoError(t, err) {
					t.Fatal()
				}
			case models.String:
				actual, err = cmdVal.StringValue()
				if !assert.NoError(t, err) {
					t.Fatal()
				}
			case models.Bool:
				value, err := cmdVal.BoolValue()
				if !assert.NoError(t, err) {
					t.Fatal()
				}
				actual = strconv.FormatBool(value)
			case models.Uint8:
				value, err := cmdVal.Uint8Value()
				if !assert.NoError(t, err) {
					t.Fatal()
				}
				actual = strconv.FormatUint(uint64(value), 10)
			case models.Uint16:
				value, err := cmdVal.Uint16Value()
				if !assert.NoError(t, err) {
					t.Fatal()
				}
				actual = strconv.FormatUint(uint64(value), 10)
			case models.Uint32:
				value, err := cmdVal.Uint32Value()
				if !assert.NoError(t, err) {
					t.Fatal()
				}
				actual = strconv.FormatUint(uint64(value), 10)
			case models.Uint64:
				value, err := cmdVal.Uint64Value()
				if !assert.NoError(t, err) {
					t.Fatal()
				}
				actual = strconv.FormatUint(value, 10)
			case models.Int8:
				value, err := cmdVal.Int8Value()
				if !assert.NoError(t, err) {
					t.Fatal()
				}
				actual = strconv.FormatInt(int64(value), 10)
			case models.Int16:
				value, err := cmdVal.Int16Value()
				if !assert.NoError(t, err) {
					t.Fatal()
				}
				actual = strconv.FormatInt(int64(value), 10)
			case models.Int32:
				value, err := cmdVal.Int32Value()
				if !assert.NoError(t, err) {
					t.Fatal()
				}
				actual = strconv.FormatInt(int64(value), 10)
			case models.Int64:
				value, err := cmdVal.Int64Value()
				if !assert.NoError(t, err) {
					t.Fatal()
				}
				actual = strconv.FormatInt(value, 10)
			}

			assert.Equal(t, currentTest.Expected, actual)
		})
	}
}

func TestStorageDriver(t *testing.T) {
	sd := StorageDriver{}
	sd.AddDevice("TestDevice", nil, "LOCKED")
	sd.UpdateDevice("TestDevuce", nil, "UNLOCKED")
	sd.HandleReadCommands("TestDevice", nil, nil)
	sd.HandleWriteCommands("TestDeviuce", nil, nil, nil)
	sd.RemoveDevice("TestingDevice", nil)
	sd.Stop(true)
}
