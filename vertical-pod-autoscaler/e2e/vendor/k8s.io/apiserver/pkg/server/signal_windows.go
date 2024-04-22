/*
Copyright 2020 The Kubernetes Authors.

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

<<<<<<<< HEAD:cluster-autoscaler/vendor/k8s.io/cloud-provider/controllers/service/config/doc.go
// +k8s:deepcopy-gen=package

package config // import "k8s.io/cloud-provider/controllers/service/config"
========
package server

import (
	"os"
)

var shutdownSignals = []os.Signal{os.Interrupt}
>>>>>>>> upstream-release-1.29.0:vertical-pod-autoscaler/e2e/vendor/k8s.io/apiserver/pkg/server/signal_windows.go
