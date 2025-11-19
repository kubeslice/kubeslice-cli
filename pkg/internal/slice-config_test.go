package internal

import (
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSliceConfiguration(t *testing.T) {
	tests := []struct {
		name              string
		config            *ConfigurationSpecs
		workers           []string
		sliceConfigName   string
		namespace         string
		expectedFileName  string
		expectedClusters  string
		expectedNamespace string
	}{
		{
			name: "generate with default workers and name",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						WorkerClusters: []Cluster{{Name: "worker-1"}, {Name: "worker-2"}},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{ProjectName: "test-project"},
				},
			},
			workers:           []string{},
			sliceConfigName:   "",
			namespace:         "",
			expectedFileName:  "slice-demo.yaml",
			expectedClusters:  "clusters: [worker-1,worker-2]",
			expectedNamespace: "namespace: kubeslice-test-project",
		},
		{
			name: "generate with custom workers",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						WorkerClusters: []Cluster{{Name: "worker-1"}, {Name: "worker-2"}},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{ProjectName: "test-project"},
				},
			},
			workers:           []string{"worker-1", "worker-3"},
			sliceConfigName:   "custom-slice",
			namespace:         "custom-namespace",
			expectedFileName:  "slice-custom-slice.yaml",
			expectedClusters:  "clusters: [worker-1,worker-3]",
			expectedNamespace: "namespace: custom-namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeFS := util.FileSystem.(*util.FakeFileSystem)
			fakeClock := util.SystemClock.(*util.FakeClock)

			GenerateSliceConfiguration(tt.config, tt.workers, tt.sliceConfigName, tt.namespace)

			filePath := kubesliceDirectory + "/" + tt.expectedFileName
			content, exists := fakeFS.WrittenFiles[filePath]
			require.True(t, exists)
			contentStr := string(content)
			assert.Contains(t, contentStr, "kind: SliceConfig")
			assert.Contains(t, contentStr, tt.expectedClusters)
			assert.Contains(t, contentStr, tt.expectedNamespace)

			require.Len(t, fakeClock.SleepCalls, 1)
			assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
		})
	}
}

func TestApplySliceConfiguration(t *testing.T) {
	config := &ConfigurationSpecs{
		Configuration: Configuration{
			ClusterConfiguration: ClusterConfiguration{
				ControllerCluster: Cluster{Name: "controller"},
			},
		},
	}

	tests := []struct {
		name          string
		mockVerify    func(*ConfigurationSpecs)
		mockApply     func(string, string, *Cluster)
		expectFatal   bool
		fatalContains string
	}{
		{
			name:       "successful application",
			mockVerify: func(cfg *ConfigurationSpecs) {},
			mockApply: func(fileName, namespace string, cluster *Cluster) {
				assert.Contains(t, fileName, sliceTemplateFileName)
				assert.Equal(t, "kubeslice-demo", namespace)
			},
		},
		{
			name:       "apply fails",
			mockVerify: func(cfg *ConfigurationSpecs) {},
			mockApply: func(fileName, namespace string, cluster *Cluster) {
				util.Fatalf("Process failed %v", errors.New("apply failed"))
			},
			expectFatal:   true,
			fatalContains: "apply failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeOutput := util.Output.(*util.FakeOutput)

			originalVerify := verifyNodeIPsInClustersFunc
			verifyNodeIPsInClustersFunc = tt.mockVerify
			defer func() { verifyNodeIPsInClustersFunc = originalVerify }()

			originalApply := applyKubectlManifestFuncSlice
			applyKubectlManifestFuncSlice = tt.mockApply
			defer func() { applyKubectlManifestFuncSlice = originalApply }()

			ApplySliceConfiguration(config)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
			}
		})
	}
}

func TestVerifyNodeIPsInClusters(t *testing.T) {
	config := &ConfigurationSpecs{
		Configuration: Configuration{
			ClusterConfiguration: ClusterConfiguration{
				ControllerCluster: Cluster{Name: "controller", ContextName: "ctrl-ctx"},
				WorkerClusters:    []Cluster{{Name: "worker-1"}},
			},
			KubeSliceConfiguration: KubeSliceConfiguration{ProjectName: "test-project"},
		},
	}

	tests := []struct {
		name           string
		mockExecutor   func(*util.FakeExecutor)
		expectedSleeps int
	}{
		{
			name: "IPs populated immediately",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("'10.0.0.1'"))
					return nil
				}
			},
			expectedSleeps: 0,
		},
		{
			name: "IPs populated after 3 retries",
			mockExecutor: func(fe *util.FakeExecutor) {
				callCount := 0
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					callCount++
					if callCount < 3 {
						stdout.Write([]byte(""))
					} else {
						stdout.Write([]byte("'10.0.0.1'"))
					}
					return nil
				}
			},
			expectedSleeps: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeClock := util.SystemClock.(*util.FakeClock)
			fakeOutput := util.Output.(*util.FakeOutput)

			tt.mockExecutor(fakeExec)

			verifyNodeIPsInClusters(config)

			assert.Empty(t, fakeOutput.FatalCalls)
			assert.Len(t, fakeClock.SleepCalls, tt.expectedSleeps)
		})
	}
}

func TestGetSliceConfig(t *testing.T) {
	cluster := &Cluster{Name: "controller"}

	tests := []struct {
		name          string
		mockGet       func(string, string, string, *Cluster, string)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful get",
			mockGet: func(resourceType, resourceName, namespace string, c *Cluster, outputFormat string) {
				assert.Equal(t, SliceConfigObject, resourceType)
				assert.Equal(t, "my-slice", resourceName)
			},
		},
		{
			name: "get fails",
			mockGet: func(resourceType, resourceName, namespace string, c *Cluster, outputFormat string) {
				util.Fatalf("Process failed %v", errors.New("get failed"))
			},
			expectFatal:   true,
			fatalContains: "get failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeClock := util.SystemClock.(*util.FakeClock)
			fakeOutput := util.Output.(*util.FakeOutput)

			originalGet := getKubectlResourcesFuncSlice
			getKubectlResourcesFuncSlice = tt.mockGet
			defer func() { getKubectlResourcesFuncSlice = originalGet }()

			GetSliceConfig("my-slice", "test-namespace", cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				require.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			}
		})
	}
}

func TestDeleteSliceConfig(t *testing.T) {
	cluster := &Cluster{Name: "controller"}

	tests := []struct {
		name          string
		mockDelete    func(string, string, string, *Cluster)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful delete",
			mockDelete: func(resourceType, resourceName, namespace string, c *Cluster) {
				assert.Equal(t, SliceConfigObject, resourceType)
			},
		},
		{
			name: "delete fails",
			mockDelete: func(resourceType, resourceName, namespace string, c *Cluster) {
				util.Fatalf("Process failed %v", errors.New("delete failed"))
			},
			expectFatal:   true,
			fatalContains: "delete failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeClock := util.SystemClock.(*util.FakeClock)
			fakeOutput := util.Output.(*util.FakeOutput)

			originalDelete := deleteKubectlResourcesFuncSlice
			deleteKubectlResourcesFuncSlice = tt.mockDelete
			defer func() { deleteKubectlResourcesFuncSlice = originalDelete }()

			DeleteSliceConfig("my-slice", "test-namespace", cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				require.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			}
		})
	}
}

func TestEditSliceConfig(t *testing.T) {
	cluster := &Cluster{Name: "controller"}

	tests := []struct {
		name          string
		mockEdit      func(string, string, string, *Cluster)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful edit",
			mockEdit: func(resourceType, resourceName, namespace string, c *Cluster) {
				assert.Equal(t, SliceConfigObject, resourceType)
			},
		},
		{
			name: "edit fails",
			mockEdit: func(resourceType, resourceName, namespace string, c *Cluster) {
				util.Fatalf("Process failed %v", errors.New("edit failed"))
			},
			expectFatal:   true,
			fatalContains: "edit failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeClock := util.SystemClock.(*util.FakeClock)
			fakeOutput := util.Output.(*util.FakeOutput)

			originalEdit := editKubectlResourcesFuncSlice
			editKubectlResourcesFuncSlice = tt.mockEdit
			defer func() { editKubectlResourcesFuncSlice = originalEdit }()

			EditSliceConfig("my-slice", "test-namespace", cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				require.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			}
		})
	}
}

func TestDescribeSliceConfig(t *testing.T) {
	cluster := &Cluster{Name: "controller"}

	tests := []struct {
		name          string
		mockDescribe  func(string, string, string, *Cluster)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful describe",
			mockDescribe: func(resourceType, resourceName, namespace string, c *Cluster) {
				assert.Equal(t, SliceConfigObject, resourceType)
			},
		},
		{
			name: "describe fails",
			mockDescribe: func(resourceType, resourceName, namespace string, c *Cluster) {
				util.Fatalf("Process failed %v", errors.New("describe failed"))
			},
			expectFatal:   true,
			fatalContains: "describe failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeClock := util.SystemClock.(*util.FakeClock)
			fakeOutput := util.Output.(*util.FakeOutput)

			originalDescribe := describeKubectlResourcesFuncSlice
			describeKubectlResourcesFuncSlice = tt.mockDescribe
			defer func() { describeKubectlResourcesFuncSlice = originalDescribe }()

			DescribeSliceConfig("my-slice", "test-namespace", cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				require.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			}
		})
	}
}

func TestCreateSliceConfig(t *testing.T) {
	cluster := &Cluster{Name: "controller"}

	tests := []struct {
		name        string
		expectFatal bool
	}{
		{name: "create slice config from file"},
		{name: "create fails", expectFatal: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeOutput := util.Output.(*util.FakeOutput)

			originalApply := applyFileFuncSlice
			defer func() { applyFileFuncSlice = originalApply }()

			applyFileFuncSlice = func(fileName, namespace string, cluster *Cluster) {
				if tt.expectFatal {
					util.Fatalf("Process failed %v", errors.New("apply failed"))
				}
			}

			CreateSliceConfig("test-namespace", cluster, "slice-config.yaml")

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), "apply failed")
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
			}
		})
	}
}
