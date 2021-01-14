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

package connectionutil

import (
	"errors"
	"net"

	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/tls"

	rafftls "github.com/raff/tls-ext"
	"github.com/raff/tls-psk"
)

type networkUtilImpl struct{}

var networkUtilIns networkUtilImpl

func init() {

}

//NetworkUtil interface declares the network methods
type NetworkUtil interface {
	ConnectToHost(string, string, bool) (net.Conn, error)
	WriteTo(net.Conn, []byte) error
	ReadFrom(net.Conn) (int, []byte, error)
	ListenIP(address string, isSecure bool) (net.Listener, error)
}

//GetInstance returns the NetworkUtil instance
func GetInstance() NetworkUtil {
	return networkUtilIns
}

//ConnectToHost connects to a tcp host
func (networkUtilImpl) ConnectToHost(ip string, port string, isSecure bool) (net.Conn, error) {
	if !isSecure {
		conn, err := net.Dial("tcp", ip+":"+port)
		return conn, err
	}
	var config = &rafftls.Config{
		CipherSuites: []uint16{psk.TLS_PSK_WITH_AES_128_CBC_SHA},
		Certificates: []rafftls.Certificate{rafftls.Certificate{}},
		Extra: psk.PSKConfig{
			GetKey:      tls.GetKey,
			GetIdentity: tls.GetIdentity,
		},
	}

	conn, err := rafftls.Dial("tcp", ip+":"+port, config)
	return conn, err

}

//WriteTo writes on a connection
func (networkUtilImpl) WriteTo(conn net.Conn, data []byte) error {
	if conn != nil {
		_, err := conn.Write(data)
		return err
	}
	return errors.New("Connection is nil")
}

//ReadFrom reads from a connection
func (networkUtilImpl) ReadFrom(conn net.Conn) (int, []byte, error) {
	if conn != nil {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		return n, buf, err
	}
	return -1, []byte(""), errors.New("Connection is nil")
}

//ListenIP starts tcp server at given address
func (networkUtilImpl) ListenIP(address string, isSecure bool) (net.Listener, error) {
	if !isSecure {
		listener, err := net.Listen("tcp", address)
		return listener, err
	}
	var config = &rafftls.Config{
		CipherSuites: []uint16{psk.TLS_PSK_WITH_AES_128_CBC_SHA},
		Certificates: []rafftls.Certificate{rafftls.Certificate{}},
		Extra: psk.PSKConfig{
			GetKey:      tls.GetKey,
			GetIdentity: tls.GetIdentity,
		},
	}

	listener, err := rafftls.Listen("tcp", address, config)
	return listener, err
}
