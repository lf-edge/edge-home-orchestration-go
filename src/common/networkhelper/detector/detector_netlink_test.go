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
package detector

import (
	"testing"

	"net"

	"github.com/vishvananda/netlink"
)

var originAddrSubscribe func(ch chan<- netlink.AddrUpdate, done <-chan struct{}) error
var lambdaAddrSubscribe func(ch chan<- netlink.AddrUpdate) error

func TestAddrSubscribe(t *testing.T) {
	originAddrSubscribe = addrSubscribe
	addrSubscribe = fakeAddrSubscribe
	defer func() {
		addrSubscribe = originAddrSubscribe
	}()

	retTrue := make(chan bool, 1)

	t.Run("Success", func(t *testing.T) {
		t.Run("Connected", func(t *testing.T) {
			lambdaAddrSubscribe = func(ch chan<- netlink.AddrUpdate) error {
				ch <- netlink.AddrUpdate{
					NewAddr: true,
					LinkAddress: net.IPNet{
						IP: []byte{'1', '2', '3', '4'},
					},
				}
				return nil
			}
			go GetInstance().AddrSubscribe(retTrue)

			ret := <-retTrue
			if !ret {
				t.Error("unexpected result")
			}
		})
		t.Run("Disconnected", func(t *testing.T) {
			lambdaAddrSubscribe = func(ch chan<- netlink.AddrUpdate) error {
				ch <- netlink.AddrUpdate{
					NewAddr: false,
					LinkAddress: net.IPNet{
						IP: []byte{'1', '2', '3', '4'},
					},
				}
				return nil
			}
			go GetInstance().AddrSubscribe(retTrue)

			ret := <-retTrue
			if !ret {
				t.Error("unexpected result")
			}
		})
	})
	t.Run("Error", func(t *testing.T) {
		t.Run("IPisEmpty", func(t *testing.T) {
			lambdaAddrSubscribe = func(ch chan<- netlink.AddrUpdate) error {
				ch <- netlink.AddrUpdate{
					NewAddr: true,
					LinkAddress: net.IPNet{
						IP: []byte{},
					},
				}
				return nil
			}
			go GetInstance().AddrSubscribe(retTrue)

			ret := <-retTrue
			if ret {
				t.Error("unexpected result")
			}
		})
	})

}

func fakeAddrSubscribe(ch chan<- netlink.AddrUpdate, done <-chan struct{}) error {
	return lambdaAddrSubscribe(ch)
}
