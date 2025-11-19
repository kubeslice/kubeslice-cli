package internal

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstallCertManager tests the cert-manager installation workflow
func TestInstallCertManager(t *testing.T) {
	tests := []struct {
		name          string
		config        *ConfigurationSpecs
		mockExecutor  func(*util.FakeExecutor)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful installation without version",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "test-controller",
							ContextName:    "test-context",
							KubeConfigPath: "/path/to/kubeconfig",
						},
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						CertManagerChart: HelmChart{
							ChartName: "cert-manager",
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
		},
		{
			name: "successful installation with version",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "test-controller",
							ContextName:    "test-context",
							KubeConfigPath: "/path/to/kubeconfig",
						},
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						CertManagerChart: HelmChart{
							ChartName: "cert-manager",
							Version:   "v1.13.0",
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
		},
		{
			name: "helm command fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "test-controller",
							ContextName:    "test-context",
							KubeConfigPath: "/path/to/kubeconfig",
						},
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						CertManagerChart: HelmChart{
							ChartName: "cert-manager",
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					if cli == "helm" {
						return errors.New("helm upgrade failed")
					}
					return nil
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
			fakeClock := util.SystemClock.(*util.FakeClock)

			podVerifyCalled := false
			var capturedNamespace string
			originalPodVerify := podVerificationFunc
			podVerificationFunc = func(msg string, cluster Cluster, namespace string) {
				podVerifyCalled = true
				capturedNamespace = namespace
			}
			defer func() { podVerificationFunc = originalPodVerify }()

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			InstallCertManager(tt.config)

			if tt.expectFatal {
				assert.NotEmpty(t, fakeOutput.FatalCalls, "Expected fatal call")
				if tt.fatalContains != "" {
					assert.Contains(t, fakeOutput.FatalCalls[0], tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls, "Unexpected fatal calls")

				require.Len(t, fakeExec.Calls, 1, "Expected exactly one helm command")
				require.Equal(t, "helm", fakeExec.Calls[0].CLI, "Expected helm command")

				args := fakeExec.Calls[0].Args
				assert.Contains(t, args, "--kube-context")
				assert.Contains(t, args, "test-context")
				assert.Contains(t, args, "--kubeconfig")
				assert.Contains(t, args, "/path/to/kubeconfig")
				assert.Contains(t, args, "upgrade")
				assert.Contains(t, args, "-i")
				assert.Contains(t, args, "cert-manager")
				assert.Contains(t, args, "kubeslice/cert-manager")
				assert.Contains(t, args, "--namespace")
				assert.Contains(t, args, "cert-manager")
				assert.Contains(t, args, "--create-namespace")
				assert.Contains(t, args, "--set")
				assert.Contains(t, args, "installCRDs=true")

				if tt.config.Configuration.HelmChartConfiguration.CertManagerChart.Version != "" {
					expectedVersion := tt.config.Configuration.HelmChartConfiguration.CertManagerChart.Version
					assert.True(t, containsSequence(args, []string{"--version", expectedVersion}),
						"Expected --version %s in helm command", expectedVersion)
				}

				assert.True(t, podVerifyCalled, "Expected pod verification to be called")
				assert.Equal(t, "cert-manager", capturedNamespace, "Expected pod verification for cert-manager namespace")
				assert.Contains(t, fakeClock.SleepCalls, 200*time.Millisecond, "Expected 200ms sleep call")
			}
		})
	}
}

// TestUninstallCertManager tests cert-manager uninstallation
func TestUninstallCertManager(t *testing.T) {
	tests := []struct {
		name          string
		config        *ConfigurationSpecs
		mockExecutor  func(*util.FakeExecutor)
		expectSuccess bool
	}{
		{
			name: "successful uninstallation",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "test-controller",
							ContextName:    "test-context",
							KubeConfigPath: "/path/to/kubeconfig",
						},
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						CertManagerChart: HelmChart{
							ChartName: "cert-manager",
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			expectSuccess: true,
		},
		{
			name: "uninstallation fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "test-controller",
							ContextName:    "test-context",
							KubeConfigPath: "/path/to/kubeconfig",
						},
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						CertManagerChart: HelmChart{
							ChartName: "cert-manager",
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("uninstall failed")
				}
			},
			expectSuccess: false,
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

			UninstallCertManager(tt.config)

			require.Len(t, fakeExec.Calls, 1, "Expected exactly one helm command")
			require.Equal(t, "helm", fakeExec.Calls[0].CLI, "Expected helm command")

			args := fakeExec.Calls[0].Args
			assert.Contains(t, args, "--kube-context")
			assert.Contains(t, args, "test-context")
			assert.Contains(t, args, "--kubeconfig")
			assert.Contains(t, args, "/path/to/kubeconfig")
			assert.Contains(t, args, "uninstall")
			assert.Contains(t, args, "cert-manager")
			assert.Contains(t, args, "--namespace")
			assert.Contains(t, args, "cert-manager")

			infoMessages := fmt.Sprint(fakeOutput.InfoCalls)
			if tt.expectSuccess {
				assert.Contains(t, infoMessages, "Successfully uninstalled cert manager")
			} else {
				assert.Contains(t, infoMessages, "Failed to uninstall cert manager")
			}
		})
	}
}
