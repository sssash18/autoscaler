//go:build !windows
// +build !windows

/*
<<<<<<<< HEAD:cluster-autoscaler/vendor/k8s.io/kubernetes/pkg/kubelet/network/dns/dns_other.go
Copyright 2023 The Kubernetes Authors.
========
Copyright 2017 The Kubernetes Authors.
>>>>>>>> upstream-release-1.29.0:vertical-pod-autoscaler/e2e/vendor/k8s.io/apiserver/pkg/server/signal_posix.go

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

<<<<<<<< HEAD:cluster-autoscaler/vendor/k8s.io/kubernetes/pkg/kubelet/network/dns/dns_other.go
package dns

// Read the DNS configuration from a resolv.conf file.
var getHostDNSConfig = getDNSConfig
========
package server

import (
	"os"
	"syscall"
)

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
>>>>>>>> upstream-release-1.29.0:vertical-pod-autoscaler/e2e/vendor/k8s.io/apiserver/pkg/server/signal_posix.go
