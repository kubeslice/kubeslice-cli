package internal

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteKubeSliceDirectory(t *testing.T) {
	tests := []struct {
		name          string
		mockRemoveAll func(string) error
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful deletion",
			mockRemoveAll: func(path string) error {
				assert.Equal(t, kubesliceDirectory, path)
				return nil
			},
		},
		{
			name: "deletion fails",
			mockRemoveAll: func(path string) error {
				return fmt.Errorf("permission denied")
			},
			expectFatal:   true,
			fatalContains: "Failed to delete directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeFS := util.FileSystem.(*util.FakeFileSystem)
			fakeOutput := util.Output.(*util.FakeOutput)

			if tt.mockRemoveAll != nil {
				fakeFS.RemoveAllFunc = tt.mockRemoveAll
			}

			originalRemoveAll := removeAllFunc
			removeAllFunc = fakeFS.RemoveAll
			defer func() { removeAllFunc = originalRemoveAll }()

			DeleteKubeSliceDirectory()

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls, "Expected fatal call but got none")
				if tt.fatalContains != "" {
					callStr := fmt.Sprint(fakeOutput.FatalCalls[0])
					assert.Contains(t, callStr, tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
			}
		})
	}
}

func TestGenerateKubeSliceDirectory(t *testing.T) {
	tests := []struct {
		name          string
		mockMkdirAll  func(string, os.FileMode) error
		mockStat      func(string) (os.FileInfo, error)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful directory creation",
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			mockMkdirAll: func(path string, perm os.FileMode) error {
				assert.Equal(t, kubesliceDirectory, path)
				return nil
			},
		},
		{
			name: "directory already exists",
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, nil
			},
			mockMkdirAll: func(path string, perm os.FileMode) error {
				return nil
			},
		},
		{
			name: "directory creation fails",
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			mockMkdirAll: func(path string, perm os.FileMode) error {
				return fmt.Errorf("permission denied")
			},
			expectFatal:   true,
			fatalContains: "Failed to create kubeslice directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeFS := util.FileSystem.(*util.FakeFileSystem)
			fakeOutput := util.Output.(*util.FakeOutput)

			if tt.mockStat != nil {
				fakeFS.StatFunc = tt.mockStat
			}
			if tt.mockMkdirAll != nil {
				fakeFS.MkdirAllFunc = tt.mockMkdirAll
			}

			GenerateKubeSliceDirectory()

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls, "Expected fatal call but got none")
				if tt.fatalContains != "" {
					callStr := fmt.Sprint(fakeOutput.FatalCalls[0])
					assert.Contains(t, callStr, tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
			}
		})
	}
}

func TestGenerateKindConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		config          *ConfigurationSpecs
		expectedFiles   []string
		validateContent func(*testing.T, map[string][]byte)
		mockFileWrite   func(string, []byte, os.FileMode) error
		expectFatal     bool
		fatalContains   string
	}{
		{
			name: "standard profile with single worker",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						Profile: "",
						ControllerCluster: Cluster{
							Name: "controller",
						},
						WorkerClusters: []Cluster{
							{Name: "worker-1"},
						},
					},
				},
			},
			expectedFiles: []string{
				"kubeslice/kind/controller.yaml",
				"kubeslice/kind/worker-1.yaml",
			},
			validateContent: func(t *testing.T, files map[string][]byte) {
				controllerContent := string(files["kubeslice/kind/controller.yaml"])
				assert.Contains(t, controllerContent, "name: controller")
				assert.Contains(t, controllerContent, "kind: Cluster")
				assert.Contains(t, controllerContent, "disableDefaultCNI: true")
				assert.NotContains(t, controllerContent, "extraPortMappings")

				workerContent := string(files["kubeslice/kind/worker-1.yaml"])
				assert.Contains(t, workerContent, "name: worker-1")
				assert.Contains(t, workerContent, "kubeslice.io/node-type=gateway")
			},
		},
		{
			name: "enterprise profile with port mappings",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						Profile: ProfileEntDemo,
						ControllerCluster: Cluster{
							Name: "ent-controller",
						},
						WorkerClusters: []Cluster{
							{Name: "ent-worker-1"},
						},
					},
				},
			},
			expectedFiles: []string{
				"kubeslice/kind/ent-controller.yaml",
				"kubeslice/kind/ent-worker-1.yaml",
			},
			validateContent: func(t *testing.T, files map[string][]byte) {
				controllerContent := string(files["kubeslice/kind/ent-controller.yaml"])
				assert.Contains(t, controllerContent, "name: ent-controller")
				assert.Contains(t, controllerContent, "extraPortMappings")
				assert.Contains(t, controllerContent, "containerPort: 31000")
				assert.Contains(t, controllerContent, "hostPort: 8443")
			},
		},
		{
			name: "multiple workers",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name: "controller",
						},
						WorkerClusters: []Cluster{
							{Name: "worker-1"},
							{Name: "worker-2"},
							{Name: "worker-3"},
						},
					},
				},
			},
			expectedFiles: []string{
				"kubeslice/kind/controller.yaml",
				"kubeslice/kind/worker-1.yaml",
				"kubeslice/kind/worker-2.yaml",
				"kubeslice/kind/worker-3.yaml",
			},
			validateContent: func(t *testing.T, files map[string][]byte) {
				assert.Len(t, files, 4)
				for i := 1; i <= 3; i++ {
					workerFile := fmt.Sprintf("kubeslice/kind/worker-%d.yaml", i)
					content := string(files[workerFile])
					assert.Contains(t, content, fmt.Sprintf("name: worker-%d", i))
					assert.Contains(t, content, "kubeslice.io/node-type=gateway")
				}
			},
		},
		{
			name: "no workers",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name: "controller-only",
						},
						WorkerClusters: []Cluster{},
					},
				},
			},
			expectedFiles: []string{
				"kubeslice/kind/controller-only.yaml",
			},
			validateContent: func(t *testing.T, files map[string][]byte) {
				assert.Len(t, files, 1)
			},
		},
		{
			name: "special characters in cluster names",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name: "my-controller-01",
						},
						WorkerClusters: []Cluster{
							{Name: "my-worker-01"},
						},
					},
				},
			},
			expectedFiles: []string{
				"kubeslice/kind/my-controller-01.yaml",
				"kubeslice/kind/my-worker-01.yaml",
			},
			validateContent: func(t *testing.T, files map[string][]byte) {
				controllerContent := string(files["kubeslice/kind/my-controller-01.yaml"])
				assert.Contains(t, controllerContent, "name: my-controller-01")
			},
		},
		{
			name: "file write fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
						WorkerClusters:    []Cluster{{Name: "worker"}},
					},
				},
			},
			mockFileWrite: func(filename string, data []byte, perm os.FileMode) error {
				return fmt.Errorf("permission denied")
			},
			expectFatal:   true,
			fatalContains: "Failed to write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeFS := util.FileSystem.(*util.FakeFileSystem)
			fakeOutput := util.Output.(*util.FakeOutput)
			fakeClock := util.SystemClock.(*util.FakeClock)

			if tt.mockFileWrite != nil {
				fakeFS.WriteFileFunc = tt.mockFileWrite
			}

			GenerateKindConfiguration(tt.config)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					callStr := fmt.Sprint(fakeOutput.FatalCalls[0])
					assert.Contains(t, callStr, tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)

				for _, expectedFile := range tt.expectedFiles {
					_, exists := fakeFS.WrittenFiles[expectedFile]
					assert.True(t, exists, "Expected file %s to be created", expectedFile)
				}

				if tt.validateContent != nil {
					tt.validateContent(t, fakeFS.WrittenFiles)
				}

				expectedSleeps := len(tt.config.Configuration.ClusterConfiguration.WorkerClusters) + 1
				assert.Len(t, fakeClock.SleepCalls, expectedSleeps)
				for _, d := range fakeClock.SleepCalls {
					assert.Equal(t, 200*time.Millisecond, d)
				}
			}
		})
	}
}

func TestTemplateFormats(t *testing.T) {
	tests := []struct {
		name                string
		template            string
		clusterName         string
		expectedContents    []string
		notExpectedContents []string
	}{
		{
			name:        "standard controller template",
			template:    kubesliceControllerTemplate,
			clusterName: "test-controller",
			expectedContents: []string{
				"kind: Cluster",
				"apiVersion: kind.x-k8s.io/v1alpha4",
				"name: test-controller",
				"disableDefaultCNI: true",
				"podSubnet: 192.168.0.0/16",
				"role: control-plane",
				"image: kindest/node:v1.25.11",
			},
			notExpectedContents: []string{
				"extraPortMappings",
				"containerPort",
			},
		},
		{
			name:        "enterprise controller template",
			template:    kubesliceEntControllerTemplate,
			clusterName: "ent-controller",
			expectedContents: []string{
				"kind: Cluster",
				"name: ent-controller",
				"extraPortMappings",
				"containerPort: 31000",
				"hostPort: 8443",
				"protocol: TCP",
			},
		},
		{
			name:        "worker template",
			template:    kubesliceWorkerTemplate,
			clusterName: "test-worker",
			expectedContents: []string{
				"kind: Cluster",
				"name: test-worker",
				"kubeadmConfigPatches",
				"kubeletExtraArgs",
				"kubeslice.io/node-type=gateway",
			},
			notExpectedContents: []string{
				"extraPortMappings",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := fmt.Sprintf(tt.template, tt.clusterName)

			for _, expected := range tt.expectedContents {
				assert.Contains(t, content, expected,
					"Template should contain: %s", expected)
			}

			for _, notExpected := range tt.notExpectedContents {
				assert.NotContains(t, content, notExpected,
					"Template should NOT contain: %s", notExpected)
			}
		})
	}
}
