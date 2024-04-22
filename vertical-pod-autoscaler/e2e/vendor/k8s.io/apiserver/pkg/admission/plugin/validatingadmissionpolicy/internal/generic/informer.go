//go:build !linux && !windows

/*
Copyright 2022 The Kubernetes Authors.

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

<<<<<<<< HEAD:cluster-autoscaler/vendor/k8s.io/kubernetes/pkg/kubelet/kubelet_server_journal_others.go
package kubelet

import (
	"context"
	"errors"
)

// getLoggingCmd on unsupported operating systems returns the echo command and a warning message (as strings)
func getLoggingCmd(n *nodeLogQuery, services []string) (string, []string, error) {
	return "", []string{}, errors.New("Operating System Not Supported")
}

// checkForNativeLogger on unsupported operating systems returns false
func checkForNativeLogger(ctx context.Context, service string) bool {
	return false
========
package generic

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

var _ Informer[runtime.Object] = informer[runtime.Object]{}

type informer[T runtime.Object] struct {
	cache.SharedIndexInformer
	lister[T]
}

func NewInformer[T runtime.Object](informe cache.SharedIndexInformer) Informer[T] {
	return informer[T]{
		SharedIndexInformer: informe,
		lister:              NewLister[T](informe.GetIndexer()),
	}
>>>>>>>> upstream-release-1.29.0:vertical-pod-autoscaler/e2e/vendor/k8s.io/apiserver/pkg/admission/plugin/validatingadmissionpolicy/internal/generic/informer.go
}
