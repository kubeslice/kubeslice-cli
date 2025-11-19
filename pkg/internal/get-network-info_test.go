package internal

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherNetworkInformation(t *testing.T) {
	tests := []struct {
		name               string
		config             *ConfigurationSpecs
		mockExecutor       func(*util.FakeExecutor)
		mockGetAllClusters func(*ClusterConfiguration) []*Cluster
		expectFatal        bool
		validateClusters   func(*testing.T, *ConfigurationSpecs)
	}{
		{
			name: "kind cluster network setup",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ClusterType: "kind",
						ControllerCluster: Cluster{
							Name: "controller",
						},
						WorkerClusters: []Cluster{
							{Name: "worker-1"},
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					if cli == "docker" {
						stdout.Write([]byte("172.18.0.2"))
					}
					return nil
				}
			},
			mockGetAllClusters: func(cc *ClusterConfiguration) []*Cluster {
				return []*Cluster{&cc.ControllerCluster, &cc.WorkerClusters[0]}
			},
			validateClusters: func(t *testing.T, config *ConfigurationSpecs) {
				cc := &config.Configuration.ClusterConfiguration
				assert.Equal(t, "172.18.0.2", cc.ControllerCluster.NodeIP)
				assert.Equal(t, "https://172.18.0.2:6443", cc.ControllerCluster.ControlPlaneAddress)
				assert.Equal(t, "172.18.0.2", cc.WorkerClusters[0].NodeIP)
			},
		},
		{
			name: "profile set triggers kind-like behavior",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ClusterType: "eks",
						Profile:     ProfileEntDemo,
						ControllerCluster: Cluster{
							Name: "controller",
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					if cli == "docker" {
						stdout.Write([]byte("172.18.0.3"))
					}
					return nil
				}
			},
			mockGetAllClusters: func(cc *ClusterConfiguration) []*Cluster {
				return []*Cluster{&cc.ControllerCluster}
			},
			validateClusters: func(t *testing.T, config *ConfigurationSpecs) {
				cc := &config.Configuration.ClusterConfiguration
				assert.Equal(t, "172.18.0.3", cc.ControllerCluster.NodeIP)
				assert.Equal(t, "https://172.18.0.3:6443", cc.ControllerCluster.ControlPlaneAddress)
			},
		},
		{
			name: "non-kind cluster network setup",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ClusterType: "eks",
						Profile:     "",
						ControllerCluster: Cluster{
							Name:           "controller",
							ContextName:    "controller-ctx",
							KubeConfigPath: "/path/config",
						},
						WorkerClusters: []Cluster{
							{
								Name:           "worker-1",
								ContextName:    "worker-ctx",
								KubeConfigPath: "/path/config",
							},
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					if cli == "kubectl" {
						argsStr := strings.Join(args, " ")
						if strings.Contains(argsStr, "config view") {
							stdout.Write([]byte("https://api.cluster.example.com:6443"))
						} else if strings.Contains(argsStr, "get nodes") {
							stdout.Write([]byte("ExternalIP=1.2.3.4\nInternalIP=10.0.0.1"))
						}
					}
					return nil
				}
			},
			mockGetAllClusters: func(cc *ClusterConfiguration) []*Cluster {
				return []*Cluster{&cc.ControllerCluster, &cc.WorkerClusters[0]}
			},
			validateClusters: func(t *testing.T, config *ConfigurationSpecs) {
				cc := &config.Configuration.ClusterConfiguration
				assert.Equal(t, "https://api.cluster.example.com:6443", cc.ControllerCluster.ControlPlaneAddress)
				assert.Equal(t, "1.2.3.4", cc.ControllerCluster.NodeIP)
			},
		},
		{
			name: "docker command fails for kind cluster",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ClusterType: "kind",
						ControllerCluster: Cluster{
							Name: "controller",
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					return errors.New("docker not running")
				}
			},
			mockGetAllClusters: func(cc *ClusterConfiguration) []*Cluster {
				return []*Cluster{&cc.ControllerCluster}
			},
			expectFatal: true,
		},
		{
			name: "kubectl command fails for non-kind cluster",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ClusterType: "eks",
						ControllerCluster: Cluster{
							Name:           "controller",
							ContextName:    "ctx",
							KubeConfigPath: "/path",
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					return errors.New("kubectl error")
				}
			},
			mockGetAllClusters: func(cc *ClusterConfiguration) []*Cluster {
				return []*Cluster{&cc.ControllerCluster}
			},
			expectFatal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeOutput := util.Output.(*util.FakeOutput)
			fakeClock := util.SystemClock.(*util.FakeClock)

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			if tt.mockGetAllClusters != nil {
				originalGetAllClusters := getAllClustersFunc
				getAllClustersFunc = tt.mockGetAllClusters
				defer func() { getAllClustersFunc = originalGetAllClusters }()
			}

			GatherNetworkInformation(tt.config)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				if tt.validateClusters != nil {
					tt.validateClusters(t, tt.config)
				}

				if tt.config.Configuration.ClusterConfiguration.ClusterType == "kind" ||
					tt.config.Configuration.ClusterConfiguration.Profile != "" {
					assert.NotEmpty(t, fakeClock.SleepCalls)
					for _, d := range fakeClock.SleepCalls {
						assert.Equal(t, 200*time.Millisecond, d)
					}
				}
			}
		})
	}
}

func TestRunDockerInspectForNodeIP(t *testing.T) {
	tests := []struct {
		name          string
		clusterName   string
		mockExecutor  func(*util.FakeExecutor)
		expectedIP    string
		expectFatal   bool
		fatalContains string
	}{
		{
			name:        "successful IP retrieval",
			clusterName: "test-cluster",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					assert.Equal(t, "docker", cli)
					assert.Contains(t, args, "inspect")
					assert.Contains(t, args, "test-cluster-control-plane")
					stdout.Write([]byte("172.18.0.5\n"))
					return nil
				}
			},
			expectedIP: "172.18.0.5",
		},
		{
			name:        "IP with whitespace trimmed",
			clusterName: "cluster",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("  192.168.1.100  \n"))
					return nil
				}
			},
			expectedIP: "192.168.1.100",
		},
		{
			name:        "empty output",
			clusterName: "cluster",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte(""))
					return nil
				}
			},
			expectedIP: "",
		},
		{
			name:        "docker command fails",
			clusterName: "cluster",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stderr.Write([]byte("container not found"))
					return errors.New("exit code 1")
				}
			},
			expectFatal:   true,
			fatalContains: "Failed to run command",
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

			ip := runDockerInspectForNodeIP(tt.clusterName)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					callStr := fmt.Sprint(fakeOutput.FatalCalls[0])
					assert.Contains(t, callStr, tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Equal(t, tt.expectedIP, ip)
			}
		})
	}
}

func TestGetControlPlaneAddress(t *testing.T) {
	tests := []struct {
		name            string
		cluster         *Cluster
		mockExecutor    func(*util.FakeExecutor)
		expectedAddress string
		expectFatal     bool
		fatalContains   string
	}{
		{
			name: "successful address retrieval",
			cluster: &Cluster{
				Name:           "test",
				ContextName:    "test-ctx",
				KubeConfigPath: "/path/config",
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					assert.Equal(t, "kubectl", cli)
					assert.Contains(t, args, "--context=test-ctx")
					assert.Contains(t, args, "--kubeconfig=/path/config")
					stdout.Write([]byte("https://api.example.com:6443"))
					return nil
				}
			},
			expectedAddress: "https://api.example.com:6443",
		},
		{
			name: "empty address returned",
			cluster: &Cluster{
				Name:           "test",
				ContextName:    "test-ctx",
				KubeConfigPath: "/path/config",
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte(""))
					return nil
				}
			},
			expectedAddress: "",
		},
		{
			name: "kubectl command fails",
			cluster: &Cluster{
				Name:           "test",
				ContextName:    "test-ctx",
				KubeConfigPath: "/path/config",
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stderr.Write([]byte("context not found"))
					return errors.New("context not found")
				}
			},
			expectFatal:   true,
			fatalContains: "Failed to run command",
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

			address := _getControlPlaneAddress(tt.cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					callStr := fmt.Sprint(fakeOutput.FatalCalls[0])
					assert.Contains(t, callStr, tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Equal(t, tt.expectedAddress, address)
			}
		})
	}
}

func TestSetControlPlaneAddress(t *testing.T) {
	tests := []struct {
		name               string
		clusterConfig      *ClusterConfiguration
		mockExecutor       func(*util.FakeExecutor)
		mockGetAllClusters func(*ClusterConfiguration) []*Cluster
		validateClusters   func(*testing.T, *ClusterConfiguration)
	}{
		{
			name: "skips cluster with existing control plane address",
			clusterConfig: &ClusterConfiguration{
				ControllerCluster: Cluster{
					Name:                "controller",
					ContextName:         "ctx",
					KubeConfigPath:      "/path",
					ControlPlaneAddress: "https://existing.com:6443",
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					t.Fatal("Should not call kubectl when address already exists")
					return nil
				}
			},
			mockGetAllClusters: func(cc *ClusterConfiguration) []*Cluster {
				return []*Cluster{&cc.ControllerCluster}
			},
			validateClusters: func(t *testing.T, cc *ClusterConfiguration) {
				assert.Equal(t, "https://existing.com:6443", cc.ControllerCluster.ControlPlaneAddress)
			},
		},
		{
			name: "fetches address for cluster without one",
			clusterConfig: &ClusterConfiguration{
				ControllerCluster: Cluster{
					Name:           "controller",
					ContextName:    "ctx",
					KubeConfigPath: "/path",
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("https://new.com:6443"))
					return nil
				}
			},
			mockGetAllClusters: func(cc *ClusterConfiguration) []*Cluster {
				return []*Cluster{&cc.ControllerCluster}
			},
			validateClusters: func(t *testing.T, cc *ClusterConfiguration) {
				assert.Equal(t, "https://new.com:6443", cc.ControllerCluster.ControlPlaneAddress)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			if tt.mockGetAllClusters != nil {
				originalGetAllClusters := getAllClustersFunc
				getAllClustersFunc = tt.mockGetAllClusters
				defer func() { getAllClustersFunc = originalGetAllClusters }()
			}

			setControlPlaneAddress(tt.clusterConfig)

			if tt.validateClusters != nil {
				tt.validateClusters(t, tt.clusterConfig)
			}
		})
	}
}

func TestSetNodeIP(t *testing.T) {
	tests := []struct {
		name               string
		clusterConfig      *ClusterConfiguration
		mockExecutor       func(*util.FakeExecutor)
		mockGetAllClusters func(*ClusterConfiguration) []*Cluster
		validateClusters   func(*testing.T, *ClusterConfiguration)
	}{
		{
			name: "skips cluster with existing node IP",
			clusterConfig: &ClusterConfiguration{
				ControllerCluster: Cluster{
					Name:           "controller",
					ContextName:    "ctx",
					KubeConfigPath: "/path",
					NodeIP:         "1.2.3.4",
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					t.Fatal("Should not call kubectl when IP already exists")
					return nil
				}
			},
			mockGetAllClusters: func(cc *ClusterConfiguration) []*Cluster {
				return []*Cluster{&cc.ControllerCluster}
			},
			validateClusters: func(t *testing.T, cc *ClusterConfiguration) {
				assert.Equal(t, "1.2.3.4", cc.ControllerCluster.NodeIP)
			},
		},
		{
			name: "fetches IP for cluster without one",
			clusterConfig: &ClusterConfiguration{
				ControllerCluster: Cluster{
					Name:           "controller",
					ContextName:    "ctx",
					KubeConfigPath: "/path",
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("ExternalIP=5.6.7.8\nInternalIP=10.0.0.1"))
					return nil
				}
			},
			mockGetAllClusters: func(cc *ClusterConfiguration) []*Cluster {
				return []*Cluster{&cc.ControllerCluster}
			},
			validateClusters: func(t *testing.T, cc *ClusterConfiguration) {
				assert.Equal(t, "5.6.7.8", cc.ControllerCluster.NodeIP)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			if tt.mockGetAllClusters != nil {
				originalGetAllClusters := getAllClustersFunc
				getAllClustersFunc = tt.mockGetAllClusters
				defer func() { getAllClustersFunc = originalGetAllClusters }()
			}

			setNodeIP(tt.clusterConfig)

			if tt.validateClusters != nil {
				tt.validateClusters(t, tt.clusterConfig)
			}
		})
	}
}
