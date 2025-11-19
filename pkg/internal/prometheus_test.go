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

func TestInstallPrometheus(t *testing.T) {
	tests := []struct {
		name                string
		config              *ConfigurationSpecs
		mockExecutor        func(*util.FakeExecutor, *int, *int)
		mockPodVerification func(string, Cluster, string)
		mockGenerateValues  func(string, *HelmChart, string) error
		expectFatal         bool
		fatalContains       string
		expectedHelmCalls   int
		expectedPatchCalls  int
	}{
		{
			name: "successful installation on multiple workers",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "controller",
							ContextName:    "controller-ctx",
							KubeConfigPath: "/path/config",
						},
						WorkerClusters: []Cluster{
							{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path", NodeIP: "10.0.0.1"},
							{Name: "worker-2", ContextName: "w2-ctx", KubeConfigPath: "/path", NodeIP: "10.0.0.2"},
						},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName: "test-project",
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "prometheus-community",
						PrometheusChart: HelmChart{
							ChartName: "prometheus",
							Version:   "v1.0.0",
						},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor, helmCalls, patchCalls *int) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					if cli == "helm" {
						*helmCalls++
						assert.Contains(t, args, "upgrade")
						assert.Contains(t, args, "-i")
						assert.Contains(t, args, "prometheus")
						assert.Contains(t, args, "--version")
						assert.Contains(t, args, "v1.0.0")
					} else if cli == "kubectl" {
						*patchCalls++
						assert.Contains(t, args, "patch")
						assert.Contains(t, args, ClusterObject)
						argsStr := strings.Join(args, " ")
						assert.Contains(t, argsStr, "prometheus")
						assert.Contains(t, argsStr, "32700")
					}
					return nil
				}
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {
				assert.Equal(t, PrometheusNamespace, namespace)
			},
			mockGenerateValues: func(fileName string, chart *HelmChart, values string) error {
				assert.Contains(t, fileName, PrometheusValuesFileName)
				return nil
			},
			expectedHelmCalls:  2,
			expectedPatchCalls: 2,
		},
		{
			name: "installation with single worker",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller", ContextName: "ctrl-ctx", KubeConfigPath: "/path"},
						WorkerClusters: []Cluster{
							{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path", NodeIP: "10.0.0.1"},
						},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName: "single-project",
					},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias:       "prometheus-community",
						PrometheusChart: HelmChart{ChartName: "prometheus"},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor, helmCalls, patchCalls *int) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					if cli == "helm" {
						*helmCalls++
					} else if cli == "kubectl" {
						*patchCalls++
					}
					return nil
				}
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {},
			mockGenerateValues: func(fileName string, chart *HelmChart, values string) error {
				return nil
			},
			expectedHelmCalls:  1,
			expectedPatchCalls: 1,
		},
		{
			name: "values file generation fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
						WorkerClusters:    []Cluster{{Name: "worker-1", NodeIP: "10.0.0.1"}},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{ProjectName: "test"},
					HelmChartConfiguration: HelmChartConfiguration{
						PrometheusChart: HelmChart{ChartName: "prometheus"},
					},
				},
			},
			mockGenerateValues: func(fileName string, chart *HelmChart, values string) error {
				return errors.New("file write failed")
			},
			expectFatal:   true,
			fatalContains: "file write failed",
		},
		{
			name: "helm installation fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller", ContextName: "ctrl-ctx", KubeConfigPath: "/path"},
						WorkerClusters:    []Cluster{{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path", NodeIP: "10.0.0.1"}},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{ProjectName: "test"},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias:       "prometheus-community",
						PrometheusChart: HelmChart{ChartName: "prometheus"},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor, helmCalls, patchCalls *int) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					if cli == "helm" {
						return errors.New("helm installation failed")
					}
					return nil
				}
			},
			mockGenerateValues: func(fileName string, chart *HelmChart, values string) error {
				return nil
			},
			expectFatal:   true,
			fatalContains: "helm installation failed",
		},
		{
			name: "cluster patch fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller", ContextName: "ctrl-ctx", KubeConfigPath: "/path"},
						WorkerClusters:    []Cluster{{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path", NodeIP: "10.0.0.1"}},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{ProjectName: "test"},
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias:       "prometheus-community",
						PrometheusChart: HelmChart{ChartName: "prometheus"},
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor, helmCalls, patchCalls *int) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					if cli == "kubectl" && strings.Contains(strings.Join(args, " "), "patch") {
						return errors.New("patch failed")
					}
					return nil
				}
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {},
			mockGenerateValues: func(fileName string, chart *HelmChart, values string) error {
				return nil
			},
			expectFatal:   true,
			fatalContains: "patch failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeFS := util.FileSystem.(*util.FakeFileSystem)
			fakeOutput := util.Output.(*util.FakeOutput)
			fakeClock := util.SystemClock.(*util.FakeClock)

			helmCalls := 0
			patchCalls := 0

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec, &helmCalls, &patchCalls)
			}

			podVerifyCallCount := 0
			originalPodVerify := podVerificationFuncPrometheus
			podVerificationFuncPrometheus = func(msg string, cluster Cluster, namespace string) {
				podVerifyCallCount++
				if tt.mockPodVerification != nil {
					tt.mockPodVerification(msg, cluster, namespace)
				}
			}
			defer func() { podVerificationFuncPrometheus = originalPodVerify }()

			originalGenerateValues := generateValuesFileFunc
			generateValuesFileFunc = func(fileName string, chart *HelmChart, values string) error {
				if tt.mockGenerateValues != nil {
					err := tt.mockGenerateValues(fileName, chart, values)
					if err == nil {
						fakeFS.WrittenFiles[fileName] = []byte("mock values")
					}
					return err
				}
				fakeFS.WrittenFiles[fileName] = []byte("mock values")
				return nil
			}
			defer func() { generateValuesFileFunc = originalGenerateValues }()

			InstallPrometheus(tt.config)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Equal(t, tt.expectedHelmCalls, helmCalls)
				assert.Equal(t, tt.expectedPatchCalls, patchCalls)
				assert.Equal(t, tt.expectedHelmCalls, podVerifyCallCount)

				valuesPath := kubesliceDirectory + "/" + PrometheusValuesFileName
				_, exists := fakeFS.WrittenFiles[valuesPath]
				assert.True(t, exists, "Prometheus values file should be created")

				// Verify exact sleeps: 1 (after generate) + N (in install loop) + 1 (after install)
				expectedSleeps := 2 + tt.expectedHelmCalls
				require.Len(t, fakeClock.SleepCalls, expectedSleeps)
				for _, d := range fakeClock.SleepCalls {
					assert.Equal(t, 200*time.Millisecond, d)
				}
			}
		})
	}
}

func TestPatchClusterObjectInControllerCluster(t *testing.T) {
	controller := &Cluster{
		Name:           "controller",
		ContextName:    "controller-ctx",
		KubeConfigPath: "/path/config",
	}

	tests := []struct {
		name          string
		workers       []Cluster
		projectNS     string
		mockExecutor  func(*util.FakeExecutor)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful patch for multiple workers",
			workers: []Cluster{
				{Name: "worker-1", NodeIP: "10.0.0.1"},
				{Name: "worker-2", NodeIP: "10.0.0.2"},
			},
			projectNS: "kubeslice-test-project",
			mockExecutor: func(fe *util.FakeExecutor) {
				callCount := 0
				fe.ExecuteFunc = func(cli string, args ...string) error {
					callCount++
					assert.Equal(t, "kubectl", cli)
					assert.Contains(t, args, "patch")
					assert.Contains(t, args, ClusterObject)

					argsStr := strings.Join(args, " ")
					assert.Contains(t, argsStr, "kubeslice-test-project")
					assert.Contains(t, argsStr, "prometheus")
					assert.Contains(t, argsStr, "32700")

					if callCount == 1 {
						assert.Contains(t, argsStr, "worker-1")
						assert.Contains(t, argsStr, "10.0.0.1")
					} else if callCount == 2 {
						assert.Contains(t, argsStr, "worker-2")
						assert.Contains(t, argsStr, "10.0.0.2")
					}
					return nil
				}
			},
		},
		{
			name: "patch fails for worker",
			workers: []Cluster{
				{Name: "worker-1", NodeIP: "10.0.0.1"},
			},
			projectNS: "kubeslice-test-project",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("cluster not found")
				}
			},
			expectFatal:   true,
			fatalContains: "cluster not found",
		},
		{
			name: "verify endpoint format",
			workers: []Cluster{
				{Name: "worker-1", NodeIP: "192.168.1.10"},
			},
			projectNS: "kubeslice-test",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					argsStr := strings.Join(args, " ")
					assert.Contains(t, argsStr, "http://192.168.1.10:32700")
					assert.Contains(t, argsStr, "telemetryProvider\":\"prometheus\"")
					assert.Contains(t, argsStr, "enabled\":true")
					return nil
				}
			},
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

			patchClusterObjectInControllerCluster(tt.workers, controller, tt.projectNS)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Len(t, fakeExec.Calls, len(tt.workers))
			}
		})
	}
}

func TestGeneratePrometheusValuesFile(t *testing.T) {
	tests := []struct {
		name               string
		helmConfig         HelmChartConfiguration
		mockGenerateValues func(string, *HelmChart, string) error
		expectFatal        bool
		fatalContains      string
	}{
		{
			name: "successful values file generation",
			helmConfig: HelmChartConfiguration{
				PrometheusChart: HelmChart{
					ChartName: "prometheus",
					Version:   "v1.0.0",
				},
			},
			mockGenerateValues: func(fileName string, chart *HelmChart, values string) error {
				assert.Contains(t, fileName, PrometheusValuesFileName)
				assert.Equal(t, "prometheus", chart.ChartName)
				return nil
			},
		},
		{
			name: "values file generation fails",
			helmConfig: HelmChartConfiguration{
				PrometheusChart: HelmChart{ChartName: "prometheus"},
			},
			mockGenerateValues: func(fileName string, chart *HelmChart, values string) error {
				return errors.New("file write failed")
			},
			expectFatal:   true,
			fatalContains: "file write failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeFS := util.FileSystem.(*util.FakeFileSystem)
			fakeOutput := util.Output.(*util.FakeOutput)

			originalGenerateValues := generateValuesFileFunc
			generateValuesFileFunc = func(fileName string, chart *HelmChart, values string) error {
				err := tt.mockGenerateValues(fileName, chart, values)
				if err == nil {
					fakeFS.WrittenFiles[fileName] = []byte("mock values")
				}
				return err
			}
			defer func() { generateValuesFileFunc = originalGenerateValues }()

			generatePrometheusValuesFile(tt.helmConfig)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				valuesPath := kubesliceDirectory + "/" + PrometheusValuesFileName
				_, exists := fakeFS.WrittenFiles[valuesPath]
				assert.True(t, exists)
			}
		})
	}
}

func TestInstallPrometheusFunction(t *testing.T) {
	controller := &Cluster{Name: "controller"}

	tests := []struct {
		name                string
		clusters            []Cluster
		helmConfig          HelmChartConfiguration
		mockExecutor        func(*util.FakeExecutor, *int)
		mockPodVerification func(string, Cluster, string)
		expectFatal         bool
		fatalContains       string
		expectedCalls       int
	}{
		{
			name: "install on multiple clusters with version",
			clusters: []Cluster{
				{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path"},
				{Name: "worker-2", ContextName: "w2-ctx", KubeConfigPath: "/path"},
			},
			helmConfig: HelmChartConfiguration{
				RepoAlias: "prometheus-community",
				PrometheusChart: HelmChart{
					ChartName: "prometheus",
					Version:   "v1.0.0",
				},
			},
			mockExecutor: func(fe *util.FakeExecutor, callCount *int) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					*callCount++
					assert.Equal(t, "helm", cli)
					assert.Contains(t, args, "upgrade")
					assert.Contains(t, args, "--version")
					assert.Contains(t, args, "v1.0.0")
					return nil
				}
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {
				assert.Equal(t, PrometheusNamespace, namespace)
			},
			expectedCalls: 2,
		},
		{
			name: "install without version",
			clusters: []Cluster{
				{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path"},
			},
			helmConfig: HelmChartConfiguration{
				RepoAlias:       "prometheus-community",
				PrometheusChart: HelmChart{ChartName: "prometheus", Version: ""},
			},
			mockExecutor: func(fe *util.FakeExecutor, callCount *int) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					*callCount++
					assert.NotContains(t, args, "--version")
					return nil
				}
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {},
			expectedCalls:       1,
		},
		{
			name: "helm install fails",
			clusters: []Cluster{
				{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path"},
			},
			helmConfig: HelmChartConfiguration{
				RepoAlias:       "prometheus-community",
				PrometheusChart: HelmChart{ChartName: "prometheus"},
			},
			mockExecutor: func(fe *util.FakeExecutor, callCount *int) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("helm repo not found")
				}
			},
			expectFatal:   true,
			fatalContains: "helm repo not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeOutput := util.Output.(*util.FakeOutput)
			fakeClock := util.SystemClock.(*util.FakeClock)

			callCount := 0
			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec, &callCount)
			}

			podVerifyCount := 0
			originalPodVerify := podVerificationFuncPrometheus
			podVerificationFuncPrometheus = func(msg string, cluster Cluster, namespace string) {
				podVerifyCount++
				if tt.mockPodVerification != nil {
					tt.mockPodVerification(msg, cluster, namespace)
				}
			}
			defer func() { podVerificationFuncPrometheus = originalPodVerify }()

			installPrometheus(tt.clusters, controller, tt.helmConfig, PrometheusValuesFileName)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Equal(t, tt.expectedCalls, callCount)
				assert.Equal(t, tt.expectedCalls, podVerifyCount)
				require.Len(t, fakeClock.SleepCalls, tt.expectedCalls)
				for _, d := range fakeClock.SleepCalls {
					assert.Equal(t, 200*time.Millisecond, d)
				}
			}
		})
	}
}
