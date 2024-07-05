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

package universal

type HttpMethod int

const (
<<<<<<<< HEAD:cluster-autoscaler/vendor/github.com/vmware/govmomi/internal/version/version.go
	// ClientName is the name of this SDK
	ClientName = "govmomi"

	// ClientVersion is the version of this SDK
	ClientVersion = "0.30.6"
========
	GET HttpMethod = iota
	HEAD
	POST
	PUT
	DELETE
)

type ContentType int

const (
	Default ContentType = iota
	FormUrlencoded
	ApplicationJSON
>>>>>>>> upstream-release-1.30.0:cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/universal/universal_const.go
)
