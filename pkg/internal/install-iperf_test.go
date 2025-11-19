package internal

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallIPerf(t *testing.T) {
	tests := []struct {
		name                   string
		config                 *ConfigurationSpecs
		mockApplyManifest      func(string, string, *Cluster)
		mockPodVerification    func(string, Cluster, string)
		expectFatal            bool
		fatalContains          string
		expectedManifestCalls  int
		expectedPodVerifyCalls int
	}{
		{
			name: "install on two workers (server + client)",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						WorkerClusters: []Cluster{
							{Name: "worker-1", ContextName: "w1-ctx"},
							{Name: "worker-2", ContextName: "w2-ctx"},
						},
					},
				},
			},
			mockApplyManifest: func(manifestPath string, namespace string, cluster *Cluster) {
				assert.Equal(t, iPerfNamespace, namespace)
				if strings.Contains(cluster.Name, "worker-1") {
					assert.Contains(t, manifestPath, iPerfServerFileName, "Should apply server to worker-1")
				}
				if strings.Contains(cluster.Name, "worker-2") {
					assert.Contains(t, manifestPath, iPerfClientFileName, "Should apply client to worker-2")
				}
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {
				assert.Equal(t, iPerfNamespace, namespace)
				if strings.Contains(cluster.Name, "worker-1") {
					assert.Contains(t, msg, "iPerf Server")
				}
				if strings.Contains(cluster.Name, "worker-2") {
					assert.Contains(t, msg, "iPerf Client")
				}
			},
			expectedManifestCalls:  2,
			expectedPodVerifyCalls: 2,
		},
		{
			name: "install on multiple workers (1 server + 2 clients)",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						WorkerClusters: []Cluster{
							{Name: "worker-1"},
							{Name: "worker-2"},
							{Name: "worker-3"},
						},
					},
				},
			},
			mockApplyManifest: func(manifestPath string, namespace string, cluster *Cluster) {
				assert.Equal(t, iPerfNamespace, namespace)
				if cluster.Name == "worker-1" {
					assert.Contains(t, manifestPath, iPerfServerFileName)
				} else {
					assert.Contains(t, manifestPath, iPerfClientFileName)
				}
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {
				assert.Equal(t, iPerfNamespace, namespace)
			},
			expectedManifestCalls:  3,
			expectedPodVerifyCalls: 3,
		},
		{
			name: "single worker - server only",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						WorkerClusters: []Cluster{
							{Name: "worker-1"},
						},
					},
				},
			},
			mockApplyManifest: func(manifestPath string, namespace string, cluster *Cluster) {
				assert.Equal(t, iPerfNamespace, namespace)
				assert.Equal(t, "worker-1", cluster.Name)
				assert.Contains(t, manifestPath, iPerfServerFileName)
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {
				assert.Equal(t, "worker-1", cluster.Name)
				assert.Contains(t, msg, "iPerf Server")
			},
			expectedManifestCalls:  1,
			expectedPodVerifyCalls: 1,
		},
		{
			name: "server manifest apply fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						WorkerClusters: []Cluster{{Name: "worker-1"}},
					},
				},
			},
			mockApplyManifest: func(manifestPath string, namespace string, cluster *Cluster) {
				util.Fatalf("Process failed %v", errors.New("apply failed"))
			},
			expectFatal:            true,
			fatalContains:          "apply failed",
			expectedManifestCalls:  1,
			expectedPodVerifyCalls: 0,
		},
		{
			name: "server pod verification fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						WorkerClusters: []Cluster{{Name: "worker-1"}},
					},
				},
			},
			mockApplyManifest: func(manifestPath string, namespace string, cluster *Cluster) {
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {
				util.Fatalf("Process failed %v", errors.New("pod timeout"))
			},
			expectFatal:            true,
			fatalContains:          "pod timeout",
			expectedManifestCalls:  1,
			expectedPodVerifyCalls: 1,
		},
		{
			name: "client manifest apply fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						WorkerClusters: []Cluster{{Name: "worker-1"}, {Name: "worker-2"}},
					},
				},
			},
			mockApplyManifest: func(manifestPath string, namespace string, cluster *Cluster) {
				if cluster.Name == "worker-2" {
					util.Fatalf("Process failed %v", errors.New("client apply failed"))
				}
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {
				// Server pod verification succeeds
			},
			expectFatal:            true,
			fatalContains:          "client apply failed",
			expectedManifestCalls:  2,
			expectedPodVerifyCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeOutput := util.Output.(*util.FakeOutput)
			fakeClock := util.SystemClock.(*util.FakeClock)

			manifestCallCount := 0
			originalApplyManifest := applyKubectlManifestFunc
			applyKubectlManifestFunc = func(manifestPath string, namespace string, cluster *Cluster) {
				manifestCallCount++
				if tt.mockApplyManifest != nil {
					tt.mockApplyManifest(manifestPath, namespace, cluster)
				}
			}
			defer func() { applyKubectlManifestFunc = originalApplyManifest }()

			podVerifyCallCount := 0
			originalPodVerify := podVerificationFuncIPerf
			podVerificationFuncIPerf = func(msg string, cluster Cluster, namespace string) {
				podVerifyCallCount++
				if tt.mockPodVerification != nil {
					tt.mockPodVerification(msg, cluster, namespace)
				}
			}
			defer func() { podVerificationFuncIPerf = originalPodVerify }()

			InstallIPerf(tt.config)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls, "Expected Fatalf to be called")
				if tt.fatalContains != "" {
					callStr := fmt.Sprint(fakeOutput.FatalCalls[0])
					assert.Contains(t, callStr, tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls, "Expected no Fatalf calls")
				assert.Equal(t, tt.expectedManifestCalls, manifestCallCount, "Manifest apply call count mismatch")
				assert.Equal(t, tt.expectedPodVerifyCalls, podVerifyCallCount, "Pod verification call count mismatch")
				expectedSleepCalls := tt.expectedManifestCalls
				assert.Len(t, fakeClock.SleepCalls, expectedSleepCalls)
				for _, d := range fakeClock.SleepCalls {
					assert.Equal(t, 200*time.Millisecond, d)
				}
			}
		})
	}
}

func TestGenerateIPerfManifests(t *testing.T) {
	cleanup := util.NewTestEnvironment()
	defer cleanup()

	fakeFS := util.FileSystem.(*util.FakeFileSystem)
	fakeClock := util.SystemClock.(*util.FakeClock)

	GenerateIPerfManifests()

	// Verify both files were created
	clientFile := kubesliceDirectory + "/" + iPerfClientFileName
	serverFile := kubesliceDirectory + "/" + iPerfServerFileName

	clientContent, clientExists := fakeFS.WrittenFiles[clientFile]
	serverContent, serverExists := fakeFS.WrittenFiles[serverFile]

	require.True(t, clientExists, "Client manifest should be created")
	require.True(t, serverExists, "Server manifest should be created")

	assert.Contains(t, string(clientContent), "name: iperf-sleep")
	assert.Contains(t, string(clientContent), "kind: Namespace")
	assert.Contains(t, string(serverContent), "name: iperf-server")
	assert.Contains(t, string(serverContent), "containerPort: 5201")
	assert.Len(t, fakeClock.SleepCalls, 2)
	assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
	assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[1])
}

func TestGenerateIPerfServiceExportManifest(t *testing.T) {
	config := &ConfigurationSpecs{
		Configuration: Configuration{
			ClusterConfiguration: ClusterConfiguration{
				WorkerClusters: []Cluster{
					{Name: "worker-1"},
				},
			},
		},
	}

	cleanup := util.NewTestEnvironment()
	defer cleanup()

	fakeFS := util.FileSystem.(*util.FakeFileSystem)
	fakeClock := util.SystemClock.(*util.FakeClock)

	GenerateIPerfServiceExportManifest(config)

	exportFile := kubesliceDirectory + "/" + iPerfServerServiceExportFileName
	content, exists := fakeFS.WrittenFiles[exportFile]

	require.True(t, exists, "Service export manifest should be created")

	contentStr := string(content)
	assert.Contains(t, contentStr, "kind: ServiceExport")
	assert.Contains(t, contentStr, "name: iperf-server")
	assert.Contains(t, contentStr, "networking.kubeslice.io/v1beta1")
	assert.Contains(t, contentStr, "slice: demo")
	assert.Len(t, fakeClock.SleepCalls, 1)
	assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
}

func TestApplyIPerfServiceExportManifest(t *testing.T) {
	config := &ConfigurationSpecs{
		Configuration: Configuration{
			ClusterConfiguration: ClusterConfiguration{
				WorkerClusters: []Cluster{
					{Name: "worker-1", ContextName: "w1-ctx"},
					{Name: "worker-2", ContextName: "w2-ctx"},
				},
			},
		},
	}

	cleanup := util.NewTestEnvironment()
	defer cleanup()

	applyCalled := false
	originalApplyManifest := applyKubectlManifestFunc
	applyKubectlManifestFunc = func(manifestPath string, namespace string, cluster *Cluster) {
		applyCalled = true
		assert.Contains(t, manifestPath, iPerfServerServiceExportFileName)
		assert.Equal(t, iPerfNamespace, namespace)
		assert.Equal(t, "worker-1", cluster.Name, "ServiceExport should only be applied to worker-1")
	}
	defer func() { applyKubectlManifestFunc = originalApplyManifest }()

	ApplyIPerfServiceExportManifest(config)

	assert.True(t, applyCalled, "ApplyKubectlManifest should be called")
}

func TestRolloutRestartIPerf(t *testing.T) {
	tests := []struct {
		name          string
		config        *ConfigurationSpecs
		mockExecutor  func(*util.FakeExecutor)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful rollout restart (1 server, 2 clients)",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
						WorkerClusters: []Cluster{
							{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path"},
							{Name: "worker-2", ContextName: "w2-ctx", KubeConfigPath: "/path"},
							{Name: "worker-3", ContextName: "w3-ctx", KubeConfigPath: "/path"},
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				callCount := 0
				fe.ExecuteFunc = func(cli string, args ...string) error {
					callCount++
					assert.Equal(t, "kubectl", cli)
					assert.Contains(t, args, "rollout")
					assert.Contains(t, args, "restart")
					assert.Contains(t, args, "-n")
					assert.Contains(t, args, iPerfNamespace)
					assert.Contains(t, args, "--kubeconfig=/path")

					argsStr := strings.Join(args, " ")
					if callCount == 1 {
						assert.Contains(t, argsStr, "deployment/iperf-server")
						assert.Contains(t, argsStr, "--context=w1-ctx")
					} else {
						assert.Contains(t, argsStr, "deployment/iperf-sleep")
						if callCount == 2 {
							assert.Contains(t, argsStr, "--context=w2-ctx")
						} else if callCount == 3 {
							assert.Contains(t, argsStr, "--context=w3-ctx")
						}
					}
					return nil
				}
			},
		},
		{
			name: "server restart fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
						WorkerClusters: []Cluster{
							{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path"},
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("deployment not found")
				}
			},
			expectFatal:   true,
			fatalContains: "deployment not found",
		},
		{
			name: "client restart fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
						WorkerClusters: []Cluster{
							{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path"},
							{Name: "worker-2", ContextName: "w2-ctx", KubeConfigPath: "/path"},
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				callCount := 0
				fe.ExecuteFunc = func(cli string, args ...string) error {
					callCount++
					if callCount == 1 {
						return nil
					}
					return errors.New("client deployment not found")
				}
			},
			expectFatal:   true,
			fatalContains: "client deployment not found",
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

			RolloutRestartIPerf(tt.config)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					callStr := fmt.Sprint(fakeOutput.FatalCalls[0])
					assert.Contains(t, callStr, tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				expectedCalls := len(tt.config.Configuration.ClusterConfiguration.WorkerClusters)
				require.Len(t, fakeExec.Calls, expectedCalls)
			}
		})
	}
}
