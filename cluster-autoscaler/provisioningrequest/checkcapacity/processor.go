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

package checkcapacity

import (
	"time"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/klog/v2"
)

const (
	defaultReservationTime = 10 * time.Minute
	defaultExpirationTime  = 7 * 24 * time.Hour // 7 days
	// defaultMaxUpdated is a limit for ProvisioningRequest to update conditions in one ClusterAutoscaler loop.
	defaultMaxUpdated = 20
)

type checkCapacityProcessor struct {
	now        func() time.Time
	maxUpdated int
	client     *provreqclient.ProvisioningRequestClient
}

// NewCheckCapacityProcessor return ProvisioningRequestProcessor for Check-capacity ProvisioningClass.
func NewCheckCapacityProcessor(client *provreqclient.ProvisioningRequestClient) *checkCapacityProcessor {
	return &checkCapacityProcessor{now: time.Now, maxUpdated: defaultMaxUpdated, client: client}
}

// Process iterates over ProvisioningRequests and apply:
// -BookingExpired condition for Provisioned ProvisioningRequest if capacity reservation time is expired.
// -Failed condition for ProvisioningRequest that were not provisioned during defaultExpirationTime.
// TODO(yaroslava): fetch reservation and expiration time from ProvisioningRequest
func (p *checkCapacityProcessor) Process(provReqs []*provreqwrapper.ProvisioningRequest) {
	expiredProvReq := []*provreqwrapper.ProvisioningRequest{}
	failedProvReq := []*provreqwrapper.ProvisioningRequest{}
	for _, provReq := range provReqs {
		if len(expiredProvReq) >= p.maxUpdated {
			break
		}
		conditions := provReq.Status.Conditions
		if provReq.Spec.ProvisioningClassName != v1beta1.ProvisioningClassCheckCapacity ||
			apimeta.IsStatusConditionTrue(conditions, v1beta1.BookingExpired) || apimeta.IsStatusConditionTrue(conditions, v1beta1.Failed) {
			continue
		}
		provisioned := apimeta.FindStatusCondition(conditions, v1beta1.Provisioned)
		if provisioned != nil && provisioned.Status == metav1.ConditionTrue {
			if provisioned.LastTransitionTime.Add(defaultReservationTime).Before(p.now()) {
				expiredProvReq = append(expiredProvReq, provReq)
			}
		} else if len(failedProvReq) < p.maxUpdated-len(expiredProvReq) {
			created := provReq.CreationTimestamp
			if created.Add(defaultExpirationTime).Before(p.now()) {
				failedProvReq = append(failedProvReq, provReq)
			}
		}
	}
	for _, provReq := range expiredProvReq {
		conditions.AddOrUpdateCondition(provReq, v1beta1.BookingExpired, metav1.ConditionTrue, conditions.CapacityReservationTimeExpiredReason, conditions.CapacityReservationTimeExpiredMsg, metav1.NewTime(p.now()))
		_, updErr := p.client.UpdateProvisioningRequest(provReq.ProvisioningRequest)
		if updErr != nil {
			klog.Errorf("failed to add BookingExpired condition to ProvReq %s/%s, err: %v", provReq.Namespace, provReq.Name, updErr)
			continue
		}
	}
	for _, provReq := range failedProvReq {
		conditions.AddOrUpdateCondition(provReq, v1beta1.Failed, metav1.ConditionTrue, conditions.ExpiredReason, conditions.ExpiredMsg, metav1.NewTime(p.now()))
		_, updErr := p.client.UpdateProvisioningRequest(provReq.ProvisioningRequest)
		if updErr != nil {
			klog.Errorf("failed to add Failed condition to ProvReq %s/%s, err: %v", provReq.Namespace, provReq.Name, updErr)
			continue
		}
	}
}

// Cleanup cleans up internal state.
func (p *checkCapacityProcessor) CleanUp() {}
