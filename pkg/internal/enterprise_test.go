package internal

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallKubeSliceUI(t *testing.T) {
	tests := []struct {
		name                   string
		config                 *ConfigurationSpecs
		mockExecutor           func(*util.FakeExecutor)
		mockGenerateValuesFile func(string, *HelmChart, string) error
		expectFatal            bool
		fatalContains          string
		shouldSkip             bool
		validateValuesContent  bool
		expectedServiceType    string
	}{
		{
			name: "successful installation with kind cluster",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "controller",
							ContextName:    "controller-context",
							KubeConfigPath: "/path/to/config",
						},
						ClusterType: "kind",
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						UIChart: HelmChart{
							ChartName: "kubeslice-ui",
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
			validateValuesContent: true,
			expectedServiceType:   "NodePort",
		},
		{
			name: "successful installation with eks cluster",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "controller",
							ContextName:    "controller-context",
							KubeConfigPath: "/path/to/config",
						},
						ClusterType: "eks",
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						UIChart: HelmChart{
							ChartName: "kubeslice-ui",
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
			validateValuesContent: true,
			expectedServiceType:   "LoadBalancer",
		},
		{
			name: "successful installation with version",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "controller",
							ContextName:    "controller-context",
							KubeConfigPath: "/path/to/config",
						},
						ClusterType: "kind",
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						UIChart: HelmChart{
							ChartName: "kubeslice-ui",
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
			validateValuesContent: true,
			expectedServiceType:   "NodePort",
		},
		{
			name: "skip installation when UI chart not configured",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "controller",
							ContextName:    "controller-context",
							KubeConfigPath: "/path/to/config",
						},
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						UIChart: HelmChart{
							ChartName: "",
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					t.Error("Should not execute helm when UI chart not configured")
					return nil
				}
			},
			mockGenerateValuesFile: func(filename string, chart *HelmChart, content string) error {
				return nil
			},
			shouldSkip: true,
		},
		{
			name: "helm installation fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "controller",
							ContextName:    "controller-context",
							KubeConfigPath: "/path/to/config",
						},
						ClusterType: "kind",
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						UIChart: HelmChart{
							ChartName: "kubeslice-ui",
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
							Name:           "controller",
							ContextName:    "controller-context",
							KubeConfigPath: "/path/to/config",
						},
						ClusterType: "kind",
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						UIChart: HelmChart{
							ChartName: "kubeslice-ui",
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
				assert.Equal(t, "kubernetes-dashboard", namespace)
			}
			defer func() { podVerificationFunc = originalPodVerify }()

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

			InstallKubeSliceUI(tt.config)

			if tt.expectFatal {
				assert.NotEmpty(t, fakeOutput.FatalCalls, "Expected fatal call")
				if tt.fatalContains != "" {
					assert.Contains(t, fakeOutput.FatalCalls[0], tt.fatalContains)
				}
			} else if tt.shouldSkip {
				assert.Empty(t, fakeExec.Calls, "Should not execute helm when UI chart not configured")
				assert.False(t, podVerifyCalled, "Should not verify pods when skipping")
			} else {
				assert.Empty(t, fakeOutput.FatalCalls, "Unexpected fatal calls")

				if tt.validateValuesContent {
					assert.Contains(t, capturedFilename, uiValuesFileName)
					assert.Contains(t, capturedContent, "kubeslice:")
					assert.Contains(t, capturedContent, "uiproxy:")
					assert.Contains(t, capturedContent, fmt.Sprintf("type: %s", tt.expectedServiceType))
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
				assert.Contains(t, args, "kubeslice-ui")
				assert.Contains(t, args, "kubeslice/kubeslice-ui")
				assert.Contains(t, args, "--namespace")
				assert.Contains(t, args, KUBESLICE_CONTROLLER_NAMESPACE)
				assert.Contains(t, args, "-f")

				if tt.config.Configuration.HelmChartConfiguration.UIChart.Version != "" {
					assert.True(t, containsSequence(args, []string{"--version", tt.config.Configuration.HelmChartConfiguration.UIChart.Version}))
				}

				assert.True(t, podVerifyCalled, "Expected pod verification to be called")
				assert.Contains(t, fakeClock.SleepCalls, 200*time.Millisecond)
			}
		})
	}
}

func TestUninstallKubeSliceUI(t *testing.T) {
	tests := []struct {
		name          string
		config        *ConfigurationSpecs
		mockExecutor  func(*util.FakeExecutor)
		expectFatal   bool
		fatalContains string
		expectSuccess bool
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
				callCount := 0
				fe.ExecuteFunc = func(cli string, args ...string) error {
					callCount++
					return nil
				}
			},
			expectSuccess: true,
		},
		{
			name: "UI not installed - skip uninstallation",
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
					return errors.New("release not found")
				}
			},
			expectSuccess: false,
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
				callCount := 0
				fe.ExecuteFunc = func(cli string, args ...string) error {
					callCount++
					if callCount == 1 {
						return nil
					}
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

			UninstallKubeSliceUI(tt.config)

			if tt.expectFatal {
				assert.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					assert.Contains(t, fakeOutput.FatalCalls[0], tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)

				infoMessages := fmt.Sprint(fakeOutput.InfoCalls)
				if tt.expectSuccess {
					assert.Contains(t, infoMessages, "Successfully uninstalled KubeSlice Manager")
					assert.Len(t, fakeExec.Calls, 2)
				} else {
					assert.Contains(t, infoMessages, "not installed, skipping uninstall")
					assert.Len(t, fakeExec.Calls, 1)
				}

				assert.Contains(t, fakeClock.SleepCalls, 200*time.Millisecond)
			}
		})
	}
}

func TestGetUIEndpoint(t *testing.T) {
	tests := []struct {
		name             string
		profile          string
		mockExecutor     func(*util.FakeExecutor)
		mockGetNodeIP    func(*Cluster) (string, error)
		expectedEndpoint string
		expectError      bool
	}{
		{
			name:    "NodePort service with EntDemo profile",
			profile: ProfileEntDemo,
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					jsonOutput := `'{"type":"NodePort","ports":[{"name":"http","nodePort":30443,"port":443}]}'`
					stdout.Write([]byte(jsonOutput))
					return nil
				}
			},
			expectedEndpoint: "https://localhost:8443",
		},
		{
			name:    "NodePort service without EntDemo profile",
			profile: "",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					jsonOutput := `'{"type":"NodePort","ports":[{"name":"http","nodePort":30443,"port":443}]}'`
					stdout.Write([]byte(jsonOutput))
					return nil
				}
			},
			mockGetNodeIP: func(cc *Cluster) (string, error) {
				return "'192.168.1.100'", nil
			},
			expectedEndpoint: "https://192.168.1.100:30443",
		},
		{
			name:    "LoadBalancer service with externalIPs",
			profile: "",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					jsonOutput := `'{"type":"LoadBalancer","externalIPs":["10.0.0.1"],"ports":[{"name":"http","port":443}]}'`
					stdout.Write([]byte(jsonOutput))
					return nil
				}
			},
			expectedEndpoint: "https://10.0.0.1:443",
		},
		{
			name:    "kubectl command fails",
			profile: "",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					return errors.New("service not found")
				}
			},
			expectedEndpoint: "",
			expectError:      true,
		},
		{
			name:    "unsupported service type",
			profile: "",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					jsonOutput := `'{"type":"ClusterIP"}'`
					stdout.Write([]byte(jsonOutput))
					return nil
				}
			},
			expectedEndpoint: "",
			expectError:      true,
		},
		{
			name:    "node IP retrieval fails",
			profile: "",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					jsonOutput := `'{"type":"NodePort","ports":[{"name":"http","nodePort":30443}]}'`
					stdout.Write([]byte(jsonOutput))
					return nil
				}
			},
			mockGetNodeIP: func(cc *Cluster) (string, error) {
				return "", errors.New("no nodes found")
			},
			expectedEndpoint: "",
			expectError:      true,
		},
		{
			name:    "invalid JSON response from kubectl",
			profile: "",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					invalidJSON := `'{"type":"NodePort","ports":[invalid json here]}'`
					stdout.Write([]byte(invalidJSON))
					return nil
				}
			},
			expectedEndpoint: "",
			expectError:      true,
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

			if tt.mockGetNodeIP != nil {
				originalGetNodeIP := getNodeIPFunc
				getNodeIPFunc = tt.mockGetNodeIP
				defer func() { getNodeIPFunc = originalGetNodeIP }()
			}

			cluster := &Cluster{
				ContextName:    "test-context",
				KubeConfigPath: "/path/to/config",
			}

			endpoint := GetUIEndpoint(cluster, tt.profile)

			if tt.expectError {
				assert.Empty(t, endpoint, "Expected empty endpoint on error")
			} else {
				assert.Equal(t, tt.expectedEndpoint, endpoint)
			}
		})
	}
}

func TestGetUIAdminToken(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		projectName   string
		mockExecutor  func(*util.FakeExecutor)
		expectedToken string
		expectFatal   bool
		fatalContains string
	}{
		{
			name:        "successful token retrieval",
			username:    "admin",
			projectName: "test-project",
			mockExecutor: func(fe *util.FakeExecutor) {
				callCount := 0
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					callCount++
					if callCount == 1 {
						stdout.Write([]byte("serviceaccount/rbac-rw-admin\nserviceaccount/other-sa"))
						return nil
					}
					token := base64.StdEncoding.EncodeToString([]byte("test-token-12345"))
					stdout.Write([]byte(token))
					return nil
				}
			},
			expectedToken: "test-token-12345",
		},
		{
			name:        "service account not found",
			username:    "admin",
			projectName: "test-project",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("serviceaccount/other-sa"))
					return nil
				}
			},
			expectFatal:   true,
			fatalContains: "failed to find secret",
		},
		{
			name:        "kubectl get sa fails",
			username:    "admin",
			projectName: "test-project",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					return errors.New("namespace not found")
				}
			},
			expectFatal:   true,
			fatalContains: "Process failed",
		},
		{
			name:        "kubectl get secret fails",
			username:    "admin",
			projectName: "test-project",
			mockExecutor: func(fe *util.FakeExecutor) {
				callCount := 0
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					callCount++
					if callCount == 1 {
						stdout.Write([]byte("serviceaccount/rbac-rw-admin"))
						return nil
					}
					return errors.New("secret not found")
				}
			},
			expectFatal:   true,
			fatalContains: "Process failed",
		},
		{
			name:        "invalid base64 token",
			username:    "admin",
			projectName: "test-project",
			mockExecutor: func(fe *util.FakeExecutor) {
				callCount := 0
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					callCount++
					if callCount == 1 {
						stdout.Write([]byte("serviceaccount/rbac-rw-admin"))
						return nil
					}
					stdout.Write([]byte("not-valid-base64!!!"))
					return nil
				}
			},
			expectFatal:   true,
			fatalContains: "Unable to decode token",
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

			cluster := &Cluster{
				ContextName:    "test-context",
				KubeConfigPath: "/path/to/config",
			}

			token := GetUIAdminToken(cluster, tt.username, tt.projectName)

			if tt.expectFatal {
				assert.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					assert.Contains(t, fakeOutput.FatalCalls[0], tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestGetNodeIP(t *testing.T) {
	tests := []struct {
		name           string
		mockExecutor   func(*util.FakeExecutor)
		expectedNodeIP string
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful node IP retrieval",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("'192.168.1.100' '192.168.1.101'"))
					return nil
				}
			},
			expectedNodeIP: "'192.168.1.100'",
		},
		{
			name: "single node IP",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("'10.0.0.1'"))
					return nil
				}
			},
			expectedNodeIP: "'10.0.0.1'",
		},
		{
			name: "no nodes found",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte(""))
					return nil
				}
			},
			expectError:   true,
			errorContains: "No nodes found",
		},
		{
			name: "kubectl command fails",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					return errors.New("connection refused")
				}
			},
			expectError:   true,
			errorContains: "connection refused",
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

			cluster := &Cluster{
				ContextName:    "test-context",
				KubeConfigPath: "/path/to/config",
			}

			nodeIP, err := getNodeIP(cluster)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedNodeIP, nodeIP)
			}
		})
	}
}
