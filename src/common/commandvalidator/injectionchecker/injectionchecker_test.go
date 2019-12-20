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

package injectionchecker

import (
	"testing"
)

func TestHasInjectionOperator(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		for _, str := range injectionOperators {
			if HasInjectionOperator(str) != true {
				t.Error("unexpected success")
			}
		}
	})
	t.Run("Success", func(t *testing.T) {
		str := "12345"
		if HasInjectionOperator(str) != false {
			t.Error("unexpected error")
		}
	})
}
