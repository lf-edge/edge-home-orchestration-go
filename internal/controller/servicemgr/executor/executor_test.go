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

package executor

import (
	"testing"

	"github.com/lf-edge/edge-home-orchestration-go/internal/controller/servicemgr/notification"
	"github.com/lf-edge/edge-home-orchestration-go/internal/restinterface/client/restclient"
)

const (
	unexpectedSuccess = "unexpected success"
	unexpectedFail    = "unexpected fail"
)

func TestSetClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		restIns := restclient.GetRestClient()
		c := new(HasClientNotification)

		c.SetNotiImpl(notification.GetInstance())
		if c.NotiImplIns == nil {
			t.Error(unexpectedFail)
		}

		c.SetClient(restIns)
		if restIns != c.Clienter {
			t.Error(unexpectedFail)
		}
	})
}
