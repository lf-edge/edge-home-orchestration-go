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

package discoverymgr

const (
	//fixed service type, search only fixed servicetype
	serviceType = "_orchestration._tcp"
	//mDNS only support local. domain
	domain      = "local."
	servicePort = 42425
	//max txt size of mdns service
	maxTXTSize = 400
	//Interval Second for active discovery
	discoveryInterval = 60 * 60
	//IP Code for Active Discovery
	ipv4                     = 0x01
	ipv6                     = 0x02
	ipv4Andipv6              = (ipv4 | ipv6) //< Default option.
	mnedcBroadcastServerPort = 3333
)
