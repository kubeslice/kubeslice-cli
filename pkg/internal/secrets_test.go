package internal

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSecrets(t *testing.T) {
	cluster := &Cluster{Name: "controller"}

	tests := []struct {
		name                    string
		workerName              string
		namespace               string
		outputFormat            string
		mockGetSecretName       string
		mockGetKubectlResources func(string, string, string, *Cluster, string)
		expectGetResourcesCall  bool
		expectFatal             bool
	}{
		{
			name:              "successful secret retrieval",
			workerName:        "worker-1",
			namespace:         "kubeslice-test",
			outputFormat:      "yaml",
			mockGetSecretName: "kubeslice-rbac-worker-worker-1-token-abc123",
			mockGetKubectlResources: func(resourceType, resourceName, namespace string, cluster *Cluster, outputFormat string) {
				assert.Equal(t, SecretObject, resourceType)
				assert.Equal(t, "kubeslice-rbac-worker-worker-1-token-abc123", resourceName)
				assert.Equal(t, "yaml", outputFormat)
			},
			expectGetResourcesCall: true,
		},
		{
			name:                   "no secret found for worker",
			workerName:             "worker-2",
			namespace:              "kubeslice-test",
			outputFormat:           "",
			mockGetSecretName:      "",
			expectGetResourcesCall: false,
		},
		{
			name:              "get kubectl resources fails",
			workerName:        "worker-1",
			namespace:         "kubeslice-test",
			outputFormat:      "yaml",
			mockGetSecretName: "kubeslice-rbac-worker-worker-1-token-abc123",
			mockGetKubectlResources: func(resourceType, resourceName, namespace string, cluster *Cluster, outputFormat string) {
				util.Fatalf("Process failed %v", errors.New("get failed"))
			},
			expectGetResourcesCall: true,
			expectFatal:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeClock := util.SystemClock.(*util.FakeClock)
			fakeOutput := util.Output.(*util.FakeOutput)

			originalGetSecretName := getSecretNameFunc
			getSecretNameFunc = func(workerName, namespace string, c *Cluster) string {
				return tt.mockGetSecretName
			}
			defer func() { getSecretNameFunc = originalGetSecretName }()

			getCalled := false
			originalGetKubectlResources := getKubectlResourcesFuncSecrets
			getKubectlResourcesFuncSecrets = func(resourceType, resourceName, namespace string, cluster *Cluster, outputFormat string) {
				getCalled = true
				if tt.mockGetKubectlResources != nil {
					tt.mockGetKubectlResources(resourceType, resourceName, namespace, cluster, outputFormat)
				}
			}
			defer func() { getKubectlResourcesFuncSecrets = originalGetKubectlResources }()

			GetSecrets(tt.workerName, tt.namespace, cluster, tt.outputFormat)

			assert.Equal(t, tt.expectGetResourcesCall, getCalled)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
			}

			if !tt.expectGetResourcesCall {
				allOutput := strings.Join(fakeOutput.InfoCalls, " ")
				assert.Contains(t, allOutput, "No secret found")
			} else {
				require.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			}
		})
	}
}

func TestGetSecrets_NilCluster(t *testing.T) {
	cleanup := util.NewTestEnvironment()
	defer cleanup()

	fakeOutput := util.Output.(*util.FakeOutput)

	originalGetSecretName := getSecretNameFunc
	defer func() { getSecretNameFunc = originalGetSecretName }()

	getSecretNameFunc = func(workerName, namespace string, c *Cluster) string {
		t.Fatal("Should not call GetSecretName with nil cluster")
		return ""
	}

	GetSecrets("worker-1", "kubeslice-test", nil, "yaml")

	allOutput := strings.Join(fakeOutput.InfoCalls, " ")
	assert.Contains(t, allOutput, "Controller cluster cannot be nil")
}

func TestGetSecretName(t *testing.T) {
	cluster := &Cluster{ContextName: "controller-ctx", KubeConfigPath: "/path/config"}

	tests := []struct {
		name           string
		workerName     string
		namespace      string
		mockExecutor   func(*util.FakeExecutor)
		expectedSecret string
		expectMessage  string
	}{
		{
			name:       "secret found for worker",
			workerName: "worker-1",
			namespace:  "kubeslice-test",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("kubeslice-rbac-worker-worker-1-token-abc123"))
					return nil
				}
			},
			expectedSecret: "kubeslice-rbac-worker-worker-1-token-abc123",
		},
		{
			name:       "no secret found for worker",
			workerName: "worker-2",
			namespace:  "kubeslice-test",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte(""))
					return nil
				}
			},
			expectedSecret: "",
			expectMessage:  "No matching secret found",
		},
		{
			name:       "kubectl command fails",
			workerName: "worker-3",
			namespace:  "kubeslice-test",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					return errors.New("namespace not found")
				}
			},
			expectedSecret: "",
			expectMessage:  "Failed to get secret",
		},
		{
			name:       "secret with whitespace trimming",
			workerName: "worker-4",
			namespace:  "kubeslice-test",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
					stdout.Write([]byte("  kubeslice-rbac-worker-worker-4-token-xyz  \n"))
					return nil
				}
			},
			expectedSecret: "kubeslice-rbac-worker-worker-4-token-xyz",
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

			secretName := GetSecretName(tt.workerName, tt.namespace, cluster)

			assert.Equal(t, tt.expectedSecret, secretName)
			assert.Empty(t, fakeOutput.FatalCalls)

			if tt.expectMessage != "" {
				allOutput := strings.Join(fakeOutput.InfoCalls, " ")
				assert.Contains(t, allOutput, tt.expectMessage)
			}
		})
	}
}

func TestGetSecretName_NilCluster(t *testing.T) {
	cleanup := util.NewTestEnvironment()
	defer cleanup()

	fakeOutput := util.Output.(*util.FakeOutput)
	fakeExec := util.CommandExecutor.(*util.FakeExecutor)

	fakeExec.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
		t.Fatal("Should not execute kubectl with nil cluster")
		return nil
	}

	secretName := GetSecretName("worker-1", "kubeslice-test", nil)

	assert.Equal(t, "", secretName)
	allOutput := strings.Join(fakeOutput.InfoCalls, " ")
	assert.Contains(t, allOutput, "Controller cluster cannot be nil")
}
