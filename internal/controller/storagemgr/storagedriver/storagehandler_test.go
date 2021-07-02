//
// Copyright (c) 2019 Intel Corporation
// Copyright (c) 2020 Samsung Electronics Co., Ltd All Rights Reserved.
// Copyright (c) 2021 IOTech Ltd
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
	"testing"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var handler *StorageHandler

func TestMain(m *testing.M) {
	service := &sdk.DeviceService{}
	logger := logger.NewMockClient()
	asyncValues := make(chan<- *models.AsyncValues)

	handler = NewStorageHandler(service, logger, asyncValues)
	os.Exit(m.Run())
}

func TestNewCommandValue(t *testing.T) {

	tests := []struct {
		Name          string
		Value         interface{}
		Expected      interface{}
		Type          string
		ErrorExpected bool
	}{
		{"Test A Binary", []byte{1, 0, 0, 1}, []byte{1, 0, 0, 1}, common.ValueTypeBinary, false},
		{"Test A String JSON", `{"name" : "My JSON"}`, `{"name" : "My JSON"}`, common.ValueTypeString, false},
		{"Test A String Text", "Random Text", "Random Text", common.ValueTypeString, false},
		{"Test A Bool true", "true", true, common.ValueTypeBool, false},
		{"Test A Bool false", "false", false, common.ValueTypeBool, false},
		{"Test A Bool error", "bad", nil, common.ValueTypeBool, true},
		{"Test A Float32+", "123.456", float32(123.456), common.ValueTypeFloat32, false},
		{"Test A Float32-", "-123.456", float32(-123.456), common.ValueTypeFloat32, false},
		{"Test A Float32 error", "-123.junk", nil, common.ValueTypeFloat32, true},
		{"Test A Float64+", "456.123", 456.123, common.ValueTypeFloat64, false},
		{"Test A Float64-", "-456.123", -456.123, common.ValueTypeFloat64, false},
		{"Test A Float64 error", "Random", nil, common.ValueTypeFloat64, true},
		{"Test A Uint8", "255", uint8(255), common.ValueTypeUint8, false},
		{"Test A Uint8 error", "FF", nil, common.ValueTypeUint8, true},
		{"Test A Uint16", "65535", uint16(65535), common.ValueTypeUint16, false},
		{"Test A Uint16 error", "FFFF", nil, common.ValueTypeUint16, true},
		{"Test A Uint32", "4294967295", uint32(4294967295), common.ValueTypeUint32, false},
		{"Test A Uint32 error", "FFFFFFFF", nil, common.ValueTypeUint32, true},
		{"Test A Uint64", "6744073709551615", uint64(6744073709551615), common.ValueTypeUint64, false},
		{"Test A Uint64 error", "FFFFFFFFFFFFFFFF", nil, common.ValueTypeUint64, true},
		{"Test A Int8+", "101", int8(101), common.ValueTypeInt8, false},
		{"Test A Int8-", "-101", int8(-101), common.ValueTypeInt8, false},
		{"Test A Int8 error", "-101.98", nil, common.ValueTypeInt8, true},
		{"Test A Int16+", "2001", int16(2001), common.ValueTypeInt16, false},
		{"Test A Int16-", "-2001", int16(-2001), common.ValueTypeInt16, false},
		{"Test A Int16 error", "-FF", nil, common.ValueTypeInt16, true},
		{"Test A Int32+", "32000", int32(32000), common.ValueTypeInt32, false},
		{"Test A Int32-", "-32000", int32(-32000), common.ValueTypeInt32, false},
		{"Test A Int32 error", "-32.456", nil, common.ValueTypeInt32, true},
		{"Test A Int64+", "214748364800", int64(214748364800), common.ValueTypeInt64, false},
		{"Test A Int64-", "-214748364800", int64(-214748364800), common.ValueTypeInt64, false},
		{"Test A Int64 error", "-21474.99", nil, common.ValueTypeInt64, true},
	}

	for _, currentTest := range tests {
		t.Run(currentTest.Name, func(t *testing.T) {
			cmdVal, err := handler.newCommandValue("test", currentTest.Value, currentTest.Type)
			if currentTest.ErrorExpected {
				assert.Error(t, err, "Expected an Error")
			} else {
				require.NoError(t, err, "Unexpected an Error")
				assert.Equal(t, cmdVal.Value, currentTest.Expected)
			}
		})
	}
}

func TestReadasBinary(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://0.0.0.0:12345", nil)
	handler.readBodyAsString(nil, req)
	handler.readBodyAsBinary(nil, req)
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
