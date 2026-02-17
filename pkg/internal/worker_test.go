package internal

import (
	"reflect"
	"testing"
)

// Test data helpers
func testHelmChart(chartName, version string, values map[string]interface{}) HelmChart {
	if values == nil {
		values = make(map[string]interface{})
	}
	return HelmChart{
		ChartName: chartName,
		Version:   version,
		Values:    values,
	}
}

func testCluster(name, contextName string, helmValues map[string]interface{}) Cluster {
	return Cluster{
		Name:                name,
		ContextName:         contextName,
		KubeConfigPath:      "/test/kubeconfig",
		ControlPlaneAddress: "https://api." + name + ".local",
		NodeIP:              "10.0.0.1",
		HelmValues:          helmValues,
	}
}

func globalValues() map[string]interface{} {
	return map[string]interface{}{
		"kubesliceNetworking.enabled": true,
		"global.profile.openshift":   false,
		"resources.limits.cpu":       "500m",
		"resources.limits.memory":    "512Mi",
		"resources.requests.cpu":     "100m",
		"resources.requests.memory":  "128Mi",
		"replicaCount":               2,
		"image.pullPolicy":           "IfNotPresent",
	}
}

func openShiftWorkerValues() map[string]interface{} {
	return map[string]interface{}{
		"global.profile.openshift":                    true,
		"nodeSelector.node-role.kubernetes.io/worker": "",
		"tolerations[0].key":                          "node-role.kubernetes.io/master",
		"tolerations[0].effect":                       "NoSchedule",
	}
}

func highPerformanceWorkerValues() map[string]interface{} {
	return map[string]interface{}{
		"resources.limits.cpu":    "2000m",
		"resources.limits.memory": "4Gi",
		"resources.requests.cpu":  "1000m",
		"resources.requests.memory": "2Gi",
		"nodeSelector.performance": "high",
		"nodeSelector.zone":       "us-west-1a",
		"replicaCount":            3,
	}
}

func gpuWorkerValues() map[string]interface{} {
	return map[string]interface{}{
		"nodeSelector.accelerator":        "nvidia-tesla-k80",
		"tolerations[0].key":              "nvidia.com/gpu",
		"tolerations[0].operator":         "Exists",
		"tolerations[0].effect":           "NoSchedule",
		"resources.limits.nvidia.com/gpu": 1,
	}
}

func edgeWorkerValues() map[string]interface{} {
	return map[string]interface{}{
		"resources.limits.cpu":    "200m",
		"resources.limits.memory": "256Mi",
		"resources.requests.cpu":  "100m",
		"resources.requests.memory": "128Mi",
		"nodeSelector.node-type":  "edge",
		"replicaCount":            1,
	}
}

func mergeValues(base, override map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Copy base values
	for k, v := range base {
		result[k] = v
	}
	
	// Override with new values
	for k, v := range override {
		result[k] = v
	}
	
	return result
}

func assertHelmChartEqual(t *testing.T, got, want HelmChart) {
	t.Helper()
	
	if got.ChartName != want.ChartName {
		t.Errorf("ChartName = %v, want %v", got.ChartName, want.ChartName)
	}
	
	if got.Version != want.Version {
		t.Errorf("Version = %v, want %v", got.Version, want.Version)
	}
	
	if !reflect.DeepEqual(got.Values, want.Values) {
		t.Errorf("Values mismatch:\nGot:  %v\nWant: %v", got.Values, want.Values)
	}
}

func assertValueExists(t *testing.T, values map[string]interface{}, key string, expectedValue interface{}) {
	t.Helper()
	
	if actualValue, exists := values[key]; !exists {
		t.Errorf("Expected key %s not found in values", key)
	} else if actualValue != expectedValue {
		t.Errorf("Value for %s = %v, want %v", key, actualValue, expectedValue)
	}
}

func TestCreateWorkerSpecificHelmChart(t *testing.T) {
	tests := []struct {
		name         string
		globalChart  HelmChart
		cluster      Cluster
		expectedChart HelmChart
	}{
		{
			name:        "Global values only - backward compatibility",
			globalChart: testHelmChart("kubeslice-worker", "1.0.0", globalValues()),
			cluster:     testCluster("standard-worker", "standard-context", nil),
			expectedChart: testHelmChart("kubeslice-worker", "1.0.0", globalValues()),
		},
		{
			name:        "OpenShift worker with overrides",
			globalChart: testHelmChart("kubeslice-worker", "1.0.0", globalValues()),
			cluster:     testCluster("openshift-worker", "openshift-context", openShiftWorkerValues()),
			expectedChart: testHelmChart("kubeslice-worker", "1.0.0", 
				mergeValues(globalValues(), openShiftWorkerValues())),
		},
		{
			name:        "High-performance worker with resource overrides",
			globalChart: testHelmChart("kubeslice-worker", "1.0.0", globalValues()),
			cluster:     testCluster("high-perf-worker", "high-perf-context", highPerformanceWorkerValues()),
			expectedChart: testHelmChart("kubeslice-worker", "1.0.0", 
				mergeValues(globalValues(), highPerformanceWorkerValues())),
		},
		{
			name:        "GPU worker with specialized configuration",
			globalChart: testHelmChart("kubeslice-worker", "1.0.0", globalValues()),
			cluster:     testCluster("gpu-worker", "gpu-context", gpuWorkerValues()),
			expectedChart: testHelmChart("kubeslice-worker", "1.0.0", 
				mergeValues(globalValues(), gpuWorkerValues())),
		},
		{
			name:        "Edge worker with minimal resources",
			globalChart: testHelmChart("kubeslice-worker", "1.0.0", globalValues()),
			cluster:     testCluster("edge-worker", "edge-context", edgeWorkerValues()),
			expectedChart: testHelmChart("kubeslice-worker", "1.0.0", 
				mergeValues(globalValues(), edgeWorkerValues())),
		},
		{
			name:        "Empty global values with cluster-specific values",
			globalChart: testHelmChart("kubeslice-worker", "1.0.0", map[string]interface{}{}),
			cluster:     testCluster("custom-worker", "custom-context", openShiftWorkerValues()),
			expectedChart: testHelmChart("kubeslice-worker", "1.0.0", openShiftWorkerValues()),
		},
		{
			name:        "No values at all",
			globalChart: testHelmChart("kubeslice-worker", "1.0.0", map[string]interface{}{}),
			cluster:     testCluster("minimal-worker", "minimal-context", nil),
			expectedChart: testHelmChart("kubeslice-worker", "1.0.0", map[string]interface{}{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateWorkerSpecificHelmChart(tt.globalChart, tt.cluster)
			assertHelmChartEqual(t, result, tt.expectedChart)
		})
	}
}

func TestCreateWorkerSpecificHelmChart_BackwardCompatibility(t *testing.T) {
	t.Run("Legacy configuration without HelmValues", func(t *testing.T) {
		globalChart := testHelmChart("kubeslice-worker", "1.0.0", globalValues())
		
		// Simulate old cluster configuration without HelmValues field
		cluster := Cluster{
			Name:                "legacy-worker",
			ContextName:         "legacy-context",
			KubeConfigPath:      "/path/to/kubeconfig",
			ControlPlaneAddress: "https://api.cluster.local",
			NodeIP:              "10.0.0.1",
			// HelmValues is nil (not set)
		}

		result := CreateWorkerSpecificHelmChart(globalChart, cluster)
		
		// Should return exactly the global chart values
		expectedChart := testHelmChart("kubeslice-worker", "1.0.0", globalValues())
		assertHelmChartEqual(t, result, expectedChart)
	})
}

func TestCreateWorkerSpecificHelmChart_ValueMerging(t *testing.T) {
	t.Run("Values are properly merged and overridden", func(t *testing.T) {
		globalValues := map[string]interface{}{
			"resources.limits.cpu":     "500m",
			"resources.limits.memory":  "512Mi",
			"nodeSelector.environment": "production",
			"replicaCount":             2,
			"image.tag":                "v1.0.0",
		}
		
		clusterValues := map[string]interface{}{
			"resources.limits.cpu":     "2000m",  // Override
			"resources.limits.memory":  "2Gi",    // Override
			"nodeSelector.performance": "high",   // New
			"replicaCount":             5,        // Override
		}
		
		globalChart := testHelmChart("kubeslice-worker", "1.0.0", globalValues)
		cluster := testCluster("test-worker", "test-context", clusterValues)
		
		result := CreateWorkerSpecificHelmChart(globalChart, cluster)
		
		// Check that global values are preserved when not overridden
		assertValueExists(t, result.Values, "nodeSelector.environment", "production")
		assertValueExists(t, result.Values, "image.tag", "v1.0.0")
		
		// Check that cluster values override global values
		assertValueExists(t, result.Values, "resources.limits.cpu", "2000m")
		assertValueExists(t, result.Values, "resources.limits.memory", "2Gi")
		assertValueExists(t, result.Values, "replicaCount", 5)
		
		// Check that new cluster values are added
		assertValueExists(t, result.Values, "nodeSelector.performance", "high")
	})
}

func TestCreateWorkerSpecificHelmChart_EdgeCases(t *testing.T) {
	t.Run("Nil global values with cluster values", func(t *testing.T) {
		globalChart := HelmChart{
			ChartName: "kubeslice-worker",
			Version:   "1.0.0",
			Values:    nil, // Nil values
		}
		
		clusterValues := map[string]interface{}{
			"test.key": "test.value",
		}
		
		cluster := testCluster("test-worker", "test-context", clusterValues)
		result := CreateWorkerSpecificHelmChart(globalChart, cluster)
		
		assertValueExists(t, result.Values, "test.key", "test.value")
	})
	
	t.Run("Empty string values are preserved", func(t *testing.T) {
		globalValues := map[string]interface{}{
			"emptyString": "",
			"normalValue": "test",
		}
		
		clusterValues := map[string]interface{}{
			"emptyString": "overridden",
			"newEmpty":    "",
		}
		
		globalChart := testHelmChart("kubeslice-worker", "1.0.0", globalValues)
		cluster := testCluster("test-worker", "test-context", clusterValues)
		
		result := CreateWorkerSpecificHelmChart(globalChart, cluster)
		
		assertValueExists(t, result.Values, "emptyString", "overridden")
		assertValueExists(t, result.Values, "normalValue", "test")
		assertValueExists(t, result.Values, "newEmpty", "")
	})
}