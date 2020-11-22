/*******************************************************************************
 * Copyright 2020 Samsung Electronics All Rights Reserved.
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

package mnedcmgr

const (
	logPrefix            = "[mnedcmgr]"
	mnedcServerPort      = 3334
	broadcastServerPort  = 3333
	mnedcServerVirtualIP = "10.0.0.1"
	internalPort         = 56002
	maxAttempts          = 5
)

type ipTypes struct {
	virtualIP string
	privateIP string
}

type requestData struct {
	DeviceID  string `json:"deviceID"`
	PrivateIP string `json:"privateIP"`
	VirtualIP string `json:"virtualIP"`
}
