/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

<<<<<<<< HEAD:cluster-autoscaler/cloudprovider/volcengine/volc-sdk-golang/service/sts/wrapper_test.go
package sts

import (
	"encoding/json"
	"fmt"
	"testing"
)

const (
	testAk = "testAK"
	testSk = "testSK"
)

func TestIAM_AssumeRole(t *testing.T) {
	DefaultInstance.Client.SetAccessKey(testAk)
	DefaultInstance.Client.SetSecretKey(testSk)

	req := &AssumeRoleRequest{
		DurationSeconds: 7200,
		Policy:          "",
		RoleTrn:         "testRoleTrn",
		RoleSessionName: "test",
	}

	list, status, err := DefaultInstance.AssumeRole(req)
	fmt.Println(status, err)
	b, _ := json.Marshal(list)
	fmt.Println(string(b))
========
package authenticatorfactory

import (
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/request/bearertoken"
	"k8s.io/apiserver/pkg/authentication/token/tokenfile"
	"k8s.io/apiserver/pkg/authentication/user"
)

// NewFromTokens returns an authenticator.Request or an error
func NewFromTokens(tokens map[string]*user.DefaultInfo, audiences authenticator.Audiences) authenticator.Request {
	return bearertoken.New(authenticator.WrapAudienceAgnosticToken(audiences, tokenfile.New(tokens)))
>>>>>>>> upstream-release-1.29.0:vertical-pod-autoscaler/e2e/vendor/k8s.io/apiserver/pkg/authentication/authenticatorfactory/loopback.go
}
