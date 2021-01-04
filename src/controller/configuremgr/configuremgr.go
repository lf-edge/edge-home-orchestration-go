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

// Package configuremgr provides interfaces between orchestrationapi and configuremgr
package configuremgr

import "github.com/lf-edge/edge-home-orchestration-go/src/common/types/configuremgrtypes"

// Notifier is the interface to get scoring information for each service application
type Notifier interface {
	Notify(serviceinfo configuremgrtypes.ServiceInfo)
}

// Watcher is the interface to check if service application is installed/updated/deleted
type Watcher interface {
	Watch(notifier Notifier)
}
