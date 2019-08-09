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

package native

import (
	errs "common/errors"
	"testing"
)

var (
	dummyInvalidError = errs.InvalidParam{Message: "dummyError"}
	dummySystemError  = errs.SystemError{Message: "dummyError"}
	dummyNotSupport   = errs.NotSupport{Message: "dummyError"}
)

func init() {
}

func TestErrorConvertToEnumWithInvalidParam_ExpectedReturnInvalidParamEnum(t *testing.T) {
	ret := errorCovertToEnum(dummyInvalidError)

	if ret != invalidParamError {
		t.Errorf("Unexpected enum value : %d", ret)
	}
}

func TestErrorConvertToEnumWithSystemError_ExpectedReturnSystemErrorEnum(t *testing.T) {
	ret := errorCovertToEnum(dummySystemError)

	if ret != systemError {
		t.Errorf("Unexpected enum value : %d", ret)
	}
}

func TestErrorConvertToEnumWithNotSupport_ExpectedReturnNotSupportEnum(t *testing.T) {
	ret := errorCovertToEnum(dummyNotSupport)

	if ret != notSupportedError {
		t.Errorf("Unexpected enum value : %d", ret)
	}
}
