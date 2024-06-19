package virtual

import (
	"errors"
	"fmt"
	gct "github.com/elankath/gardener-cluster-types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const PoolLabel = "worker.gardener.cloud/pool"

var ZoneLabels = []string{"topology.gke.io/zone", "topology.ebs.csi.aws.com/zone"}

var ErrKeyNotFound = errors.New("key not found")

func GetZone(labelsMap map[string]any) string {
	var zone string
	for _, zoneLabel := range ZoneLabels {
		z, ok := labelsMap[zoneLabel]
		if ok {
			zone = z.(string)
			break
		}
	}
	return zone
}

func GetInnerMap(parentMap map[string]any, keys ...string) (map[string]any, error) {
	var mapPath []string
	childMap := parentMap
	for _, k := range keys {
		mapPath = append(mapPath, k)
		mp, ok := childMap[k]
		if !ok {
			return nil, fmt.Errorf("cannot find the child map under mapPath: %s", mapPath)
		}
		childMap, ok = mp.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("child map is not of type map[string] any under the mapPath: %s", mapPath)
		}
	}
	return childMap, nil
}

func GetInnerMapValue(parentMap map[string]any, keys ...string) (any, error) {
	subkeys := keys[:len(keys)-1]
	childMap, err := GetInnerMap(parentMap, subkeys...)
	if err != nil {
		return nil, err
	}
	val, ok := childMap[keys[len(keys)-1]]
	if !ok {
		return nil, fmt.Errorf("could not find value for keys %q : %w", keys, ErrKeyNotFound)
	}
	return val, nil
}

func AsIntOrString(val any) (target intstr.IntOrString, err error) {
	switch v := val.(type) {
	case int64:
		target = intstr.FromInt32(int32(val.(int64)))
	case int32:
		target = intstr.FromInt32(v)
	case string:
		return intstr.FromString(v), nil
	default:
		err = fmt.Errorf("cannot parse value %q as intstr.IntOrString", val)
	}
	return
}

type ClusterInfo struct {
	NodeTemplates map[string]gct.NodeTemplate
	NodeGroups    map[string]gct.NodeGroupInfo
	WorkerPools   []gct.WorkerPoolInfo
	CASettings    gct.CASettingsInfo
}

func (c *ClusterInfo) Init() error {
	for name, minMax := range c.CASettings.NodeGroupsMinMax {
		nodeGroup, ok := c.NodeGroups[name]
		if !ok {
			return fmt.Errorf("no nodegroup with the name %s", name)
		}
		nodeGroup.MinSize = minMax.Min
		nodeGroup.MaxSize = minMax.Max
		c.NodeGroups[name] = nodeGroup
	}
	return nil
}
