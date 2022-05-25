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

package resthelper

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

const (
	unexpectedSuccess = "unexpected success"
	unexpectedFail    = "unexpected fail"
)

var (
	expectMethod = http.MethodGet
	statusCode   = http.StatusOK
	handler      = func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectMethod {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(statusCode)
	}
)

func TestGetHelper(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		helper := GetHelper()
		if helper == nil {
			t.Error("unexpected return value")
		}
	})
}

func getTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestDoGet(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ts := getTestServer(handler)
		defer ts.Close()

		_, code, err := GetHelper().DoGet(ts.URL)
		if code != http.StatusOK {
			t.Error("unexpected error code " + http.StatusText(code))
		} else if err != nil {
			t.Error("unexpected error " + err.Error())
		}
	})
	t.Run("Fail", func(t *testing.T) {
		_, _, err := GetHelper().DoGet("")
		if err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestDoGetWithBody(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ts := getTestServer(handler)
		defer ts.Close()

		_, code, err := GetHelper().DoGetWithBody(ts.URL, make([]byte, 0))
		if code != http.StatusOK {
			t.Error("unexpected error code " + http.StatusText(code))
		} else if err != nil {
			t.Error("unexpected error " + err.Error())
		}
	})
	t.Run("Fail", func(t *testing.T) {
		_, _, err := GetHelper().DoGetWithBody("", nil)
		if err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestDoPost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectMethod = http.MethodPost
		ts := getTestServer(handler)
		defer ts.Close()

		_, code, err := GetHelper().DoPost(ts.URL, make([]byte, 0))
		if code != http.StatusOK {
			t.Error("unexpected error code " + http.StatusText(code))
		} else if err != nil {
			t.Error("unexpected error " + err.Error())
		}
	})
	t.Run("Fail", func(t *testing.T) {
		_, _, err := GetHelper().DoPost("", make([]byte, 0))
		if err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestDoDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectMethod = http.MethodDelete
		ts := getTestServer(handler)
		defer ts.Close()

		_, code, err := GetHelper().DoDelete(ts.URL)
		if code != http.StatusOK {
			t.Error("unexpected error code " + http.StatusText(code))
		} else if err != nil {
			t.Error("unexpected error " + err.Error())
		}
	})
	t.Run("Fail", func(t *testing.T) {
		_, _, err := GetHelper().DoDelete("")
		if err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func TestMakeTargetURL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("GetHelper-MakeTargetURL", func(t *testing.T) {
			target := "testserver.test"
			port := 1234
			restapi := "/api/v1/test"
			expected := "http://" + target + ":" + strconv.Itoa(port) + restapi

			fullURL := GetHelper().MakeTargetURL(target, port, restapi)

			if expected != fullURL {
				t.Error("expect same, but not same")
			}
		})
		t.Run("GetHelperWithCertificate-MakeTargetURL", func(t *testing.T) {
			target := "testserver.test"
			port := 1234
			restapi := "/api/v1/test"
			expected := "http://" + target + ":" + strconv.Itoa(port) + restapi

			fullURL := GetHelperWithCertificate().MakeTargetURL(target, port, restapi)
			if expected != fullURL {
				t.Error("expect same, but not same")
			}
		})
		t.Run("HTTPS-MakeTargetURL", func(t *testing.T) {
			target := "testserver.test"
			port := 1234
			restapi := "/api/v1/test"
			expected := "https://" + target + ":" + strconv.Itoa(port) + restapi

			var hh helperImpl
			hh.SetCertificateFilePath("fakecerts")

			fullURL := hh.MakeTargetURL(target, port, restapi)
			if expected != fullURL {
				t.Error("expect same, but not same")
			}
		})
	})
}

func TestResponse(t *testing.T) {
	contentsType := "application/json; charset=UTF-8"
	body := "test"

	t.Run("Success", func(t *testing.T) {
		t.Run("WithBody", func(t *testing.T) {
			w := httptest.NewRecorder()
			GetHelper().Response(w, []byte(body), http.StatusOK)

			if w.Code != http.StatusOK {
				t.Error("unexpected code")
			} else if w.Header().Get("Content-Type") != contentsType {
				t.Error("unexpected content type")
			} else if s, _ := ioutil.ReadAll(w.Body); string(s) != body {
				t.Error("unexpected body")
			}
		})
		t.Run("WithoutBody", func(t *testing.T) {
			w := httptest.NewRecorder()
			GetHelper().Response(w, nil, http.StatusOK)

			if w.Code != http.StatusOK {
				t.Error("unexpected code")
			} else if w.Header().Get("Content-Type") != contentsType {
				t.Error("unexpected content type")
			} else if s, _ := ioutil.ReadAll(w.Body); len(s) > 0 {
				t.Error("unexpected body")
			}
		})
	})
}
