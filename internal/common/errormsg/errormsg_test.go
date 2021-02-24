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

package errormsg

import (
	"strings"
	"testing"
)

func TestToString(t *testing.T) {
	errString := ToString(ErrorNoNetworkInterface)
	assertEqualStr(t, errString, orchestrationErrorString[ErrorNoNetworkInterface*(-1)])

	realErrString := ToString("string")
	assertEqualStr(t, realErrString, "NOT SUPPORT TYPE")
}

func TestToInt(t *testing.T) {
	err := ToError(ErrorNoNetworkInterface)
	num := ToInt(err)
	if num != ErrorNoNetworkInterface {
		t.Errorf("%d != %d", num, ErrorNoNetworkInterface)
	}
}

func assertEqualStr(t *testing.T, a, b string) {
	t.Helper()
	if strings.Compare(a, b) != 0 {
		t.Errorf("%s != %s", a, b)
	}
}
