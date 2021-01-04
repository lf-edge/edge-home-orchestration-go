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
	"errors"
	"io/ioutil"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
	"net/http"
	"os"

	"github.com/casbin/casbin"
)

// AuthorizerImpl structure
type AuthorizerImpl struct{}

const (
	rbacPolicyFileName = "policy.csv"
	policy_template    = "p, admin, /*, *\n" +
		"p, member, /api/v1/orchestration/services, *\n"
	rbacAuthModelFileName = "auth_model.conf"
	auth_model_template   = "[request_definition]\n" +
		"r = sub, obj, act\n\n" +
		"[policy_definition]\n" +
		"p = sub, obj, act\n\n" +
		"[policy_effect]\n" +
		"e = some(where (p.eft == allow))\n\n" +
		"[matchers]\n" +
		"m = r.sub == p.sub && keyMatch(r.obj, p.obj) && (r.act == p.act || p.act == \"*\")\n"
)

var (
	logPrefix             = "[securemgr: RBAC]"
	log                   = logmgr.GetInstance()
	rbacPolicyFilePath    = ""
	rbacAuthModelFilePath = ""
	authorizerIns         *AuthorizerImpl
	initialized           = false
	users                 Users
	enf                   *casbin.Enforcer
)

func init() {
	authorizerIns = new(AuthorizerImpl)
}

type User struct {
	Name string
	Role string
}

// Users is a list of users
type Users []User

// findByName returns the user with the given name
func (u Users) findByName(name string) (User, error) {
	for _, user := range u {
		if user.Name == name {
			return user, nil
		}
	}
	return User{}, errors.New("User is not found")
}

// Init sets the environments for securemgr
func Init(rbacRulePath string) {
	if _, err := os.Stat(rbacRulePath); err != nil {
		err := os.MkdirAll(rbacRulePath, os.ModePerm)
		if err != nil {
			log.Panicf("%s Failed to create rbacRulePath %s: %s\n", logPrefix, rbacRulePath, err)
			return
		}
	}
	rbacPolicyFilePath = rbacRulePath + "/" + rbacPolicyFileName
	if _, err := os.Stat(rbacPolicyFilePath); err != nil {
		err = ioutil.WriteFile(rbacPolicyFilePath, []byte(policy_template), 0666)
		if err != nil {
			log.Panicf("%s Cannot create %s: %s\n", logPrefix, rbacPolicyFilePath, err)
		}
	}

	rbacAuthModelFilePath = rbacRulePath + "/" + rbacAuthModelFileName
	if _, err := os.Stat(rbacAuthModelFilePath); err != nil {
		err = ioutil.WriteFile(rbacAuthModelFilePath, []byte(auth_model_template), 0666)
		if err != nil {
			log.Panicf("%s Cannot create %s: %s\n", logPrefix, rbacAuthModelFilePath, err)
		}
	}

	users = append(users, User{Name: "Admin", Role: "admin"})
	users = append(users, User{Name: "Member", Role: "member"})

	enf = casbin.NewEnforcer(rbacAuthModelFilePath, rbacPolicyFilePath)

	initialized = true
}

func Authorizer(name string, r *http.Request) error {
	user, err := users.findByName(name)
	if err != nil {
		log.Println(logPrefix, err)
		return err
	}

	role := user.Role
	if role == "" {
		role = "unknow"
	}

	// log.Print("user.Name = ", user.Name)
	// log.Print("user.Role = ", user.Role)
	// log.Print("r.URL.Path = ", r.URL.Path)
	// log.Print("r.Method = ", r.Method)

	// casbin enforce
	res, err := enf.EnforceSafe(role, r.URL.Path, r.Method)
	if err != nil {
		log.Println(logPrefix, "Error: ", err)
		return err
	}
	if res {
		return nil
	} else {
		log.Println(logPrefix, "Unauthorized request")
		return errors.New("Unauthorized request")
	}
}
