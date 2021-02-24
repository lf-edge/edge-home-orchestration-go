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

package mockscoringmgr

import (
	"log"
	"math/rand"
	"time"
)

// GetScoreRandom100Mock is a mock function
func GetScoreRandom100Mock(endpoint string, libName string) (score float64, err error) {

	log.Printf("libName : %s ", libName)
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)

	score = random.Float64() * 100

	return
}
