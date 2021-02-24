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

// Package mockscoringmgr implements the mock functions for scoringmgr
package mockscoringmgr

/*
int
wrap_add(void *f, int a, int b){
  return ((int (*)(int a, int b))f)(a,b);
}
*/
import "C"
import (
	"log"
)

// StartResourceServiceAdd is a mock function
func StartResourceServiceAdd() {
	log.Println("[mockscoringmgr] startResourceService of add library not meaning")
}

// StopResourceServiceAdd is a mock function
func StopResourceServiceAdd() {
	log.Println("[mockscoringmgr] stopResourceService of add library not meaning")
}
