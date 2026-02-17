package internal

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

func assertValuesEqual(t *testing.T, got, want map[string]interface{}) {
	t.Helper()
	
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Values mismatch:\nGot:  %v\nWant: %v", got, want)
	}
}

func assertValueExistsIntegration(t *testing.T, values map[string]interface{}, key string, expectedValue interface{}) {
	t.Helper()
	
	if actualValue, exists := values[key]; !exists {
		t.Errorf("Expected key %s not found in values", key)
	} else if actualValue != expectedValue {
		t.Errorf("Value for %s = %v, want %v", key, actualValue, expectedValue)
	}
}

func TestWorkerChartCreation_Integration(t *testing.T) {
	globalValues := map[string]interface{}{
		"kubesliceNetworking.enabled": true,
		"global.profile.openshift":   false,
		"resources.limits.cpu":       "500m",
		"resources.limits.memory":    "512Mi",
	}

	tests := []struct {
		name           string
		cluster        Cluster
		expectedValues map[string]interface{}
	}{
		{
			name: "Standard worker with global values only",
			cluster: Cluster{
				Name:        "standard-worker",
				ContextName: "standard-context",
				HelmValues:  nil,
			},
			expectedValues: globalValues,
		},
		{
			name: "OpenShift worker with overrides",
			cluster: Cluster{
				Name:        "openshift-worker",
				ContextName: "openshift-context",
				HelmValues: map[string]interface{}{
					"global.profile.openshift": true,
					"nodeSelector.zone":        "us-west-1a",
				},
			},
			expectedValues: map[string]interface{}{
				"kubesliceNetworking.enabled": true,
				"global.profile.openshift":   true, // Overridden
				"resources.limits.cpu":       "500m",
				"resources.limits.memory":    "512Mi",
				"nodeSelector.zone":          "us-west-1a", // New
			},
		},
		{
			name: "High-performance worker with resource overrides",
			cluster: Cluster{
				Name:        "high-perf-worker",
				ContextName: "high-perf-context",
				HelmValues: map[string]interface{}{
					"resources.limits.cpu":    "2000m",
					"resources.limits.memory": "4Gi",
					"nodeSelector.performance": "high",
				},
			},
			expectedValues: map[string]interface{}{
				"kubesliceNetworking.enabled": true,
				"global.profile.openshift":   false,
				"resources.limits.cpu":       "2000m", // Overridden
				"resources.limits.memory":    "4Gi",   // Overridden
				"nodeSelector.performance":   "high",  // New
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalChart := HelmChart{
				ChartName: "kubeslice-worker",
				Version:   "1.0.0",
				Values:    globalValues,
			}
			
			workerChart := CreateWorkerSpecificHelmChart(globalChart, tt.cluster)
			assertValuesEqual(t, workerChart.Values, tt.expectedValues)
		})
	}
}

func TestConfigurationParsing_WithPerWorkerValues(t *testing.T) {
	yamlContent := `
configuration:
  cluster_configuration:
    controller:
      name: controller
      context_name: controller-context
    workers:
      - name: standard-worker
        context_name: standard-worker-context
      - name: openshift-worker
        context_name: openshift-worker-context
        helm_values:
          global.profile.openshift: true
          nodeSelector.zone: us-west-1a
      - name: high-perf-worker
        context_name: high-perf-worker-context
        helm_values:
          resources.limits.cpu: "2000m"
          resources.limits.memory: "4Gi"
          nodeSelector.performance: "high"
  kubeslice_configuration:
    project_name: test-project
  helm_chart_configuration:
    repo_alias: kubeslice
    worker_chart:
      chart_name: kubeslice-worker
      values:
        kubesliceNetworking.enabled: true
        global.profile.openshift: false
        resources.limits.cpu: "500m"
        resources.limits.memory: "512Mi"
`

	var config ConfigurationSpecs
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	t.Run("Configuration parsing", func(t *testing.T) {
		if len(config.Configuration.ClusterConfiguration.WorkerClusters) != 3 {
			t.Errorf("Expected 3 worker clusters, got %d", len(config.Configuration.ClusterConfiguration.WorkerClusters))
		}
	})

	t.Run("Standard worker has no helm values", func(t *testing.T) {
		worker := config.Configuration.ClusterConfiguration.WorkerClusters[0]
		if worker.Name != "standard-worker" {
			t.Errorf("Expected worker name 'standard-worker', got '%s'", worker.Name)
		}
		if worker.HelmValues != nil {
			t.Errorf("Expected standard worker HelmValues to be nil, got %v", worker.HelmValues)
		}
	})

	t.Run("OpenShift worker has correct helm values", func(t *testing.T) {
		worker := config.Configuration.ClusterConfiguration.WorkerClusters[1]
		if worker.Name != "openshift-worker" {
			t.Errorf("Expected worker name 'openshift-worker', got '%s'", worker.Name)
		}
		
		if worker.HelmValues == nil {
			t.Fatal("Expected openshift worker to have helm_values, but it was nil")
		}
		
		assertValueExistsIntegration(t, worker.HelmValues, "global.profile.openshift", true)
		assertValueExistsIntegration(t, worker.HelmValues, "nodeSelector.zone", "us-west-1a")
	})

	t.Run("High-performance worker has correct helm values", func(t *testing.T) {
		worker := config.Configuration.ClusterConfiguration.WorkerClusters[2]
		if worker.Name != "high-perf-worker" {
			t.Errorf("Expected worker name 'high-perf-worker', got '%s'", worker.Name)
		}
		
		if worker.HelmValues == nil {
			t.Fatal("Expected high-perf worker to have helm_values, but it was nil")
		}
		
		assertValueExistsIntegration(t, worker.HelmValues, "resources.limits.cpu", "2000m")
		assertValueExistsIntegration(t, worker.HelmValues, "resources.limits.memory", "4Gi")
		assertValueExistsIntegration(t, worker.HelmValues, "nodeSelector.performance", "high")
	})
}

func TestBackwardCompatibility_ExistingConfigurations(t *testing.T) {
	yamlContent := `
configuration:
  cluster_configuration:
    controller:
      name: controller
      context_name: controller-context
    workers:
      - name: legacy-worker
        context_name: legacy-context
  kubeslice_configuration:
    project_name: test-project
  helm_chart_configuration:
    repo_alias: kubeslice
    worker_chart:
      chart_name: kubeslice-worker
      values:
        kubesliceNetworking.enabled: true
        global.profile.openshift: false
`

	var config ConfigurationSpecs
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	t.Run("Legacy configuration parsing", func(t *testing.T) {
		if len(config.Configuration.ClusterConfiguration.WorkerClusters) != 1 {
			t.Errorf("Expected 1 worker cluster, got %d", len(config.Configuration.ClusterConfiguration.WorkerClusters))
		}

		worker := config.Configuration.ClusterConfiguration.WorkerClusters[0]
		if worker.Name != "legacy-worker" {
			t.Errorf("Expected worker name 'legacy-worker', got '%s'", worker.Name)
		}

		if worker.HelmValues != nil {
			t.Errorf("Expected legacy worker HelmValues to be nil, got %v", worker.HelmValues)
		}
	})

	t.Run("Legacy worker chart creation", func(t *testing.T) {
		globalChart := config.Configuration.HelmChartConfiguration.WorkerChart
		worker := config.Configuration.ClusterConfiguration.WorkerClusters[0]
		
		workerChart := CreateWorkerSpecificHelmChart(globalChart, worker)

		expectedValues := map[string]interface{}{
			"kubesliceNetworking.enabled": true,
			"global.profile.openshift":   false,
		}

		assertValuesEqual(t, workerChart.Values, expectedValues)
	})
}

func TestMultiWorkerScenario_Integration(t *testing.T) {
	t.Run("Multiple workers with different configurations", func(t *testing.T) {
		globalValues := map[string]interface{}{
			"kubesliceNetworking.enabled": true,
			"global.profile.openshift":   false,
			"resources.limits.cpu":       "500m",
			"resources.limits.memory":    "512Mi",
			"replicaCount":               2,
		}
		
		globalChart := HelmChart{
			ChartName: "kubeslice-worker",
			Version:   "1.0.0",
			Values:    globalValues,
		}
		
		workers := []struct {
			cluster        Cluster
			expectedValues map[string]interface{}
		}{
			{
				cluster: Cluster{
					Name:        "standard-worker",
					ContextName: "standard-context",
					HelmValues:  nil,
				},
				expectedValues: globalValues,
			},
			{
				cluster: Cluster{
					Name:        "openshift-worker",
					ContextName: "openshift-context",
					HelmValues: map[string]interface{}{
						"global.profile.openshift": true,
						"nodeSelector.zone":        "us-east-1a",
					},
				},
				expectedValues: map[string]interface{}{
					"kubesliceNetworking.enabled": true,
					"global.profile.openshift":   true, // Overridden
					"resources.limits.cpu":       "500m",
					"resources.limits.memory":    "512Mi",
					"replicaCount":               2,
					"nodeSelector.zone":          "us-east-1a", // New
				},
			},
			{
				cluster: Cluster{
					Name:        "gpu-worker",
					ContextName: "gpu-context",
					HelmValues: map[string]interface{}{
						"nodeSelector.accelerator":        "nvidia-tesla-k80",
						"resources.limits.nvidia.com/gpu": 1,
					},
				},
				expectedValues: map[string]interface{}{
					"kubesliceNetworking.enabled":     true,
					"global.profile.openshift":       false,
					"resources.limits.cpu":           "500m",
					"resources.limits.memory":        "512Mi",
					"replicaCount":                   2,
					"nodeSelector.accelerator":       "nvidia-tesla-k80", // New
					"resources.limits.nvidia.com/gpu": 1,                // New
				},
			},
			{
				cluster: Cluster{
					Name:        "edge-worker",
					ContextName: "edge-context",
					HelmValues: map[string]interface{}{
						"resources.limits.cpu":    "200m",
						"resources.limits.memory": "256Mi",
						"replicaCount":            1,
					},
				},
				expectedValues: map[string]interface{}{
					"kubesliceNetworking.enabled": true,
					"global.profile.openshift":   false,
					"resources.limits.cpu":       "200m",  // Overridden
					"resources.limits.memory":    "256Mi", // Overridden
					"replicaCount":               1,       // Overridden
				},
			},
		}

		for _, w := range workers {
			t.Run("Worker_"+w.cluster.Name, func(t *testing.T) {
				workerChart := CreateWorkerSpecificHelmChart(globalChart, w.cluster)
				assertValuesEqual(t, workerChart.Values, w.expectedValues)
			})
		}
	})
}