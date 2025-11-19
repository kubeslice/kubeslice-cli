package internal

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallCalico(t *testing.T) {
	tests := []struct {
		name                     string
		clusterConfig            *ClusterConfiguration
		mockExecutor             func(*util.FakeExecutor)
		mockPodVerification      func(string, Cluster, string)
		expectFatal              bool
		expectedPodVerifications int
	}{
		{
			name: "install on fresh clusters",
			clusterConfig: &ClusterConfiguration{
				ControllerCluster: Cluster{
					Name:           "controller",
					ContextName:    "controller-ctx",
					KubeConfigPath: "/path/config",
				},
				WorkerClusters: []Cluster{
					{
						Name:           "worker-1",
						ContextName:    "worker-1-ctx",
						KubeConfigPath: "/path/config",
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stderr.Write([]byte("Error from server (NotFound): namespaces \"calico-system\" not found"))
					return errors.New("not found")
				}
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {
				assert.Equal(t, calicoNamespace, namespace)
			},
			expectedPodVerifications: 2,
		},
		{
			name: "calico already installed on all clusters",
			clusterConfig: &ClusterConfiguration{
				ControllerCluster: Cluster{
					Name:           "controller",
					ContextName:    "controller-ctx",
					KubeConfigPath: "/path/config",
				},
				WorkerClusters: []Cluster{
					{
						Name:           "worker-1",
						ContextName:    "worker-1-ctx",
						KubeConfigPath: "/path/config",
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("NAME            STATUS   AGE\ncalico-system   Active   5d"))
					return nil
				}
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {
				// Verification called for existing installation
			},
			expectedPodVerifications: 2,
		},
		{
			name: "mixed installation states",
			clusterConfig: &ClusterConfiguration{
				ControllerCluster: Cluster{
					Name:           "controller",
					ContextName:    "controller-ctx",
					KubeConfigPath: "/path/config",
				},
				WorkerClusters: []Cluster{
					{
						Name:           "worker-1",
						ContextName:    "worker-1-ctx",
						KubeConfigPath: "/path/config",
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				checkCount := 0
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					checkCount++
					if checkCount == 1 {
						stdout.Write([]byte("calico-system   Active   5d"))
						return nil
					}
					stderr.Write([]byte("NotFound"))
					return errors.New("not found")
				}
			},
			expectedPodVerifications: 2,
		},
		{
			name: "operator prerequisites installation fails",
			clusterConfig: &ClusterConfiguration{
				ControllerCluster: Cluster{
					Name:           "controller",
					ContextName:    "controller-ctx",
					KubeConfigPath: "/path/config",
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("failed to apply operator")
				}
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stderr.Write([]byte("NotFound"))
					return errors.New("not found")
				}
			},
			expectFatal: true,
		},
		{
			name: "custom resources creation fails",
			clusterConfig: &ClusterConfiguration{
				ControllerCluster: Cluster{
					Name:           "controller",
					ContextName:    "controller-ctx",
					KubeConfigPath: "/path/config",
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				callCount := 0
				fe.ExecuteFunc = func(cli string, args ...string) error {
					callCount++
					if callCount == 1 {
						return nil
					}
					return errors.New("failed to create custom resources")
				}
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stderr.Write([]byte("NotFound"))
					return errors.New("not found")
				}
			},
			expectFatal: true,
		},
		{
			name: "multiple worker clusters",
			clusterConfig: &ClusterConfiguration{
				ControllerCluster: Cluster{
					Name:           "controller",
					ContextName:    "controller-ctx",
					KubeConfigPath: "/path/config",
				},
				WorkerClusters: []Cluster{
					{
						Name:           "worker-1",
						ContextName:    "worker-1-ctx",
						KubeConfigPath: "/path/config",
					},
					{
						Name:           "worker-2",
						ContextName:    "worker-2-ctx",
						KubeConfigPath: "/path/config",
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stderr.Write([]byte("NotFound"))
					return errors.New("not found")
				}
			},
			expectedPodVerifications: 3,
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

			podVerificationCallCount := 0
			originalPodVerify := podVerificationFuncCalico
			podVerificationFuncCalico = func(msg string, cluster Cluster, namespace string) {
				podVerificationCallCount++
				if tt.mockPodVerification != nil {
					tt.mockPodVerification(msg, cluster, namespace)
				}
			}
			defer func() { podVerificationFuncCalico = originalPodVerify }()

			InstallCalico(tt.clusterConfig)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Equal(t, tt.expectedPodVerifications, podVerificationCallCount)
			}
		})
	}
}

func TestCalicoAlreadyInstalled(t *testing.T) {
	tests := []struct {
		name                string
		cluster             *Cluster
		mockExecutor        func(*util.FakeExecutor)
		mockPodVerification func(string, Cluster, string)
		expectedInstalled   bool
	}{
		{
			name: "calico namespace exists",
			cluster: &Cluster{
				Name:           "test-cluster",
				ContextName:    "test-ctx",
				KubeConfigPath: "/path/config",
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("NAME            STATUS   AGE\ncalico-system   Active   5d"))
					return nil
				}
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {
				assert.Equal(t, calicoNamespace, namespace)
				assert.Equal(t, "test-cluster", cluster.Name)
			},
			expectedInstalled: true,
		},
		{
			name: "calico namespace not found",
			cluster: &Cluster{
				Name:           "test-cluster",
				ContextName:    "test-ctx",
				KubeConfigPath: "/path/config",
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stderr.Write([]byte("Error from server (NotFound): namespaces \"calico-system\" not found"))
					return errors.New("not found")
				}
			},
			expectedInstalled: false,
		},
		{
			name: "empty output treated as installed",
			cluster: &Cluster{
				Name:           "test-cluster",
				ContextName:    "test-ctx",
				KubeConfigPath: "/path/config",
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte(""))
					return nil
				}
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {
			},
			expectedInstalled: true,
		},
		{
			name: "non-NotFound error treated as installed",
			cluster: &Cluster{
				Name:           "test-cluster",
				ContextName:    "test-ctx",
				KubeConfigPath: "/path/config",
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stderr.Write([]byte("connection refused"))
					return errors.New("connection error")
				}
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {
			},
			expectedInstalled: true,
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

			podVerifyCalled := false
			originalPodVerify := podVerificationFuncCalico
			podVerificationFuncCalico = func(msg string, cluster Cluster, namespace string) {
				podVerifyCalled = true
				if tt.mockPodVerification != nil {
					tt.mockPodVerification(msg, cluster, namespace)
				}
			}
			defer func() { podVerificationFuncCalico = originalPodVerify }()

			result := calicoAlreadyInstalled(tt.cluster)

			assert.Equal(t, tt.expectedInstalled, result)

			if tt.expectedInstalled {
				assert.True(t, podVerifyCalled, "Expected pod verification to be called")
			}
		})
	}
}

func TestInstallCalicoOperatorPrerequisites(t *testing.T) {
	tests := []struct {
		name          string
		cluster       *Cluster
		mockExecutor  func(*util.FakeExecutor)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful prerequisites installation",
			cluster: &Cluster{
				Name:           "test-cluster",
				ContextName:    "test-ctx",
				KubeConfigPath: "/path/config",
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					assert.Equal(t, "kubectl", cli)
					assert.Contains(t, args, "--context=test-ctx")
					assert.Contains(t, args, "--kubeconfig=/path/config")
					assert.Contains(t, args, "create")
					assert.Contains(t, args, "-f")
					assert.Contains(t, args, calicoOperatorURL)
					return nil
				}
			},
		},
		{
			name: "kubectl create fails",
			cluster: &Cluster{
				Name:           "test-cluster",
				ContextName:    "test-ctx",
				KubeConfigPath: "/path/config",
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("unable to apply manifest")
				}
			},
			expectFatal:   true,
			fatalContains: "Process failed",
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

			installCalicoOperatorPrerequisites(tt.cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
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

func TestCreateCalicoOperator(t *testing.T) {
	tests := []struct {
		name          string
		cluster       *Cluster
		mockExecutor  func(*util.FakeExecutor)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful operator creation",
			cluster: &Cluster{
				Name:           "test-cluster",
				ContextName:    "test-ctx",
				KubeConfigPath: "/path/config",
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					assert.Equal(t, "kubectl", cli)
					assert.Contains(t, args, "--context=test-ctx")
					assert.Contains(t, args, "--kubeconfig=/path/config")
					assert.Contains(t, args, "create")
					assert.Contains(t, args, "-f")
					assert.Contains(t, args, calicoCustomResourceURL)
					return nil
				}
			},
		},
		{
			name: "kubectl create fails",
			cluster: &Cluster{
				Name:           "test-cluster",
				ContextName:    "test-ctx",
				KubeConfigPath: "/path/config",
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("failed to create custom resources")
				}
			},
			expectFatal:   true,
			fatalContains: "Process failed",
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

			createCalicoOperator(tt.cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
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
