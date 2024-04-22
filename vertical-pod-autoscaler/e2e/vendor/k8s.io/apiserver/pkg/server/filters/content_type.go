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

<<<<<<<< HEAD:cluster-autoscaler/estimator/threshold.go
package estimator

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// Threshold provides resources configuration for threshold based estimation limiter.
// Return value of 0 means that no limit is set.
type Threshold interface {
	NodeLimit(cloudprovider.NodeGroup, EstimationContext) int
	DurationLimit(cloudprovider.NodeGroup, EstimationContext) time.Duration
========
package filters

import "net/http"

// WithContentType sets both the Content-Type and the X-Content-Type-Options (nosniff) header
func WithContentType(handler http.Handler, contentType string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		handler.ServeHTTP(w, r)
	})
>>>>>>>> upstream-release-1.29.0:vertical-pod-autoscaler/e2e/vendor/k8s.io/apiserver/pkg/server/filters/content_type.go
}
