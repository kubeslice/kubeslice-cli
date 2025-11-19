package internal

import (
	"errors"
	"testing"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallKubeSliceController(t *testing.T) {
	tests := []struct {
		name                      string
		config                    *ConfigurationSpecs
		mockExecutor              func(*util.FakeExecutor)
		mockGenerateValuesFile    func(string, *HelmChart, string) error
		expectFatal               bool
		fatalContains             string
		expectLicenseVerification bool
		validateValuesContent     bool
	}{
		{
			name: "successful installation without version",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:                "controller",
							ContextName:         "controller-context",
							KubeConfigPath:      "/path/to/config",
							ControlPlaneAddress: "https://controller.example.com",
						},
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						ControllerChart: HelmChart{
							ChartName: "kubeslice-controller",
						},
						ImagePullSecret: ImagePullSecrets{},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			mockGenerateValuesFile: func(filename string, chart *HelmChart, content string) error {
				return nil
			},
			expectLicenseVerification: false,
			validateValuesContent:     true,
		},
		{
			name: "successful installation with version",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:                "controller",
							ContextName:         "controller-context",
							KubeConfigPath:      "/path/to/config",
							ControlPlaneAddress: "https://controller.example.com",
						},
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						ControllerChart: HelmChart{
							ChartName: "kubeslice-controller",
							Version:   "v1.0.0",
						},
						ImagePullSecret: ImagePullSecrets{},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			mockGenerateValuesFile: func(filename string, chart *HelmChart, content string) error {
				return nil
			},
			expectLicenseVerification: false,
			validateValuesContent:     true,
		},
		{
			name: "successful installation with enterprise profile",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:                "controller",
							ContextName:         "controller-context",
							KubeConfigPath:      "/path/to/config",
							ControlPlaneAddress: "https://controller.example.com",
						},
						Profile: ProfileEntDemo,
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						ControllerChart: HelmChart{
							ChartName: "kubeslice-controller",
						},
						ImagePullSecret: ImagePullSecrets{},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			mockGenerateValuesFile: func(filename string, chart *HelmChart, content string) error {
				return nil
			},
			expectLicenseVerification: true,
			validateValuesContent:     true,
		},
		{
			name: "helm installation fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:                "controller",
							ContextName:         "controller-context",
							KubeConfigPath:      "/path/to/config",
							ControlPlaneAddress: "https://controller.example.com",
						},
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						ControllerChart: HelmChart{
							ChartName: "kubeslice-controller",
						},
						ImagePullSecret: ImagePullSecrets{},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("helm upgrade failed")
				}
			},
			mockGenerateValuesFile: func(filename string, chart *HelmChart, content string) error {
				return nil
			},
			expectFatal:   true,
			fatalContains: "Process failed",
		},
		{
			name: "values file generation fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:                "controller",
							ContextName:         "controller-context",
							KubeConfigPath:      "/path/to/config",
							ControlPlaneAddress: "https://controller.example.com",
						},
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						ControllerChart: HelmChart{
							ChartName: "kubeslice-controller",
						},
						ImagePullSecret: ImagePullSecrets{},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			mockGenerateValuesFile: func(filename string, chart *HelmChart, content string) error {
				return errors.New("failed to write file")
			},
			expectFatal:   true,
			fatalContains: "failed to write file",
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
			originalPodVerify := podVerificationFunc
			podVerificationFunc = func(msg string, cluster Cluster, namespace string) {
				podVerifyCalled = true
				assert.Equal(t, KUBESLICE_CONTROLLER_NAMESPACE, namespace)
			}
			defer func() { podVerificationFunc = originalPodVerify }()

			licenseVerifyCalled := false
			originalLicenseVerify := licenseVerificationFunc
			licenseVerificationFunc = func(msg string, cluster Cluster, namespace string) {
				licenseVerifyCalled = true
				assert.Equal(t, KUBESLICE_CONTROLLER_NAMESPACE, namespace)
			}
			defer func() { licenseVerificationFunc = originalLicenseVerify }()

			var capturedFilename string
			var capturedContent string
			originalGenerateValues := generateValuesFileFunc
			generateValuesFileFunc = func(filename string, chart *HelmChart, content string) error {
				capturedFilename = filename
				capturedContent = content
				return tt.mockGenerateValuesFile(filename, chart, content)
			}
			defer func() { generateValuesFileFunc = originalGenerateValues }()

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			InstallKubeSliceController(tt.config)

			if tt.expectFatal {
				assert.NotEmpty(t, fakeOutput.FatalCalls, "Expected fatal call")
				if tt.fatalContains != "" {
					assert.Contains(t, fakeOutput.FatalCalls[0], tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls, "Unexpected fatal calls")

				if tt.validateValuesContent {
					assert.Contains(t, capturedFilename, controllerValuesFileName)
					assert.Contains(t, capturedContent, "kubeslice:")
					assert.Contains(t, capturedContent, "controller:")
					assert.Contains(t, capturedContent, tt.config.Configuration.ClusterConfiguration.ControllerCluster.ControlPlaneAddress)
				}

				require.Len(t, fakeExec.Calls, 1, "Expected exactly one helm command")
				require.Equal(t, "helm", fakeExec.Calls[0].CLI)

				args := fakeExec.Calls[0].Args
				assert.Contains(t, args, "--kube-context")
				assert.Contains(t, args, "controller-context")
				assert.Contains(t, args, "--kubeconfig")
				assert.Contains(t, args, "/path/to/config")
				assert.Contains(t, args, "upgrade")
				assert.Contains(t, args, "-i")
				assert.Contains(t, args, KUBESLICE_CONTROLLER_NAMESPACE)
				assert.Contains(t, args, "kubeslice/kubeslice-controller")
				assert.Contains(t, args, "--namespace")
				assert.Contains(t, args, "--create-namespace")
				assert.Contains(t, args, "-f")

				if tt.config.Configuration.HelmChartConfiguration.ControllerChart.Version != "" {
					assert.True(t, containsSequence(args, []string{"--version", tt.config.Configuration.HelmChartConfiguration.ControllerChart.Version}))
				}

				assert.True(t, podVerifyCalled, "Expected pod verification to be called")
				assert.Equal(t, tt.expectLicenseVerification, licenseVerifyCalled,
					"License verification call mismatch")

				assert.Contains(t, fakeClock.SleepCalls, 200*time.Millisecond)
				assert.Contains(t, fakeClock.SleepCalls, 2*time.Second)
			}
		})
	}
}

func TestUninstallKubeSliceController(t *testing.T) {
	tests := []struct {
		name          string
		config        *ConfigurationSpecs
		mockExecutor  func(*util.FakeExecutor)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful uninstallation",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "controller",
							ContextName:    "controller-context",
							KubeConfigPath: "/path/to/config",
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
			name: "uninstallation fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "controller",
							ContextName:    "controller-context",
							KubeConfigPath: "/path/to/config",
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("uninstall failed")
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

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			UninstallKubeSliceController(tt.config)

			if tt.expectFatal {
				assert.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					assert.Contains(t, fakeOutput.FatalCalls[0], tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)

				require.Len(t, fakeExec.Calls, 1)
				require.Equal(t, "helm", fakeExec.Calls[0].CLI)

				args := fakeExec.Calls[0].Args
				assert.Contains(t, args, "--kube-context")
				assert.Contains(t, args, "controller-context")
				assert.Contains(t, args, "--kubeconfig")
				assert.Contains(t, args, "/path/to/config")
				assert.Contains(t, args, "uninstall")
				assert.Contains(t, args, KUBESLICE_CONTROLLER_NAMESPACE)
				assert.Contains(t, args, "--namespace")

				assert.Len(t, fakeClock.SleepCalls, 2)
			}
		})
	}
}

func TestControllerValuesContentGeneration(t *testing.T) {
	tests := []struct {
		name            string
		cluster         Cluster
		helmConfig      HelmChartConfiguration
		expectInContent []string
	}{
		{
			name: "basic values without image pull secrets",
			cluster: Cluster{
				ControlPlaneAddress: "https://controller.example.com",
			},
			helmConfig: HelmChartConfiguration{
				ControllerChart: HelmChart{
					ChartName: "kubeslice-controller",
				},
				ImagePullSecret: ImagePullSecrets{},
			},
			expectInContent: []string{
				"kubeslice:",
				"controller:",
				"loglevel: info",
				"endpoint: https://controller.example.com",
			},
		},
		{
			name: "values with image pull secrets",
			cluster: Cluster{
				ControlPlaneAddress: "https://prod.controller.com",
			},
			helmConfig: HelmChartConfiguration{
				ControllerChart: HelmChart{
					ChartName: "kubeslice-controller",
				},
				ImagePullSecret: ImagePullSecrets{
					Registry: "docker.io",
					Username: "testuser",
					Password: "testpass",
					Email:    "test@example.com",
				},
			},
			expectInContent: []string{
				"endpoint: https://prod.controller.com",
				"imagePullSecrets:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			var capturedContent string
			originalGenerateValues := generateValuesFileFunc
			generateValuesFileFunc = func(filename string, chart *HelmChart, content string) error {
				capturedContent = content
				assert.Contains(t, filename, controllerValuesFileName)
				return nil
			}
			defer func() { generateValuesFileFunc = originalGenerateValues }()

			generateControllerValuesFile(tt.cluster, tt.helmConfig)

			require.NotEmpty(t, capturedContent, "Expected values content to be generated")

			for _, expected := range tt.expectInContent {
				assert.Contains(t, capturedContent, expected,
					"Expected '%s' in values content", expected)
			}
		})
	}
}
