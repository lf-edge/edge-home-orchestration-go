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

package blacklist

import (
	"github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/common"
)

var blackList = []string{
	"sudo",
	"su",
	"bash",
	"bsh",
	"csh",
	"adb",
	"sh",
	"ssh",
	"scp",
	"cat",
	"chage",
	"chpasswd",
	"dmidecode",
	"dmsetup",
	"fcinfo",
	"fdisk",
	"iscsiadm",
	"lsof",
	"multipath",
	"oratab",
	"prtvtoc",
	"ps",
	"pburn",
	"pfexec",
	"dzdo",
}

func IsBlack(command string) bool {
	return common.HasElem(blackList, command)
}
