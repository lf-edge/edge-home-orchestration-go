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
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"

)

type networkUtilImpl struct{}

var (
	networkUtilIns networkUtilImpl
	log            = logmgr.GetInstance()
)

func init() {
	// Do nothing because there is no need to initialize anything
}

func createClientConfig() (*tls.Config, error) {
	caCertPEM, err := ioutil.ReadFile("/var/edge-orchestration/data/cert/ca.crt")
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertPEM)
	if !ok {
		panic("failed to parse root certificate")
	}

	cert, err := tls.LoadX509KeyPair("/var/edge-orchestration/data/cert/hen.crt", "/var/edge-orchestration/data/cert/hen.key")
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      roots,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
            tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_RSA_WITH_AES_256_CBC_SHA,
        },
	}, nil
}

func createServerConfig() (*tls.Config, error) {
	caCertPEM, err := ioutil.ReadFile("/var/edge-orchestration/data/cert/ca.crt")
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertPEM)
	if !ok {
		panic("failed to parse root certificate")
	}

	cert, err := tls.LoadX509KeyPair("/var/edge-orchestration/data/cert/hen.crt", "/var/edge-orchestration/data/cert/hen.key")
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    roots,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
            tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_RSA_WITH_AES_256_CBC_SHA,
        },
	}, nil
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

	config, err := createClientConfig()
	if err != nil {
		log.Fatal("config failed: ", err)
	}

	conn, err := tls.Dial("tcp", ip+":"+port, config)
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

	config, err := createServerConfig()
	if err != nil {
		log.Fatal("config failed: ", err)
	}

	listener, err := tls.Listen("tcp", address, config)
	return listener, err
}
