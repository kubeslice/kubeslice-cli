package internal

import (
	"strings"
	"testing"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateClusterRegistrationManifest tests YAML generation logic
func TestGenerateClusterRegistrationManifest(t *testing.T) {
	tests := []struct {
		name             string
		config           *ConfigurationSpecs
		filename         string
		namespace        string
		expectContent    []string
		expectRegionInfo bool
	}{
		{
			name: "single worker without enterprise profile",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						WorkerClusters: []Cluster{
							{Name: "worker-1"},
						},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName: "test-project",
					},
				},
			},
			filename:  "test-output.yaml",
			namespace: "",
			expectContent: []string{
				"apiVersion: controller.kubeslice.io/v1alpha1",
				"kind: Cluster",
				"name: worker-1",
				"namespace: kubeslice-test-project",
				"clusterProperty: {}",
			},
			expectRegionInfo: false,
		},
		{
			name: "multiple workers with custom namespace",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						WorkerClusters: []Cluster{
							{Name: "worker-1"},
							{Name: "worker-2"},
						},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName: "test-project",
					},
				},
			},
			filename:  "test-output.yaml",
			namespace: "custom-namespace",
			expectContent: []string{
				"name: worker-1",
				"name: worker-2",
				"namespace: custom-namespace",
			},
			expectRegionInfo: false,
		},
		{
			name: "enterprise profile with region templates",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						Profile: ProfileEntDemo,
						WorkerClusters: []Cluster{
							{Name: "ks-w-1"},
							{Name: "ks-w-2"},
						},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName: "ent-project",
					},
				},
			},
			filename:  "test-output.yaml",
			namespace: "",
			expectContent: []string{
				"geoLocation:",
				"cloudProvider: GCP",
				"cloudProvider: DATACENTER",
			},
			expectRegionInfo: true,
		},
		{
			name: "no worker clusters",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						WorkerClusters: []Cluster{},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName: "empty-project",
					},
				},
			},
			filename:         "test-output.yaml",
			namespace:        "",
			expectContent:    []string{},
			expectRegionInfo: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeFS := util.FileSystem.(*util.FakeFileSystem)

			generateClusterRegistrationManifest(tt.config, tt.filename, tt.namespace)

			content, exists := fakeFS.WrittenFiles[tt.filename]
			require.True(t, exists, "Expected file %s to be written", tt.filename)

			contentStr := string(content)

			for _, expected := range tt.expectContent {
				assert.Contains(t, contentStr, expected,
					"Expected content not found in generated manifest")
			}

			hasRegionInfo := strings.Contains(contentStr, "geoLocation")
			assert.Equal(t, tt.expectRegionInfo, hasRegionInfo,
				"Region info presence mismatch")

			// Validate YAML document count
			if len(tt.config.Configuration.ClusterConfiguration.WorkerClusters) > 0 {
				expectedDocs := len(tt.config.Configuration.ClusterConfiguration.WorkerClusters)
				actualDocs := strings.Count(contentStr, "---")
				assert.Equal(t, expectedDocs, actualDocs,
					"Expected %d YAML documents", expectedDocs)
			}
		})
	}
}

// TestRegisterWorkerClusters tests the full registration workflow
func TestRegisterWorkerClusters(t *testing.T) {
	tests := []struct {
		name                 string
		config               *ConfigurationSpecs
		cliOptions           *CliOptionsStruct
		mockApply            func(*testing.T, string, string, *Cluster)
		applyCalled          int
		expectFileGeneration bool
	}{
		{
			name: "with CLI options - custom filename provided",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						WorkerClusters: []Cluster{{Name: "worker-1"}},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName: "test",
					},
				},
			},
			cliOptions: &CliOptionsStruct{
				FileName:  "custom-file.yaml",
				Namespace: "custom-ns",
				Cluster:   &Cluster{Name: "controller"},
			},
			mockApply: func(t *testing.T, filename, namespace string, cluster *Cluster) {
				assert.Equal(t, "custom-file.yaml", filename)
				assert.Equal(t, "custom-ns", namespace)
				assert.Equal(t, "controller", cluster.Name)
			},
			applyCalled:          1,
			expectFileGeneration: false,
		},
		{
			name: "with CLI options - filename auto-generated",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						WorkerClusters: []Cluster{{Name: "worker-1"}},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName: "test",
					},
				},
			},
			cliOptions: &CliOptionsStruct{
				FileName:  "",
				Namespace: "custom-ns",
				Cluster:   &Cluster{Name: "controller"},
			},
			mockApply: func(t *testing.T, filename, namespace string, cluster *Cluster) {
				assert.Contains(t, filename, "custom-cluster-registration.yaml")
				assert.Equal(t, "custom-ns", namespace)
			},
			applyCalled:          1,
			expectFileGeneration: true,
		},
		{
			name: "without CLI options - default flow",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
						WorkerClusters:    []Cluster{{Name: "worker-1"}},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName: "demo",
					},
				},
			},
			cliOptions: nil,
			mockApply: func(t *testing.T, filename, namespace string, cluster *Cluster) {
				assert.Contains(t, filename, clusterRegistrationFileName)
				assert.Equal(t, "kubeslice-demo", namespace)
				assert.Equal(t, "controller", cluster.Name)
			},
			applyCalled:          1,
			expectFileGeneration: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeFS := util.FileSystem.(*util.FakeFileSystem)
			fakeClock := util.SystemClock.(*util.FakeClock)

			applyCalls := 0
			originalApply := applyManifestFunc
			applyManifestFunc = func(filename, namespace string, cluster *Cluster) {
				applyCalls++
				if tt.mockApply != nil {
					tt.mockApply(t, filename, namespace, cluster)
				}
			}
			defer func() { applyManifestFunc = originalApply }()

			RegisterWorkerClusters(tt.config, tt.cliOptions)

			if tt.expectFileGeneration {
				assert.NotEmpty(t, fakeFS.WrittenFiles, "Expected files to be generated")
			} else {
				assert.Empty(t, fakeFS.WrittenFiles, "Should not generate files when custom filename provided")
			}

			assert.Equal(t, tt.applyCalled, applyCalls,
				"ApplyKubectlManifest call count mismatch")

			assert.Len(t, fakeClock.SleepCalls, 2, "Expected 2 sleep calls")
			for _, d := range fakeClock.SleepCalls {
				assert.Equal(t, 200*time.Millisecond, d)
			}
		})
	}
}

// TestGetKubeSliceCluster tests fetching cluster info
func TestGetKubeSliceCluster(t *testing.T) {
	tests := []struct {
		name         string
		clusterName  string
		namespace    string
		outputFormat string
	}{
		{
			name:         "get cluster with YAML output",
			clusterName:  "worker-1",
			namespace:    "kubeslice-test",
			outputFormat: "yaml",
		},
		{
			name:         "get cluster with JSON output",
			clusterName:  "worker-2",
			namespace:    "default",
			outputFormat: "json",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeClock := util.SystemClock.(*util.FakeClock)

			getCalled := false
			original := getResourceFunc
			getResourceFunc = func(resourceType, resourceName, namespace string, cluster *Cluster, outputFormat string) {
				getCalled = true
				assert.Equal(t, ClusterObject, resourceType)
				assert.Equal(t, tt.clusterName, resourceName)
				assert.Equal(t, tt.namespace, namespace)
				assert.Equal(t, tt.outputFormat, outputFormat)
			}
			defer func() { getResourceFunc = original }()

			cluster := &Cluster{Name: "controller"}
			GetKubeSliceCluster(tt.clusterName, tt.namespace, cluster, tt.outputFormat)

			assert.True(t, getCalled, "Expected GetKubectlResources to be called")
			assert.Contains(t, fakeClock.SleepCalls, 200*time.Millisecond)
		})
	}
}

// TestDeleteKubeSliceCluster tests cluster deletion
func TestDeleteKubeSliceCluster(t *testing.T) {
	cleanup := util.NewTestEnvironment()
	defer cleanup()

	fakeClock := util.SystemClock.(*util.FakeClock)

	deleteCalled := false
	original := deleteResourceFunc
	deleteResourceFunc = func(resourceType, resourceName, namespace string, cluster *Cluster) {
		deleteCalled = true
		assert.Equal(t, ClusterObject, resourceType)
		assert.Equal(t, "worker-1", resourceName)
		assert.Equal(t, "kubeslice-test", namespace)
		assert.Equal(t, "controller", cluster.Name)
	}
	defer func() { deleteResourceFunc = original }()

	cluster := &Cluster{Name: "controller"}
	DeleteKubeSliceCluster("worker-1", "kubeslice-test", cluster)

	assert.True(t, deleteCalled, "Expected DeleteKubectlResources to be called")
	assert.Contains(t, fakeClock.SleepCalls, 200*time.Millisecond)
}

// TestEditKubeSliceCluster tests cluster editing
func TestEditKubeSliceCluster(t *testing.T) {
	cleanup := util.NewTestEnvironment()
	defer cleanup()

	fakeClock := util.SystemClock.(*util.FakeClock)

	editCalled := false
	original := editResourceFunc
	editResourceFunc = func(resourceType, resourceName, namespace string, cluster *Cluster) {
		editCalled = true
		assert.Equal(t, ClusterObject, resourceType)
		assert.Equal(t, "worker-1", resourceName)
		assert.Equal(t, "kubeslice-test", namespace)
		assert.Equal(t, "controller", cluster.Name)
	}
	defer func() { editResourceFunc = original }()

	cluster := &Cluster{Name: "controller"}
	EditKubeSliceCluster("worker-1", "kubeslice-test", cluster)

	assert.True(t, editCalled, "Expected EditKubectlResources to be called")
	assert.Contains(t, fakeClock.SleepCalls, 200*time.Millisecond)
}

// TestDescribeKubeSliceCluster tests cluster description
func TestDescribeKubeSliceCluster(t *testing.T) {
	cleanup := util.NewTestEnvironment()
	defer cleanup()

	fakeClock := util.SystemClock.(*util.FakeClock)

	describeCalled := false
	original := describeResourceFunc
	describeResourceFunc = func(resourceType, resourceName, namespace string, cluster *Cluster) {
		describeCalled = true
		assert.Equal(t, ClusterObject, resourceType)
		assert.Equal(t, "worker-1", resourceName)
		assert.Equal(t, "kubeslice-test", namespace)
		assert.Equal(t, "controller", cluster.Name)
	}
	defer func() { describeResourceFunc = original }()

	cluster := &Cluster{Name: "controller"}
	DescribeKubeSliceCluster("worker-1", "kubeslice-test", cluster)

	assert.True(t, describeCalled, "Expected DescribeKubectlResources to be called")
	assert.Contains(t, fakeClock.SleepCalls, 200*time.Millisecond)
}
