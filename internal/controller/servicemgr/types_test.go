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

package servicemgr

import (
	"strings"
	"testing"
)

func TestMapSetGet(t *testing.T) {
	ServiceMap.Set(uint64(1), serviceName)

	str, _ := ServiceMap.Get(uint64(1))
	assertEqualStr(t, str.(string), serviceName)
}

func TestMapRemove(t *testing.T) {
	ServiceMap.Set(uint64(1), serviceName)

	ServiceMap.Remove(uint64(1))

	exist, _ := ServiceMap.Get(uint64(1))
	if exist != nil {
		t.Error("ConcurrentMap Remove API is failed")
	}
}

func TestMapModify(t *testing.T) {
	ServiceMap.Set(uint64(1), serviceName)
	ServiceMap.Set(uint64(1), serviceName2)

	str, _ := ServiceMap.Get(uint64(1))
	assertEqualStr(t, str.(string), serviceName2)
}

func TestMapIter(t *testing.T) {
	ServiceMap.Set(uint64(1), serviceName)
	ServiceMap.Set(uint64(2), serviceName2)

	mapItem := ServiceMap.Iter()
	compareStr := make([]string, 10)
	idx := 0

	for {
		msg := <-mapItem

		if msg.Value == nil && msg.Key == 0 {
			break
		} else {
			compareStr[idx] = msg.Value.(string)
		}
	}

	for _, str := range compareStr {
		if len(str) == 0 {
			break
		}

		if strings.Compare(str, serviceName) == 0 || strings.Compare(str, serviceName2) == 0 {
			continue
		} else {
			t.Fail()
		}
	}
}
