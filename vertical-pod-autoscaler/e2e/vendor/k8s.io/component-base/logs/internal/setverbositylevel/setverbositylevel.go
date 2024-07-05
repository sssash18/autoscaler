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

<<<<<<<< HEAD:vertical-pod-autoscaler/e2e/vendor/k8s.io/apiserver/pkg/admission/plugin/validatingadmissionpolicy/internal/generic/informer.go
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
}
========
// Package setverbositylevel stores callbacks that will be invoked by logs.GlogLevel.
//
// This is a separate package to avoid a dependency from
// k8s.io/component-base/logs (uses the callbacks) to
// k8s.io/component-base/logs/api/v1 (adds them). Not all users of the logs
// package also use the API.
package setverbositylevel

import (
	"sync"
)

var (
	// Mutex controls access to the callbacks.
	Mutex sync.Mutex

	Callbacks []func(v uint32) error
)
>>>>>>>> upstream-release-1.30.0:vertical-pod-autoscaler/e2e/vendor/k8s.io/component-base/logs/internal/setverbositylevel/setverbositylevel.go
