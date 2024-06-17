package virtual

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
	"os"
)

type VirtualNodeGroup struct {
	Name       string
	poolName   string
	zone       string
	targetSize int
	minSize    int
	maxSize    int
}

var _ cloudprovider.NodeGroup = (*VirtualNodeGroup)(nil)

type VirtualCloudProvider struct {
	clusterInfo *clusterInfo
}

type VirtualWorkerPool struct {
	MachineType    string
	Architecture   string
	Minimum        int
	Maximum        int
	MaxSurge       intstr.IntOrString
	MaxUnavailable intstr.IntOrString
	Zones          []string
}

type VirtualMachineDeployment struct {
}

type clusterInfo struct {
	nodeTemplates map[string]corev1.Node
	nodeGroups    map[string]VirtualNodeGroup
	workerPools   map[string]VirtualWorkerPool
}

func BuildVirtual(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {

	if opts.CloudProviderName != "virtual" {
		return nil
	}

	clusterInfoPath := os.Getenv("GARDENER_CLUSTER_INFO")
	clusterHistoryPath := os.Getenv("GARDENER_CLUSTER_HISTORY")
	if clusterHistoryPath == "" && clusterInfoPath == "" {
		klog.Fatalf("cannot create virtual provider one of env var GARDENER_CLUSTER_INFO or GARDENER_CLUSTER_HISTORY needs to be set.")
		return nil
	}

	if clusterInfoPath != "" {
		readInitClusterInfo(clusterInfoPath)
	}
	return nil

}

func readInitClusterInfo(clusterInfoPath string) clusterInfo {

}

func (v VirtualCloudProvider) Name() string {
	//TODO implement me
	panic("implement me")
}

func (v VirtualCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	//TODO implement me
	panic("implement me")
}

func (v VirtualCloudProvider) NodeGroupForNode(node *corev1.Node) (cloudprovider.NodeGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualCloudProvider) HasInstance(node *corev1.Node) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []corev1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualCloudProvider) GPULabel() string {
	//TODO implement me
	panic("implement me")
}

func (v VirtualCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	//TODO implement me
	panic("implement me")
}

func (v VirtualCloudProvider) GetNodeGpuConfig(node *corev1.Node) *cloudprovider.GpuConfig {
	//TODO implement me
	panic("implement me")
}

func (v VirtualCloudProvider) Cleanup() error {
	//TODO implement me
	panic("implement me")
}

func (v VirtualCloudProvider) Refresh() error {
	//TODO implement me
	panic("implement me")
}

var _ cloudprovider.CloudProvider = (*VirtualCloudProvider)(nil)

func (v VirtualNodeGroup) MaxSize() int {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) MinSize() int {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) TargetSize() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) IncreaseSize(delta int) error {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) AtomicIncreaseSize(delta int) error {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) DeleteNodes(nodes []*corev1.Node) error {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) DecreaseTargetSize(delta int) error {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) Id() string {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) Debug() string {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) Exist() bool {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) Delete() error {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) Autoprovisioned() bool {
	//TODO implement me
	panic("implement me")
}

func (v VirtualNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	//TODO implement me
	panic("implement me")
}
