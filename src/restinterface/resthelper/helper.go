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

// Package resthelper implements rest helper functions
package resthelper

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"

	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/resthelper/client"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/resthelper/client/httphelper"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/resthelper/client/tlshelper"
	"github.com/lf-edge/edge-home-orchestration-go/src/restinterface/tls"
)

// RestHelper is the interface implemented by rest helper functions
type RestHelper interface {
	urlHelper
	requestHelper
	responseHelper
}

type RestHelperWithCertificateSetter interface {
	urlHelper
	requestHelper
	responseHelper
	tls.CertificateSetter
}

type urlHelper interface {
	MakeTargetURL(target string, port int, restapi string) string
}

type requestHelper interface {
	DoGet(targetURL string) (respBytes []byte, statusCode int, err error)
	DoGetWithBody(targetURL string, bodybytes []byte) (respBytes []byte, statusCode int, err error)
	DoPost(targetURL string, bodybytes []byte) (respBytes []byte, statusCode int, err error)
	DoDelete(targetURL string) (respBytes []byte, statusCode int, err error)
}

type responseHelper interface {
	Response(w http.ResponseWriter, httpStatus int)
	ResponseJSON(w http.ResponseWriter, bytes []byte, httpStatus int)
}

type helperImpl struct {
	c client.Requester
	tls.HasCertificate
}

var (
	helper *helperImpl
	log    = logmgr.GetInstance()
)

func init() {
	helper = new(helperImpl)
}

// GetHelper returns helperImpl instance
func GetHelper() RestHelper {
	switch helper.IsSetCert {
	case true:
		helper.c = tlshelper.TLSHelper{}
	default:
		helper.c = httphelper.HttpHelper{}
	}
	return helper
}

func GetHelperWithCertificate() RestHelperWithCertificateSetter {
	return helper
}

// DoGet is for get request
func (h helperImpl) DoGet(targetURL string) (respBytes []byte, statusCode int, err error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return
	}

	resp, err := h.c.Do(req)
	if err != nil {
		log.Printf("[%v] reqeust get failed !!, err = %v", targetURL, err)
		return
	}

	defer resp.Body.Close()

	statusCode = resp.StatusCode
	respBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[%v] read resp.Body failed !!, err = %v", targetURL, err)
		return
	}

	return
}

// DoGetWithBody is for get request with req' body
func (h helperImpl) DoGetWithBody(targetURL string, bodybytes []byte) (respBytes []byte, statusCode int, err error) {
	if len(bodybytes) == 0 {
		log.Printf("DoGetWithBody body length is zero(0) !!")
	}

	buff := bytes.NewBuffer(bodybytes)

	req, err := http.NewRequest("GET", targetURL, buff)
	if err != nil {
		return
	}

	// Content-Type Header
	req.Header.Add("Content-Type", "application/json")

	resp, err := h.c.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	statusCode = resp.StatusCode
	respBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("read resp.Body failed !!", err)
	}

	return
}

// DoPost is for post request
func (h helperImpl) DoPost(targetURL string, bodybytes []byte) (respBytes []byte, statusCode int, err error) {
	if len(bodybytes) == 0 {
		log.Printf("DoPost body length is zero(0) !!")
	}

	buff := bytes.NewBuffer(bodybytes)

	req, err := http.NewRequest("POST", targetURL, buff)
	if err != nil {
		return
	}

	// Content-Type Header
	req.Header.Add("Content-Type", "application/json")

	resp, err := h.c.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	statusCode = resp.StatusCode
	respBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("read resp.Body failed !!", err)
		return
	}

	return
}

// DoDelete is for delete request
func (h helperImpl) DoDelete(targetURL string) (respBytes []byte, statusCode int, err error) {
	req, err := http.NewRequest("DELETE", targetURL, nil)
	if err != nil {
		return
	}

	resp, err := h.c.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	statusCode = resp.StatusCode
	respBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("read resp.Body failed !!", err)
		return
	}

	return
}

// MakeTargetURL function
func (h helperImpl) MakeTargetURL(target string, port int, restapi string) string {
	var protocol string
	if h.IsSetCert {
		protocol = "https"
	} else {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s:%d%s", protocol, target, port, restapi)
}

// ResponseJSON function
func (helperImpl) ResponseJSON(w http.ResponseWriter, bytes []byte, httpStatus int) {
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpStatus)
	w.Write(bytes)
}

// Response function
func (helperImpl) Response(w http.ResponseWriter, httpStatus int) {
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpStatus)
}
