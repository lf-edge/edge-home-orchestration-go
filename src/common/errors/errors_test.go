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
package errors

import (
	"testing"

	"strings"
)

func TestInvalidParam(t *testing.T) {
	err := InvalidParam{Message: "test"}
	if !strings.Contains(err.Error(), "invalid param : ") {
		t.Error("unexpected result : " + err.Error())
	}
}

func TestSystemError(t *testing.T) {
	err := SystemError{Message: "test"}
	if !strings.Contains(err.Error(), "system error : ") {
		t.Error("unexpected result : " + err.Error())
	}
}

func TestNotSupport(t *testing.T) {
	err := NotSupport{Message: "test"}
	if !strings.Contains(err.Error(), "not support error : ") {
		t.Error("unexpected result : " + err.Error())
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound{Message: "test"}
	if !strings.Contains(err.Error(), "not found error : ") {
		t.Error("unexpected result : " + err.Error())
	}
}

func TestDBConnectionError(t *testing.T) {
	err := DBConnectionError{Message: "test"}
	if !strings.Contains(err.Error(), "db connection error : ") {
		t.Error("unexpected result : " + err.Error())
	}
}

func TestDBOperationError(t *testing.T) {
	err := DBOperationError{Message: "test"}
	if !strings.Contains(err.Error(), "db operation error : ") {
		t.Error("unexpected result : " + err.Error())
	}
}

func TestInvalidJSON(t *testing.T) {
	err := InvalidJSON{Message: "test"}
	if !strings.Contains(err.Error(), "invalid json error : ") {
		t.Error("unexpected result : " + err.Error())
	}
}

func TestNetworkError(t *testing.T) {
	err := NetworkError{Message: "test"}
	if !strings.Contains(err.Error(), "network error : ") {
		t.Error("unexpected result : " + err.Error())
	}
}
