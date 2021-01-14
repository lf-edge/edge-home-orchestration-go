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

package tunmgr

import (
	"fmt"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
	"net"
	"os/exec"
	"strings"

	"github.com/songgao/water"
)

//TunImpl is a struct for Tun related methods
type TunImpl struct{}

var (
	tunIns TunImpl
	log    = logmgr.GetInstance()
)

func init() {

}

//Tun interface declares methods related to setting tun network interface
type Tun interface {
	CreateTUN() (*water.Interface, error)
	SetTUNIP(iName string, localAddr net.IP, addr *net.IPNet, debug bool) error
	SetTUNStatus(iName string, up bool, debug bool) error
}

//GetInstance returns Tun interface instance
func GetInstance() Tun {
	return tunIns
}

//CreateTUN creates a virtual tun interface
func (TunImpl) CreateTUN() (*water.Interface, error) {
	intf, err := water.NewTUN("")
	return intf, err
}

//SetTUNIP sets an IP to the interface
func (TunImpl) SetTUNIP(iName string, localAddr net.IP, addr *net.IPNet, debug bool) error {
	sargs := fmt.Sprintf("%s %s netmask %s", iName, localAddr.String(), net.IP(addr.Mask).String())
	args := strings.Split(sargs, " ")
	return commandExec("ifconfig", args, debug)
}

//SetTUNStatus starts/stops an interface
func (TunImpl) SetTUNStatus(iName string, up bool, debug bool) error {
	statusString := "down"
	if up {
		statusString = "up"
	}
	sargs := fmt.Sprintf("link set dev %s %s mtu %d qlen %d", iName, statusString, 1400, 300)
	args := strings.Split(sargs, " ")
	return commandExec("ip", args, debug)
}

func commandExec(command string, args []string, debug bool) error {
	cmd := exec.Command(command, args...)
	e := cmd.Run()
	if e != nil {
		log.Println("Command failed: ", e.Error())
	}
	return e
}
