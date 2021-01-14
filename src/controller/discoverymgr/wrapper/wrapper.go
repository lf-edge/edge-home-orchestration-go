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

// Package wrapper serves wrapper functions of zeroconf for orchestration
package wrapper

import (
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
	"net"

	"github.com/grandcat/zeroconf"
)

const logPrefix = "[discovery][wrapper]"
var (
	log = logmgr.GetInstance()
)

// ZeroconfInterface is the interface implemented by wrapped functions using zeroconf
// ToDo : How to deal w/ data conver function?
type ZeroconfInterface interface {
	RegisterProxy(string, string, string, int,
		string, []string, []string,
		[]net.Interface) (Entity, error)
	GetSubscriberChan() (chan *Entity, error)
	ResetServer([]net.IP)
	Advertise()
	SetText([]string)
	GetText() []string
	Shutdown()
}

// Entity provides wrapper entity info
type Entity struct {
	DeviceID          string
	TTL               uint32
	OrchestrationInfo OrchestrationInformation
}

// OrchestrationInformation provides orchestration info
type OrchestrationInformation struct {
	IPv4          []string
	Platform      string
	ExecutionType string
	ServiceList   []string
}

// ZeroconfImpl struct
type ZeroconfImpl struct {
	server *zeroconf.Server
}

// ZeroconfIns struct
var ZeroconfIns *ZeroconfImpl

func init() {
	ZeroconfIns = new(ZeroconfImpl)
}

// GetZeroconfImpl provides ZeroconfImpl instance which have zerconf server address
func GetZeroconfImpl() ZeroconfInterface {
	return ZeroconfIns
}

// RegisterProxy is the wrapper of zeroconf.RegisterProxy
func (zc *ZeroconfImpl) RegisterProxy(instance, service, domain string,
	port int, host string, ips []string, text []string,
	ifaces []net.Interface) (Entity, error) {

	server, serviceEntry, err := zeroconf.EdgeRegisterProxy(instance, service,
		domain, port, host, ips, text, ifaces)
	if err != nil {
		return Entity{}, err
	}
	zc.server = server
	entity := Entity{
		DeviceID:          serviceEntry.ServiceRecord.Instance,
		TTL:               serviceEntry.TTL,
		OrchestrationInfo: convertServiceEntrytoDB(serviceEntry)}
	return entity, nil

}

// GetSubscriberChan subscribers serviceEntry info of other devices
func (zc ZeroconfImpl) GetSubscriberChan() (chan *Entity, error) {
	subchan, err := zeroconf.EdgeGetSubscriberChan()
	if err != nil {
		log.Println(logPrefix, err)
		return nil, err
	}
	exportchan := make(chan *Entity, 32)

	go func() {
		for {
			select {
			case data := <-subchan:
				// log.Println(logPrefix, "[detectDevice] ", data)
				if data == nil {
					select {
					case exportchan <- nil:
					default:
						log.Println("send Chan Full")
					}
					continue
				}
				entity := Entity{
					DeviceID:          data.ServiceRecord.Instance,
					TTL:               data.TTL,
					OrchestrationInfo: convertServiceEntrytoDB(data)}
				select {
				case exportchan <- &entity:
				default:
					log.Println("send Chan Full")
				}
				// case default:
				//resource return
			}
		}
	}()

	return exportchan, err
}

// ResetServer resets local server
func (zc ZeroconfImpl) ResetServer(ips []net.IP) {
	zc.server.EdgeResetServer(ips)
}

// Advertise advertises local server to other servers
func (zc ZeroconfImpl) Advertise() {
	zc.server.EdgeAdvertise()
}

// SetText sets text field
func (zc ZeroconfImpl) SetText(txt []string) {
	zc.server.SetText(txt)
}

// GetText gets text field
func (zc ZeroconfImpl) GetText() []string {
	return zc.server.EdgeGetText()
}

// Shutdown shutdowns local server
func (zc ZeroconfImpl) Shutdown() {
	zc.server.Shutdown()
}

// convertServiceEntrytoDB converts zeroconf data format to OrchestrationInformation format
func convertServiceEntrytoDB(data *zeroconf.ServiceEntry) (newDevice OrchestrationInformation) {
	for _, val := range data.AddrIPv4 {
		newDevice.IPv4 = append(newDevice.IPv4, val.String())
	}
	//Todo : Remove
	//tmp error defense code
	//from old version
	if len(data.Text) < 2 {
		newDevice.ServiceList = data.Text
	} else {
		newDevice.Platform = data.Text[0]
		newDevice.ExecutionType = data.Text[1]
		if len(data.Text) > 2 {
			newDevice.ServiceList = append(newDevice.ServiceList, data.Text[2:]...)
		}
	}
	return
}
