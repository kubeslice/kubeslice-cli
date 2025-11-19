package internal

import (
	"encoding/json"
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

func TestInstallKubeSliceWorker(t *testing.T) {
	baseConfig := &ConfigurationSpecs{
		Configuration: Configuration{
			ClusterConfiguration: ClusterConfiguration{
				ClusterType:       "eks",
				ControllerCluster: Cluster{Name: "controller", ContextName: "controller-ctx"},
				WorkerClusters: []Cluster{
					{Name: "worker-1", ContextName: "worker-1-ctx"},
					{Name: "worker-2", ContextName: "worker-2-ctx"},
				},
			},
			KubeSliceConfiguration: KubeSliceConfiguration{ProjectName: "test-project"},
			HelmChartConfiguration: HelmChartConfiguration{
				RepoAlias:   "kubeslice",
				WorkerChart: HelmChart{ChartName: "kubeslice-worker"},
			},
		},
	}

	mockSecretData := map[string]string{
		"namespace":          "ns",
		"controllerEndpoint": "https://controller",
		"ca.crt":             "cert",
		"token":              "token",
	}

	tests := []struct {
		name                string
		config              *ConfigurationSpecs
		mockExecutor        func(*util.FakeExecutor)
		mockPodVerification func(string, Cluster, string)
		mockRetry           func(int, time.Duration, func() error) error
		mockGenerateValues  func(string, *HelmChart, string) error
		mockFetchSecret     func(string, Cluster, string) map[string]string
		expectFatal         bool
		fatalContains       string
		expectedWorkerCount int
	}{
		{
			name:   "successful installation",
			config: baseConfig,
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error { return nil }
			},
			mockPodVerification: func(msg string, cluster Cluster, namespace string) {},
			mockRetry:           func(backoff int, sleep time.Duration, f func() error) error { return f() },
			mockFetchSecret:     func(clusterName string, cc Cluster, projectName string) map[string]string { return mockSecretData },
			mockGenerateValues:  func(fileName string, chart *HelmChart, values string) error { return nil },
			expectedWorkerCount: 2,
		},
		{
			name:          "secret fetch fails",
			config:        baseConfig,
			mockRetry:     func(backoff int, sleep time.Duration, f func() error) error { return errors.New("retry failed") },
			expectFatal:   true,
			fatalContains: "Unable to fetch secrets",
		},
		{
			name:   "helm install fails",
			config: baseConfig,
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error { return errors.New("helm failed") }
			},
			mockRetry:          func(backoff int, sleep time.Duration, f func() error) error { return f() },
			mockFetchSecret:    func(clusterName string, cc Cluster, projectName string) map[string]string { return mockSecretData },
			mockGenerateValues: func(fileName string, chart *HelmChart, values string) error { return nil },
			expectFatal:        true,
			fatalContains:      "helm failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeOutput := util.Output.(*util.FakeOutput)
			fakeClock := util.SystemClock.(*util.FakeClock)

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			podVerifyCallCount := 0
			originalPodVerify := podVerificationFuncWorker
			podVerificationFuncWorker = func(msg string, cluster Cluster, namespace string) {
				podVerifyCallCount++
				if tt.mockPodVerification != nil {
					tt.mockPodVerification(msg, cluster, namespace)
				}
			}
			defer func() { podVerificationFuncWorker = originalPodVerify }()

			originalRetry := retryFunc
			if tt.mockRetry != nil {
				retryFunc = tt.mockRetry
			}
			defer func() { retryFunc = originalRetry }()

			originalGenerateValues := generateValuesFileFunc
			if tt.mockGenerateValues != nil {
				generateValuesFileFunc = tt.mockGenerateValues
			}
			defer func() { generateValuesFileFunc = originalGenerateValues }()

			originalFetchSecret := fetchSecretFunc
			if tt.mockFetchSecret != nil {
				fetchSecretFunc = tt.mockFetchSecret
			}
			defer func() { fetchSecretFunc = originalFetchSecret }()

			InstallKubeSliceWorker(tt.config)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Equal(t, tt.expectedWorkerCount, podVerifyCallCount)
				expectedSleeps := tt.expectedWorkerCount*2 + 1
				require.Len(t, fakeClock.SleepCalls, expectedSleeps)
			}

			cleanup()
		})
	}
}

func TestUninstallKubeSliceWorker(t *testing.T) {
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

	tests := []struct {
		name               string
		workersToUninstall map[string]string
		expectedUninstalls int
	}{
		{
			name:               "uninstall all with wildcard",
			workersToUninstall: map[string]string{"*": ""},
			expectedUninstalls: 2,
		},
		{
			name:               "uninstall specific worker",
			workersToUninstall: map[string]string{"worker-1": ""},
			expectedUninstalls: 1,
		},
		{
			name:               "no workers to uninstall",
			workersToUninstall: map[string]string{},
			expectedUninstalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeClock := util.SystemClock.(*util.FakeClock)

			fakeExec.ExecuteFunc = func(cli string, args ...string) error {
				assert.Equal(t, "helm", cli)
				return nil
			}

			UninstallKubeSliceWorker(config, tt.workersToUninstall)

			assert.Len(t, fakeExec.Calls, tt.expectedUninstalls)
			expectedSleeps := tt.expectedUninstalls + 1
			require.Len(t, fakeClock.SleepCalls, expectedSleeps)

			cleanup()
		})
	}
}

func TestRetry(t *testing.T) {
	tests := []struct {
		name           string
		backoffLimit   int
		sleep          time.Duration
		function       func() error
		expectError    bool
		errorContains  string
		expectedCalls  int
		expectedSleeps int
	}{
		{
			name:           "succeeds on first attempt",
			backoffLimit:   3,
			sleep:          1 * time.Second,
			function:       func() error { return nil },
			expectedCalls:  1,
			expectedSleeps: 0,
		},
		{
			name:         "succeeds on third attempt",
			backoffLimit: 5,
			sleep:        1 * time.Second,
			function: func() func() error {
				callCount := 0
				return func() error {
					callCount++
					if callCount < 3 {
						return errors.New("temp error")
					}
					return nil
				}
			}(),
			expectedCalls:  3,
			expectedSleeps: 2,
		},
		{
			name:           "fails after all retries",
			backoffLimit:   3,
			sleep:          1 * time.Second,
			function:       func() error { return errors.New("persistent") },
			expectError:    true,
			errorContains:  "retry failed after 3 attempts",
			expectedCalls:  3,
			expectedSleeps: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			fakeClock := util.SystemClock.(*util.FakeClock)

			callCount := 0
			testFunc := func() error {
				callCount++
				return tt.function()
			}

			err := Retry(tt.backoffLimit, tt.sleep, testFunc)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedCalls, callCount)
			assert.Len(t, fakeClock.SleepCalls, tt.expectedSleeps)

			cleanup()
		})
	}
}

func TestFetchSecret(t *testing.T) {
	cluster := Cluster{ContextName: "test-ctx"}

	tests := []struct {
		name          string
		clusterName   string
		projectName   string
		mockExecutor  func(*util.FakeExecutor)
		expectedData  map[string]string
		expectFatal   bool
		fatalContains string
	}{
		{
			name:        "successful secret fetch",
			clusterName: "worker-1",
			projectName: "test-project",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					cmdStr := strings.Join(args, " ")
					if strings.Contains(cmdStr, "get sa") {
						stdout.Write([]byte("serviceaccount/rbac-worker-worker-1"))
					} else if strings.Contains(cmdStr, "get secrets") {
						secretData := map[string]string{"token": "token-data"}
						data, _ := json.Marshal(secretData)
						stdout.Write(data)
					}
					return nil
				}
			},
			expectedData: map[string]string{"token": "token-data"},
		},
		{
			name:        "kubectl get sa fails",
			clusterName: "worker-1",
			projectName: "test-project",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					return errors.New("sa not found")
				}
			},
			expectFatal:   true,
			fatalContains: "sa not found",
		},
		{
			name:        "secret not found for worker",
			clusterName: "worker-1",
			projectName: "test-project",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("serviceaccount/default"))
					return nil
				}
			},
			expectFatal:   true,
			fatalContains: "failed to find secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeOutput := util.Output.(*util.FakeOutput)

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			secrets := fetchSecret(tt.clusterName, cluster, tt.projectName)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				if tt.fatalContains != "" {
					assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Equal(t, tt.expectedData, secrets)
			}

			cleanup()
		})
	}
}
