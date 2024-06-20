package virtual

import (
	"fmt"
	gct "github.com/elankath/gardener-cluster-types"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	kubeletapis "k8s.io/kubelet/pkg/apis"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type VirtualNodeGroup struct {
	gct.NodeGroupInfo
	nodeTemplate gct.NodeTemplate
	instances    []cloudprovider.Instance
	clientSet    *kubernetes.Clientset
}

var _ cloudprovider.NodeGroup = (*VirtualNodeGroup)(nil)

const GPULabel = "virtual/gpu"

type VirtualCloudProvider struct {
	clusterInfo       *gct.ClusterInfo
	resourceLimiter   *cloudprovider.ResourceLimiter
	clientSet         *kubernetes.Clientset
	virtualNodeGroups map[string]*VirtualNodeGroup
}

func BuildVirtual(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {

	if opts.CloudProviderName != "virtual" {
		return nil
	}

	kubeConfigPath := opts.KubeClientOpts.KubeConfigPath

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		klog.Fatalf("cannot build config from kubeConfigPath: %s, error: %s", kubeConfigPath, err.Error())
	}

	config.Burst = opts.KubeClientOpts.KubeClientBurst
	config.QPS = opts.KubeClientOpts.KubeClientQPS
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("failed to create the client: %s", err.Error())
	}

	clusterInfoPath := os.Getenv("GARDENER_CLUSTER_INFO")
	clusterHistoryPath := os.Getenv("GARDENER_CLUSTER_HISTORY")
	if clusterHistoryPath == "" && clusterInfoPath == "" {
		klog.Fatalf("cannot create virtual provider one of env var GARDENER_CLUSTER_INFO or GARDENER_CLUSTER_HISTORY needs to be set.")
		return nil
	}

	if clusterInfoPath != "" {
		clusterInfo, err := readInitClusterInfo(clusterInfoPath)
		if err != nil {
			klog.Fatalf("cannot build the virtual cloud provider: %s", err.Error())
		}
		virtualNodeGroups := make(map[string]*VirtualNodeGroup)
		for name, ng := range clusterInfo.NodeGroups {
			virtualNodeGroup := VirtualNodeGroup{
				NodeGroupInfo: ng,
				nodeTemplate:  gct.NodeTemplate{},
				instances:     nil,
				clientSet:     clientSet,
			}
			//populateNodeTemplateTaints(nodeTemplates,mcdData)
			virtualNodeGroups[name] = &virtualNodeGroup
		}
		err = populateNodeTemplates(virtualNodeGroups, clusterInfo.NodeTemplates)
		if err != nil {
			klog.Fatalf("cannot construct the virtual cloud provider: %s", err.Error())
		}
		return &VirtualCloudProvider{
			virtualNodeGroups: virtualNodeGroups,
			clusterInfo:       &clusterInfo,
			resourceLimiter:   rl,
			clientSet:         clientSet,
		}
	}

	return nil

}

func getIntOrString(val any) intstr.IntOrString {
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

func getWorkerPoolInfo(workerPool map[string]any) (wP gct.WorkerPoolInfo, err error) {
	architecture := workerPool["architecture"].(string)
	machineType := workerPool["machineType"].(string)
	minimum := workerPool["minimum"].(int64)
	maximum := workerPool["maximum"].(int64)
	maxSurge, err := gct.AsIntOrString(workerPool["maxSurge"])
	if err != nil {
		return
	}
	maxUnavailable, err := gct.AsIntOrString(workerPool["maxUnavailable"])
	if err != nil {
		return
	}
	zonesVal := workerPool["zones"].([]any)
	var zones []string
	for _, zone := range zonesVal {
		zones = append(zones, zone.(string))
	}
	wP = gct.WorkerPoolInfo{
		MachineType:    machineType,
		Architecture:   architecture,
		Minimum:        int(minimum),
		Maximum:        int(maximum),
		MaxSurge:       maxSurge,
		MaxUnavailable: maxUnavailable,
		Zones:          zones,
	}
	return
}

func getWorkerPoolsFromShootWorker(workerDataMap map[string]any) (virtualWorkerPools []gct.WorkerPoolInfo, err error) {
	workerPools := workerDataMap["spec"].(map[string]any)["pools"].([]any)
	for _, pool := range workerPools {
		var wp gct.WorkerPoolInfo
		wp, err = getWorkerPoolInfo(pool.(map[string]any))
		if err != nil {
			return
		}
		virtualWorkerPools = append(virtualWorkerPools, wp)
	}
	return
}

func getVirtualNodeTemplateFromMCC(mcc map[string]any) (nt gct.NodeTemplate, err error) {
	nodeTemplate := mcc["nodeTemplate"].(map[string]any)
	providerSpec := mcc["providerSpec"].(map[string]any)
	metadata := mcc["metadata"].(map[string]any)
	cpuVal := nodeTemplate["capacity"].(map[string]any)["cpu"].(string)
	gpuVal := nodeTemplate["capacity"].(map[string]any)["gpu"].(string)
	memoryVal := nodeTemplate["capacity"].(map[string]any)["memory"].(string)
	instanceType := nodeTemplate["instanceType"].(string)
	region := nodeTemplate["region"].(string)
	zone := nodeTemplate["zone"].(string)
	tags := providerSpec["tags"].(map[string]any)
	name := metadata["name"].(string)

	tagsStrMap := make(map[string]string)
	for tagKey, tagVal := range tags {
		tagsStrMap[tagKey] = tagVal.(string)
	}

	cpu, err := resource.ParseQuantity(cpuVal)
	if err != nil {
		return
	}
	gpu, err := resource.ParseQuantity(gpuVal)
	if err != nil {
		return
	}
	memory, err := resource.ParseQuantity(memoryVal)
	if err != nil {
		return
	}

	nt = gct.NodeTemplate{
		Name:         name,
		CPU:          cpu,
		GPU:          gpu,
		Memory:       memory,
		InstanceType: instanceType,
		Region:       region,
		Zone:         zone,
		Tags:         tagsStrMap,
	}
	return
}

func getNodeTemplatesFromMCC(mccData map[string]any) (map[string]gct.NodeTemplate, error) {
	mccList := mccData["items"].([]any)
	var nodeTemplates []gct.NodeTemplate
	for _, mcc := range mccList {
		nodeTemplate, err := getVirtualNodeTemplateFromMCC(mcc.(map[string]any))
		if err != nil {
			return nil, fmt.Errorf("cannot build nodeTemplate: %w", err)
		}
		nodeTemplates = append(nodeTemplates, nodeTemplate)
	}
	namespace := mccList[0].(map[string]any)["metadata"].(map[string]any)["namespace"].(string)
	return lo.KeyBy(nodeTemplates, func(item gct.NodeTemplate) string {
		name := item.Name
		idx := strings.LastIndex(name, "-")
		// mcc name - shoot--i585976--suyash-local-worker-1-z1-0af3f , we omit the hash from the mcc name to match it with the nodegroup name
		trimmedName := name[0:idx]
		return fmt.Sprintf("%s.%s", namespace, trimmedName)
	}), nil
}

func getVirtualNodeGroupFromMCD(mcd map[string]any) gct.NodeGroupInfo {
	specMap := mcd["spec"].(map[string]any)
	metadataMap := mcd["metadata"].(map[string]any)
	replicasVal, ok := specMap["replicas"]
	var replicas int
	if !ok {
		replicas = 0
	} else {
		replicas = int(replicasVal.(float64))
	}
	name := metadataMap["name"].(string)
	namespace := metadataMap["namespace"].(string)
	poolName := specMap["template"].(map[string]any)["spec"].(map[string]any)["nodeTemplate"].(map[string]any)["metadata"].(map[string]any)["labels"].(map[string]any)["worker.gardener.cloud/pool"].(string)
	zone := gct.GetZone(specMap["template"].(map[string]any)["spec"].(map[string]any)["nodeTemplate"].(map[string]any)["metadata"].(map[string]any)["labels"].(map[string]any))
	return gct.NodeGroupInfo{
		Name:       fmt.Sprintf("%s.%s", namespace, name),
		PoolName:   poolName,
		Zone:       zone,
		TargetSize: replicas,
		MinSize:    0,
		MaxSize:    0,
	}
}

func getNodeGroupsFromMCD(mcdData map[string]any) map[string]gct.NodeGroupInfo {
	mcdList := mcdData["items"].([]any)
	var nodeGroups []gct.NodeGroupInfo
	for _, mcd := range mcdList {
		nodeGroups = append(nodeGroups, getVirtualNodeGroupFromMCD(mcd.(map[string]any)))
	}
	return lo.KeyBy(nodeGroups, func(item gct.NodeGroupInfo) string {
		return item.Name
	})
}

func parseCASettingsInfo(caDeploymentData map[string]any) (caSettings gct.CASettingsInfo, err error) {
	caSettings.NodeGroupsMinMax = make(map[string]gct.NameMinMax)
	containersVal, err := gct.GetInnerMapValue(caDeploymentData, "spec", "template", "spec", "containers")
	if err != nil {
		return
	}
	containers := containersVal.([]any)
	if len(containers) == 0 {
		err = fmt.Errorf("len of containers is zero, no CA container found")
		return
	}
	caContainer := containers[0].(map[string]any)
	caCommands := caContainer["command"].([]any)
	for _, commandVal := range caCommands {
		command := commandVal.(string)
		vals := strings.Split(command, "=")
		if len(vals) <= 1 {
			continue
		}
		key := vals[0]
		val := vals[1]
		switch key {
		case "--max-graceful-termination-sec":
			caSettings.MaxGracefulTerminationSeconds, err = strconv.Atoi(val)
		case "--max-node-provision-time":
			caSettings.MaxNodeProvisionTime, err = time.ParseDuration(val)
		case "--scan-interval":
			caSettings.ScanInterval, err = time.ParseDuration(val)
		case "--max-empty-bulk-delete":
			caSettings.MaxEmptyBulkDelete, err = strconv.Atoi(val)
		case "--new-pod-scale-up-delay":
			caSettings.NewPodScaleUpDelay, err = time.ParseDuration(val)
		case "--nodes":
			var ngMinMax gct.NameMinMax
			ngVals := strings.Split(val, ":")
			ngMinMax.Min, err = strconv.Atoi(ngVals[0])
			ngMinMax.Max, err = strconv.Atoi(ngVals[1])
			ngMinMax.Name = ngVals[2]
			caSettings.NodeGroupsMinMax[ngMinMax.Name] = ngMinMax
		}
		if err != nil {
			return
		}
	}
	return
}

func readInitClusterInfo(clusterInfoPath string) (cI gct.ClusterInfo, err error) {
	workerJsonFile := fmt.Sprintf("%s/shoot-worker.json", clusterInfoPath)
	data, err := os.ReadFile(workerJsonFile)
	if err != nil {
		klog.Errorf("cannot read the shoot json file: %s", err.Error())
		return
	}

	var workerDataMap map[string]any
	err = json.Unmarshal(data, &workerDataMap)
	if err != nil {
		klog.Errorf("cannot unmarshal the worker json: %s", err.Error())
		return
	}
	workerPools, err := getWorkerPoolsFromShootWorker(workerDataMap)
	if err != nil {
		klog.Errorf("cannot parse the worker pools: %s", err.Error())
		return
	}
	cI.WorkerPools = workerPools

	mccJsonFile := fmt.Sprintf("%s/mcc.json", clusterInfoPath)
	data, err = os.ReadFile(mccJsonFile)
	if err != nil {
		klog.Errorf("cannot read the mcc json file: %s", err.Error())
		return
	}
	var mccData map[string]any
	err = json.Unmarshal(data, &mccData)
	if err != nil {
		klog.Errorf("cannot unmarshal the mcc json: %s", err.Error())
		return
	}
	cI.NodeTemplates, err = getNodeTemplatesFromMCC(mccData)
	if err != nil {
		klog.Errorf("cannot build the nodeTemplates: %s", err.Error())
	}

	mcdJsonFile := fmt.Sprintf("%s/mcd.json", clusterInfoPath)
	data, err = os.ReadFile(mcdJsonFile)
	if err != nil {
		klog.Errorf("cannot read the mcd json file: %s", err.Error())
		return
	}
	var mcdData map[string]any
	err = json.Unmarshal(data, &mcdData)
	if err != nil {
		klog.Errorf("cannot unmarshal the mcd json: %s", err.Error())
		return
	}

	cI.NodeGroups = getNodeGroupsFromMCD(mcdData)

	caDeploymentJsonFile := fmt.Sprintf("%s/ca-deployment.json", clusterInfoPath)
	data, err = os.ReadFile(caDeploymentJsonFile)
	if err != nil {
		klog.Errorf("cannot read the ca-deployment json file: %s", err.Error())
		return
	}
	var caDeploymentData map[string]any
	err = json.Unmarshal(data, &caDeploymentData)
	cI.CASettings, err = parseCASettingsInfo(caDeploymentData)
	if err != nil {
		klog.Errorf("cannot parse the ca settings from deployment json: %s", err.Error())
		return
	}
	err = cI.Init()
	return
}

func populateNodeTemplates(nodeGroups map[string]*VirtualNodeGroup, nodeTemplates map[string]gct.NodeTemplate) error {
	for name, template := range nodeTemplates {
		ng, ok := nodeGroups[name]
		if !ok {
			return fmt.Errorf("nodegroup name not found: %s", name)
		}
		ng.nodeTemplate = template
		nodeGroups[name] = ng
	}
	return nil
}

func (v *VirtualCloudProvider) Name() string {
	return cloudprovider.VirtualProviderName
}

func (v *VirtualCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, len(v.clusterInfo.NodeGroups))
	counter := 0
	for _, ng := range v.virtualNodeGroups {
		result[counter] = ng
		counter++
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
	return v.NodeGroupInfo.MaxSize
}

func (v *VirtualNodeGroup) MinSize() int {
	return v.NodeGroupInfo.MinSize
}

func (v *VirtualNodeGroup) TargetSize() (int, error) {
	return v.NodeGroupInfo.TargetSize, nil
}

func (v *VirtualNodeGroup) changeCreatingInstancesToRunning() {
	for i, _ := range v.instances {
		klog.Infof("changing the instace %s state from creating to running", v.instances[i].Id)
		v.instances[i].Status.State = cloudprovider.InstanceRunning
	}
}

func (v *VirtualNodeGroup) IncreaseSize(delta int) error {
	//TODO add flags for simulating provider errors ex : ResourceExhaustion
	for i := 0; i < delta; i++ {
		node, err := v.buildCoreNodeFromTemplate()
		if err != nil {
			return err
		}
		v.instances = append(v.instances, cloudprovider.Instance{
			Id: node.Name,
			Status: &cloudprovider.InstanceStatus{
				State:     cloudprovider.InstanceCreating,
				ErrorInfo: nil,
			},
		})

	}
	time.AfterFunc(10*time.Second, v.changeCreatingInstancesToRunning)
	return nil
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
	return v.instances, nil
}

func buildGenericLabels(template *gct.NodeTemplate, nodeName string) map[string]string {
	result := make(map[string]string)
	// TODO: extract from MCM
	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[corev1.LabelArchStable] = cloudprovider.DefaultArch

	result[kubeletapis.LabelOS] = cloudprovider.DefaultOS
	result[corev1.LabelOSStable] = cloudprovider.DefaultOS

	result[corev1.LabelInstanceType] = template.InstanceType
	result[corev1.LabelInstanceTypeStable] = template.InstanceType

	result[corev1.LabelZoneRegion] = template.Region
	result[corev1.LabelZoneRegionStable] = template.Region

	result[corev1.LabelZoneFailureDomain] = template.Zone
	result[corev1.LabelZoneFailureDomainStable] = template.Zone

	result[corev1.LabelHostname] = nodeName
	return result
}

func (v *VirtualNodeGroup) buildCoreNodeFromTemplate() (corev1.Node, error) {
	node := corev1.Node{}
	nodeName := fmt.Sprintf("%s-%d", v.Name, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = corev1.NodeStatus{
		Capacity: corev1.ResourceList{},
	}

	node.Status.Capacity[corev1.ResourcePods] = resource.MustParse("110") //Fixme must take it dynamically from node object
	node.Status.Capacity[corev1.ResourceCPU] = v.nodeTemplate.CPU
	if v.nodeTemplate.GPU.Cmp(resource.MustParse("0")) != 0 {
		node.Status.Capacity[gpu.ResourceNvidiaGPU] = v.nodeTemplate.GPU
	}
	node.Status.Capacity[corev1.ResourceMemory] = v.nodeTemplate.Memory
	node.Status.Capacity[corev1.ResourceEphemeralStorage] = v.nodeTemplate.EphemeralStorage
	// added most common hugepages sizes. This will help to consider the template node while finding similar node groups
	node.Status.Capacity["hugepages-1Gi"] = *resource.NewQuantity(0, resource.DecimalSI)
	node.Status.Capacity["hugepages-2Mi"] = *resource.NewQuantity(0, resource.DecimalSI)

	node.Status.Allocatable = node.Status.Capacity

	// NodeLabels
	node.Labels = v.nodeTemplate.Tags
	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(&v.nodeTemplate, nodeName))

	//TODO populate taints from mcd
	//node.Spec.Taints = v.nodeTemplate.Taints

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return node, nil
}

func (v *VirtualNodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	coreNode, err := v.buildCoreNodeFromTemplate()
	if err != nil {
		return nil, err
	}
	//TODO Discuss it
	nodeInfo := schedulerframework.NewNodeInfo(cloudprovider.BuildKubeProxy(v.Name))
	nodeInfo.SetNode(&coreNode)
	return nodeInfo, nil
}

func (v *VirtualNodeGroup) Exist() bool {
	return true
}

func (v *VirtualNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

func (v *VirtualNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

func (v *VirtualNodeGroup) Autoprovisioned() bool {
	return false
}

func (v *VirtualNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	//TODO copy from mcm get options
	return &defaults, nil
}
