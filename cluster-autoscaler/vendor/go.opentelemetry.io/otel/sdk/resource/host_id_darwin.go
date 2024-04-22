// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

<<<<<<<< HEAD:cluster-autoscaler/vendor/go.opentelemetry.io/otel/semconv/v1.17.0/exception.go
package semconv // import "go.opentelemetry.io/otel/semconv/v1.17.0"

const (
	// ExceptionEventName is the name of the Span event representing an exception.
	ExceptionEventName = "exception"
)
========
package resource // import "go.opentelemetry.io/otel/sdk/resource"

var platformHostIDReader hostIDReader = &hostIDReaderDarwin{
	execCommand: execCommand,
}
>>>>>>>> upstream-release-1.29.0:cluster-autoscaler/vendor/go.opentelemetry.io/otel/sdk/resource/host_id_darwin.go
