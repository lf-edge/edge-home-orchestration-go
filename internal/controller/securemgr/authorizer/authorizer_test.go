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
package authorizer

import (
	"net/http"
	"os"
	"testing"
)

const (
	fakerbacPath              = "fakerbac"
	fakerbacPolicyFilePath    = fakerbacPath + "/" + rbacPolicyFileName
	fakerbacAuthModelFilePath = fakerbacPath + "/" + rbacAuthModelFileName
	unexpectedSuccess         = "unexpected success"
	unexpectedFail            = "unexpected fail"
)

func TestInit(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error(r)
			}
			os.RemoveAll(fakerbacPath)
		}()

		Init(fakerbacPath)
		if _, err := os.Stat(fakerbacPath); err != nil {
			t.Error(err.Error())
		}
		if _, err := os.Stat(fakerbacPolicyFilePath); os.IsNotExist(err) {
			t.Error(err.Error())
		}
		if _, err := os.Stat(fakerbacAuthModelFilePath); os.IsNotExist(err) {
			t.Error(err.Error())
		}
		os.RemoveAll(fakerbacPath)

		if _, err := os.Stat(fakerbacPath); err != nil {
			err := os.MkdirAll(fakerbacPath, 0444)
			if err != nil {
				t.Error(err.Error())
			}
		}
		Init(fakerbacPath)
		if _, err := os.Stat(rbacPolicyFileName); os.IsNotExist(err) {
			t.Error(err.Error())
		}

		if _, err := os.Stat(rbacAuthModelFileName); os.IsNotExist(err) {
			t.Error(err.Error())
		}

	})

	t.Run("Error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error(r)
			}
			os.RemoveAll(fakerbacPath)
		}()
		Init("/fakerbacPath")

		if _, err := os.Stat("/fakerbacPath"); os.IsNotExist(err) {
			t.Error(err.Error())
		}
	})
}

func TestFindByName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if _, err := users.findByName("Admin"); err != nil {
			t.Error(unexpectedFail)
		}
	})
	t.Run("Error", func(t *testing.T) {
		if _, err := users.findByName("admin"); err == nil {
			t.Error(unexpectedSuccess)
		}
	})

}

func TestAuthorizer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/api/v1/orchestration/securemgr", nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := Authorizer("Admin", req); err != nil {
			t.Error(unexpectedFail)
		}
	})
	t.Run("Error", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/api/v1/orchestration/securemgr", nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := Authorizer("admin", req); err == nil {
			t.Error(unexpectedSuccess)
		}

		if err := Authorizer("member", req); err == nil {
			t.Error(unexpectedSuccess)
		}

		users = append(users, User{Name: "Member1", Role: ""})
		if err := Authorizer("Member1", req); err == nil {
			t.Error(unexpectedSuccess)
		}
	})
}

func FuzzTestFindByName(f *testing.F) {
	testcases := []string{"Admin", "!12345"}
	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}

	users = append(users, User{Name: "Admin", Role: "admin"})
	users = append(users, User{Name: "Member", Role: "member"})

	f.Fuzz(func(t *testing.T, orig string) {

		if _, err := users.findByName(orig); err != nil {
			return
		}
		if orig == "Admin" || orig == "Member" {
			return
		}
		t.Error(unexpectedSuccess)
	})
}

func FuzzTestAuthorizer(f *testing.F) {

	defer func() {
		os.RemoveAll(fakerbacPath)
	}()

	testcases := []string{"Admin", "Member", "!12345"}
	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}

	Init(fakerbacPath)

	req, err := http.NewRequest("POST", "/api/v1/orchestration/securemgr", nil)
	if err != nil {
		return
	}

	f.Fuzz(func(t *testing.T, orig string) {

		if err := Authorizer(orig, req); err != nil {
			return
		}

		if orig == "Admin" || orig == "Member" {
			return
		}
		t.Error(unexpectedSuccess)
	})
}
