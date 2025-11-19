package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateKindClusters(t *testing.T) {
	tests := []struct {
		name                  string
		config                *ConfigurationSpecs
		mockExecutor          func(*util.FakeExecutor, *[]string)
		expectCreatedClusters []string
		expectedSleeps        int
		expectFatal           bool
		fatalContains         string
		clustersToGetFromKind string
	}{
		{
			name: "create new clusters (controller + 2 workers)",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
						WorkerClusters: []Cluster{
							{Name: "worker-1"},
							{Name: "worker-2"},
						},
					},
				},
			},
			clustersToGetFromKind: "",
			mockExecutor: func(fe *util.FakeExecutor, created *[]string) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					assert.Equal(t, "kind", cli)
					require.Len(t, args, 3, "Expected 3 arguments for 'kind create cluster'")
					assert.Equal(t, "create", args[0])
					assert.Equal(t, "cluster", args[1])
					configArg := args[2]
					assert.True(t, strings.HasPrefix(configArg, "--config="), "Expected --config= flag")

					parts := strings.Split(configArg, "/")
					fileNameWithExt := parts[len(parts)-1]
					fileName := strings.TrimSuffix(fileNameWithExt, ".yaml")
					*created = append(*created, fileName)
					return nil
				}
			},
			expectCreatedClusters: []string{"controller", "worker-1", "worker-2"},
			expectedSleeps:        3,
		},
		{
			name: "clusters already exist",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
						WorkerClusters: []Cluster{
							{Name: "worker-1"},
						},
					},
				},
			},
			clustersToGetFromKind: "controller\nworker-1\n",
			mockExecutor:          func(fe *util.FakeExecutor, created *[]string) {},
			expectCreatedClusters: []string{},
			expectedSleeps:        0,
		},
		{
			name: "partial cluster exists (creates 2 workers)",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
						WorkerClusters: []Cluster{
							{Name: "worker-1"},
							{Name: "worker-2"},
						},
					},
				},
			},
			clustersToGetFromKind: "controller\n",
			mockExecutor: func(fe *util.FakeExecutor, created *[]string) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					assert.Equal(t, "kind", cli)
					require.Len(t, args, 3, "Expected 3 arguments for 'kind create cluster'")
					configArg := args[2]
					parts := strings.Split(configArg, "/")
					fileNameWithExt := parts[len(parts)-1]
					fileName := strings.TrimSuffix(fileNameWithExt, ".yaml")
					*created = append(*created, fileName)
					return nil
				}
			},
			expectCreatedClusters: []string{"worker-1", "worker-2"},
			expectedSleeps:        2,
		},
		{
			name: "kind get clusters fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
					},
				},
			},
			clustersToGetFromKind: "",
			mockExecutor: func(fe *util.FakeExecutor, created *[]string) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					return errors.New("kind not found")
				}
			},
			expectFatal:   true,
			fatalContains: "kind not found",
		},
		{
			name: "kind create cluster fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
					},
				},
			},
			clustersToGetFromKind: "",
			mockExecutor: func(fe *util.FakeExecutor, created *[]string) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("failed to create cluster")
				}
			},
			expectFatal:    true,
			fatalContains:  "failed to create cluster",
			expectedSleeps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeOutput := util.Output.(*util.FakeOutput)
			fakeClock := util.SystemClock.(*util.FakeClock)

			var createdClusters []string

			fakeExec.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
				assert.Equal(t, "kind", cli)
				assert.Equal(t, "get", args[0])
				assert.Equal(t, "clusters", args[1])
				stdout.Write([]byte(tt.clustersToGetFromKind))
				return nil
			}

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec, &createdClusters)
			}

			CreateKindClusters(tt.config)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls, "Expected a fatal error")
				if tt.fatalContains != "" {
					callStr := fmt.Sprint(fakeOutput.FatalCalls[0])
					assert.Contains(t, callStr, tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls, "Expected no fatal errors")

				assert.ElementsMatch(t, tt.expectCreatedClusters, createdClusters, "Mismatch in clusters created")

				assert.Len(t, fakeClock.SleepCalls, tt.expectedSleeps)
				for _, d := range fakeClock.SleepCalls {
					assert.Equal(t, 200*time.Millisecond, d)
				}
			}
		})
	}
}

func TestSetKubeConfigPath(t *testing.T) {
	tests := []struct {
		name       string
		mockSetEnv func(string, string) error
		expectWarn bool
	}{
		{
			name: "successful env set",
			mockSetEnv: func(key, value string) error {
				assert.Equal(t, "KUBECONFIG", key)
				assert.Equal(t, KubeconfigPath, value)
				return nil
			},
			expectWarn: false,
		},
		{
			name: "env set fails",
			mockSetEnv: func(key, value string) error {
				return errors.New("permission denied")
			},
			expectWarn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeOutput := util.Output.(*util.FakeOutput)

			originalSetEnv := setEnvFunc
			setEnvFunc = tt.mockSetEnv
			defer func() { setEnvFunc = originalSetEnv }()

			SetKubeConfigPath()

			if tt.expectWarn {
				require.NotEmpty(t, fakeOutput.InfoCalls)
				allOutput := strings.Join(fakeOutput.InfoCalls, " ")
				assert.Contains(t, allOutput, "Warning: Failed to set KUBECONFIG")
			} else {
				assert.Empty(t, fakeOutput.InfoCalls)
			}
		})
	}
}

func TestCreateKubeConfig(t *testing.T) {
	tests := []struct {
		name           string
		fileExists     bool
		expectCreation bool
	}{
		{
			name:           "file does not exist - create it",
			fileExists:     false,
			expectCreation: true,
		},
		{
			name:           "file already exists - skip creation",
			fileExists:     true,
			expectCreation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeFS := util.FileSystem.(*util.FakeFileSystem)
			fakeClock := util.SystemClock.(*util.FakeClock)

			fakeFS.StatFunc = func(name string) (os.FileInfo, error) {
				assert.Equal(t, KubeconfigPath, name)
				if tt.fileExists {
					return nil, nil
				}
				return nil, os.ErrNotExist
			}

			CreateKubeConfig()

			if tt.expectCreation {
				_, exists := fakeFS.WrittenFiles[KubeconfigPath]
				assert.True(t, exists, "Expected kubeconfig file to be created")
				assert.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			} else {
				_, exists := fakeFS.WrittenFiles[KubeconfigPath]
				assert.False(t, exists, "Should not create file if it already exists")
				assert.Empty(t, fakeClock.SleepCalls, "Should not sleep if file exists")
			}
		})
	}
}

func TestGetExistingClusters(t *testing.T) {
	tests := []struct {
		name             string
		clusters         []*Cluster
		mockExecutor     func(*util.FakeExecutor)
		expectedExisting []bool
		expectFatal      bool
		fatalContains    string
	}{
		{
			name: "all clusters exist",
			clusters: []*Cluster{
				{Name: "controller"},
				{Name: "worker-1"},
				{Name: "worker-2"},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("controller\nworker-1\nworker-2\n"))
					return nil
				}
			},
			expectedExisting: []bool{true, true, true},
		},
		{
			name: "no clusters exist",
			clusters: []*Cluster{
				{Name: "controller"},
				{Name: "worker-1"},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte(""))
					return nil
				}
			},
			expectedExisting: []bool{false, false},
		},
		{
			name: "some clusters exist",
			clusters: []*Cluster{
				{Name: "controller"},
				{Name: "worker-1"},
				{Name: "worker-2"},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("controller\nworker-2\n"))
					return nil
				}
			},
			expectedExisting: []bool{true, false, true},
		},
		{
			name: "kind command fails",
			clusters: []*Cluster{
				{Name: "controller"},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					return errors.New("kind not installed")
				}
			},
			expectFatal:   true,
			fatalContains: "kind not installed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeOutput := util.Output.(*util.FakeOutput)

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			result := getExistingClusters(tt.clusters)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					callStr := fmt.Sprint(fakeOutput.FatalCalls[0])
					assert.Contains(t, callStr, tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Equal(t, tt.expectedExisting, result)
			}
		})
	}
}

func TestDeleteKindClusters(t *testing.T) {
	tests := []struct {
		name                  string
		config                *ConfigurationSpecs
		mockExecutor          func(*util.FakeExecutor, *[]string)
		expectDeleted         []string
		expectFatal           bool
		fatalContains         string
		clustersToGetFromKind string
	}{
		{
			name: "delete existing clusters",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
						WorkerClusters: []Cluster{
							{Name: "worker-1"},
						},
					},
				},
			},
			clustersToGetFromKind: "controller\nworker-1\nworker-2\n",
			mockExecutor: func(fe *util.FakeExecutor, deleted *[]string) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					assert.Equal(t, "kind", cli)
					assert.Equal(t, "delete", args[0])
					assert.Equal(t, "clusters", args[1])
					*deleted = append(*deleted, args[2:]...)
					return nil
				}
			},
			expectDeleted: []string{"controller", "worker-1"},
		},
		{
			name: "no clusters to delete",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
					},
				},
			},
			clustersToGetFromKind: "",
			mockExecutor:          func(fe *util.FakeExecutor, deleted *[]string) {},
			expectDeleted:         nil,
		},
		{
			name: "deletion fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
					},
				},
			},
			clustersToGetFromKind: "controller\n",
			mockExecutor: func(fe *util.FakeExecutor, deleted *[]string) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("failed to delete")
				}
			},
			expectFatal:   true,
			fatalContains: "failed to delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeOutput := util.Output.(*util.FakeOutput)

			var deletedClusters []string

			fakeExec.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
				stdout.Write([]byte(tt.clustersToGetFromKind))
				return nil
			}

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec, &deletedClusters)
			}

			DeleteKindClusters(tt.config)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				callStr := fmt.Sprint(fakeOutput.FatalCalls[0])
				assert.Contains(t, callStr, tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Equal(t, tt.expectDeleted, deletedClusters)
				if tt.expectDeleted == nil {
					allOutput := strings.Join(fakeOutput.InfoCalls, " ")
					assert.Contains(t, allOutput, "No Kind Clusters found for deletion")
				}
			}
		})
	}
}

func TestGetAllClusters(t *testing.T) {
	tests := []struct {
		name          string
		config        *ClusterConfiguration
		expectedCount int
		expectedNames []string
	}{
		{
			name: "controller and workers",
			config: &ClusterConfiguration{
				ControllerCluster: Cluster{Name: "controller"},
				WorkerClusters: []Cluster{
					{Name: "worker-1"},
					{Name: "worker-2"},
				},
			},
			expectedCount: 3,
			expectedNames: []string{"controller", "worker-1", "worker-2"},
		},
		{
			name: "controller only",
			config: &ClusterConfiguration{
				ControllerCluster: Cluster{Name: "controller"},
				WorkerClusters:    []Cluster{},
			},
			expectedCount: 1,
			expectedNames: []string{"controller"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusters := getAllClusters(tt.config)

			assert.Len(t, clusters, tt.expectedCount)

			names := make([]string, len(clusters))
			for i, cluster := range clusters {
				names[i] = cluster.Name
			}
			assert.Equal(t, tt.expectedNames, names)
		})
	}
}

func TestGetControllerCluster(t *testing.T) {
	config := ClusterConfiguration{
		ControllerCluster: Cluster{
			Name:        "my-controller",
			ContextName: "my-ctx",
		},
		WorkerClusters: []Cluster{
			{Name: "worker"},
		},
	}

	controller := getControllerCluster(config)

	require.NotNil(t, controller)
	assert.Equal(t, "my-controller", controller.Name)
	assert.Equal(t, "my-ctx", controller.ContextName)
}
