package virtual

import (
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
	"os"
	"reflect"
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

const GPULabel = "virtual/gpu"

type VirtualCloudProvider struct {
	clusterInfo     *clusterInfo
	resourceLimiter *cloudprovider.ResourceLimiter
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

type VirtualNodeTemplate struct {
	Name         string
	CPU          string
	GPU          string
	Memory       string
	InstanceType string
	Region       string
	Zone         string
	Tags         map[string]string
}

type clusterInfo struct {
	nodeTemplates map[string]VirtualNodeTemplate
	nodeGroups    map[string]VirtualNodeGroup
	workerPools   []VirtualWorkerPool
}

func BuildVirtual(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {

	if opts.CloudProviderName != "virtual" {
		return nil
	}

	clusterInfoPath := os.Getenv("GARDENER_CLUSTER_INFO")
	clusterHistoryPath := os.Getenv("GARDENER_CLUSTER_HISTORY")
	shootName := os.Getenv("SHOOT_NAME")
	if clusterHistoryPath == "" && clusterInfoPath == "" {
		klog.Fatalf("cannot create virtual provider one of env var GARDENER_CLUSTER_INFO or GARDENER_CLUSTER_HISTORY needs to be set.")
		return nil
	}

	if clusterInfoPath != "" && shootName != "" {
		clusterInfo, err := readInitClusterInfo(clusterInfoPath, shootName)
		if err != nil {
			klog.Fatalf("cannot build the virtual cloud provider: %s", err.Error())
		}
		return &VirtualCloudProvider{
			clusterInfo:     &clusterInfo,
			resourceLimiter: rl,
		}
	}

	return nil

}

func getIntOrString(val interface{}) intstr.IntOrString {
	var valIntOrString intstr.IntOrString
	if reflect.TypeOf(val) == reflect.TypeOf("") {
		valIntOrString = intstr.FromString(val.(string))
	} else {
		if reflect.TypeOf(val) == reflect.TypeOf(0.0) {
			valIntOrString = intstr.FromInt32(int32(val.(float64)))
		} else {
			valIntOrString = intstr.FromInt32(int32(val.(int)))
		}
	}
	return valIntOrString
}

func getVirtualWorkerPoolFromWorker(worker map[string]interface{}) VirtualWorkerPool {
	architecture := worker["machine"].(map[string]interface{})["architecture"].(string)
	machineType := worker["machine"].(map[string]interface{})["type"].(string)
	minimum := worker["minimum"].(float64)
	maximum := worker["maximum"].(float64)
	maxSurge := worker["maxSurge"]
	maxUnavailable := worker["maxUnavailable"].(float64)
	zones := worker["zones"].([]interface{})
	var zonesStr []string
	for _, zone := range zones {
		zonesStr = append(zonesStr, zone.(string))
	}
	return VirtualWorkerPool{
		MachineType:    machineType,
		Architecture:   architecture,
		Minimum:        int(minimum),
		Maximum:        int(maximum),
		MaxSurge:       getIntOrString(maxSurge),
		MaxUnavailable: getIntOrString(int(maxUnavailable)),
		Zones:          zonesStr,
	}
}

func getVirtualWorkerPoolsFromShoot(shootDataMap map[string]interface{}) (virtualWorkerPools []VirtualWorkerPool) {
	workers := ((shootDataMap["spec"].(map[string]interface{})["provider"]).(map[string]interface{}))["workers"].([]interface{})
	for _, worker := range workers {
		virtualWorkerPools = append(virtualWorkerPools, getVirtualWorkerPoolFromWorker(worker.(map[string]interface{})))
	}
	return
}

func getVirtualNodeTemplateFromMCC(mcc map[string]interface{}) VirtualNodeTemplate {
	nodeTemplate := mcc["nodeTemplate"].(map[string]interface{})
	providerSpec := mcc["providerSpec"].(map[string]interface{})
	metadata := mcc["metadata"].(map[string]interface{})
	cpu := nodeTemplate["capacity"].(map[string]interface{})["cpu"].(string)
	gpu := nodeTemplate["capacity"].(map[string]interface{})["gpu"].(string)
	memory := nodeTemplate["capacity"].(map[string]interface{})["memory"].(string)
	instanceType := nodeTemplate["instanceType"].(string)
	region := nodeTemplate["region"].(string)
	zone := nodeTemplate["zone"].(string)
	tags := providerSpec["tags"].(map[string]interface{})
	name := metadata["name"].(string)

	tagsStrMap := make(map[string]string)
	for tagKey, tagVal := range tags {
		tagsStrMap[tagKey] = tagVal.(string)
	}
	return VirtualNodeTemplate{
		Name:         name,
		CPU:          cpu,
		GPU:          gpu,
		Memory:       memory,
		InstanceType: instanceType,
		Region:       region,
		Zone:         zone,
		Tags:         tagsStrMap,
	}
}

func getVirtualNodeTemplatesFromMCC(mccData map[string]interface{}) map[string]VirtualNodeTemplate {
	mccList := mccData["items"].([]interface{})
	var nodeTemplates []VirtualNodeTemplate
	for _, mcc := range mccList {
		nodeTemplates = append(nodeTemplates, getVirtualNodeTemplateFromMCC(mcc.(map[string]interface{})))
	}
	return lo.KeyBy(nodeTemplates, func(item VirtualNodeTemplate) string {
		return item.Name
	})
}

func getVirtualNodeGroupFromMCD(mcd map[string]interface{}) VirtualNodeGroup {
	specMap := mcd["spec"].(map[string]interface{})
	metadataMap := mcd["metadata"].(map[string]interface{})
	replicas := specMap["replicas"].(float64)
	name := metadataMap["name"].(string)
	poolName := specMap["template"].(map[string]interface{})["spec"].(map[string]interface{})["nodeTemplate"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["worker.gardener.cloud/pool"].(string)
	//TODO zone has different labels for each provider
	//zone := specMap["template"].(map[string]interface{})["spec"].(map[string]interface{})["nodeTemplate"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["worker.gardener.cloud/zone"].(string)
	return VirtualNodeGroup{
		Name:       name,
		poolName:   poolName,
		zone:       "zone",
		targetSize: int(replicas),
		minSize:    0,
		maxSize:    0,
	}
}

func getVirtualNodeGroupsFromMCD(mcdData map[string]interface{}) map[string]VirtualNodeGroup {
	mcdList := mcdData["items"].([]interface{})
	var nodeGroups []VirtualNodeGroup
	for _, mcd := range mcdList {
		nodeGroups = append(nodeGroups, getVirtualNodeGroupFromMCD(mcd.(map[string]interface{})))
	}
	return lo.KeyBy(nodeGroups, func(item VirtualNodeGroup) string {
		return item.Name
	})
}

func readInitClusterInfo(clusterInfoPath string, shootName string) (cI clusterInfo, err error) {
	shootJsonFile := fmt.Sprintf("%s/%s.json", clusterInfoPath, shootName)
	data, err := os.ReadFile(shootJsonFile)
	if err != nil {
		klog.Errorf("cannot read the shoot json file: %s", err.Error())
		return
	}

	var shootDataMap map[string]interface{}
	err = json.Unmarshal(data, &shootDataMap)
	if err != nil {
		klog.Errorf("cannot unmarshal the shoot json: %s", err.Error())
		return
	}
	workerPools := getVirtualWorkerPoolsFromShoot(shootDataMap)
	cI.workerPools = workerPools

	mccJsonFile := fmt.Sprintf("%s/%s-mcc.json", clusterInfoPath, shootName)
	data, err = os.ReadFile(mccJsonFile)
	if err != nil {
		klog.Errorf("cannot read the mcc json file: %s", err.Error())
		return
	}
	var mccData map[string]interface{}
	err = json.Unmarshal(data, &mccData)
	if err != nil {
		klog.Errorf("cannot unmarshal the mcc json: %s", err.Error())
		return
	}
	cI.nodeTemplates = getVirtualNodeTemplatesFromMCC(mccData)

	mcdJsonFile := fmt.Sprintf("%s/%s-mcd.json", clusterInfoPath, shootName)
	data, err = os.ReadFile(mcdJsonFile)
	if err != nil {
		klog.Errorf("cannot read the mcd json file: %s", err.Error())
		return
	}
	var mcdData map[string]interface{}
	err = json.Unmarshal(data, &mcdData)
	if err != nil {
		klog.Errorf("cannot unmarshal the mcd json: %s", err.Error())
		return
	}
	cI.nodeGroups = getVirtualNodeGroupsFromMCD(mcdData)
	return
}

func (v *VirtualCloudProvider) Name() string {
	return cloudprovider.VirtualProviderName
}

func (v *VirtualCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, len(v.clusterInfo.nodeGroups))
	for _, ng := range v.clusterInfo.nodeGroups {
		result = append(result, &ng)
	}
	return result
}

func (v *VirtualCloudProvider) NodeGroupForNode(node *corev1.Node) (cloudprovider.NodeGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (v *VirtualCloudProvider) HasInstance(node *corev1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

func (v *VirtualCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

func (v *VirtualCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

func (v *VirtualCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []corev1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (v *VirtualCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return v.resourceLimiter, nil
}

func (v *VirtualCloudProvider) GPULabel() string {
	return GPULabel
}

func (v *VirtualCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

func (v *VirtualCloudProvider) GetNodeGpuConfig(node *corev1.Node) *cloudprovider.GpuConfig {
	return nil
}

func (v *VirtualCloudProvider) Cleanup() error {
	return nil
}

func (v *VirtualCloudProvider) Refresh() error {
	return nil
}

var _ cloudprovider.CloudProvider = (*VirtualCloudProvider)(nil)

func (v *VirtualNodeGroup) MaxSize() int {
	return v.maxSize
}

func (v *VirtualNodeGroup) MinSize() int {
	return v.minSize
}

func (v *VirtualNodeGroup) TargetSize() (int, error) {
	return v.targetSize, nil
}

func (v *VirtualNodeGroup) IncreaseSize(delta int) error {
	//TODO implement me
	panic("implement me")
}

func (v *VirtualNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

func (v *VirtualNodeGroup) DeleteNodes(nodes []*corev1.Node) error {
	//TODO implement me
	panic("implement me")
}

func (v *VirtualNodeGroup) DecreaseTargetSize(delta int) error {
	//TODO implement me
	panic("implement me")
}

func (v *VirtualNodeGroup) Id() string {
	return v.Name
}

func (v *VirtualNodeGroup) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", v.Id(), v.MinSize(), v.MaxSize())
}

func (v *VirtualNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	//TODO implement me
	panic("implement me")
}

func (v *VirtualNodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (v *VirtualNodeGroup) Exist() bool {
	//TODO implement me
	panic("implement me")
}

func (v *VirtualNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (v *VirtualNodeGroup) Delete() error {
	//TODO implement me
	panic("implement me")
}

func (v *VirtualNodeGroup) Autoprovisioned() bool {
	//TODO implement me
	panic("implement me")
}

func (v *VirtualNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	//TODO implement me
	panic("implement me")
}
