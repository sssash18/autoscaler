/*
Copyright 2024 The Kubernetes Authors.

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

package provreq

import (
	"k8s.io/autoscaler/cluster-autoscaler/observers/loopstart"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/klog/v2"
)

// ProvisioningRequestProcessor process ProvisioningRequests in the cluster.
type ProvisioningRequestProcessor interface {
	Process([]*provreqwrapper.ProvisioningRequest)
	CleanUp()
}

// CombinedProvReqProcessor is responsible for processing ProvisioningRequest for each ProvisioningClass
// every CA loop and updating conditions for expired ProvisioningRequests.
type CombinedProvReqProcessor struct {
	client     *provreqclient.ProvisioningRequestClient
	processors []ProvisioningRequestProcessor
}

// NewCombinedProvReqProcessor return new CombinedProvReqProcessor.
func NewCombinedProvReqProcessor(client *provreqclient.ProvisioningRequestClient, processors []ProvisioningRequestProcessor) loopstart.Observer {
	return &CombinedProvReqProcessor{client: client, processors: processors}
}

// Refresh iterates over ProvisioningRequests and updates its conditions/state.
func (cp *CombinedProvReqProcessor) Refresh() {
	provReqs, err := cp.client.ProvisioningRequests()
	if err != nil {
		klog.Errorf("Failed to get ProvisioningRequests list, err: %v", err)
		return
	}
	for _, p := range cp.processors {
		p.Process(provReqs)
	}
}

// CleanUp cleans up internal state
func (cp *CombinedProvReqProcessor) CleanUp() {}
