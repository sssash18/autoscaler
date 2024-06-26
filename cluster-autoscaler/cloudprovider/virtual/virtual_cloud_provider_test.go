package virtual

import (
	gct "github.com/elankath/gardener-cluster-types"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/json"
	"testing"
	"time"
)

func TestReadClusterInfo(t *testing.T) {
	clusterInfoPath := "/tmp"
	readInitClusterInfo(clusterInfoPath)
}

func TestLoadAutoScalerConfig(t *testing.T) {
	expectedConfig := gct.AutoScalerConfig{
		NodeTemplates: map[string]gct.NodeTemplate{
			"a": {
				Name:             "a",
				CPU:              gct.MustParseQuantity("10Mi"),
				GPU:              gct.MustParseQuantity("12Mi"),
				Memory:           gct.MustParseQuantity("10Gi"),
				EphemeralStorage: gct.MustParseQuantity("11Gi"),
				InstanceType:     "m5.large",
				Region:           "eu-west-1",
				Zone:             "eu-west-1a",
			},
		},
		NodeGroups: map[string]gct.NodeGroupInfo{
			"a": {
				Name:       "a",
				PoolName:   "p1",
				Zone:       "eu-west-1a",
				TargetSize: 2,
				MinSize:    1,
				MaxSize:    5,
			},
		},
		/*WorkerPools: []gct.WorkerPoolInfo{
			{
				Architecture: "arm64",
				Minimum:      1,
				Maximum:      5,
				MaxSurge: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "10%",
				},
				MaxUnavailable: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "15%",
				},
				Zones: []string{"eu-west-1a"},
			},
		},*/
		CASettings: gct.CASettingsInfo{
			Expander: "least-waste",
			NodeGroupsMinMax: map[string]gct.MinMax{
				"a": gct.MinMax{
					Min: 1,
					Max: 5,
				},
			},
			MaxNodeProvisionTime:          10 * time.Minute,
			ScanInterval:                  10 * time.Second,
			MaxGracefulTerminationSeconds: 10,
			NewPodScaleUpDelay:            5,
			MaxEmptyBulkDelete:            2,
			IgnoreDaemonSetUtilization:    false,
			MaxNodesTotal:                 10,
			Priorities:                    "dummy",
		},
	}
	bytes, err := json.Marshal(expectedConfig)
	assert.Nil(t, err)
	var actualLoadedConfig gct.AutoScalerConfig
	err = json.Unmarshal(bytes, &actualLoadedConfig)
	assert.Nil(t, err)
	assert.Equal(t, expectedConfig, actualLoadedConfig)

}
