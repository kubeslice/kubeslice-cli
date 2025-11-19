package internal

import (
	"errors"
	"io"
	"testing"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyKubectlManifest(t *testing.T) {
	cluster := &Cluster{
		Name:           "test-cluster",
		ContextName:    "test-ctx",
		KubeConfigPath: "/path/to/config",
	}

	tests := []struct {
		name          string
		fileName      string
		namespace     string
		cluster       *Cluster
		mockExecutor  func(*util.FakeExecutor)
		expectFatal   bool
		fatalContains string
	}{
		{
			name:      "successful apply with cluster",
			fileName:  "manifest.yaml",
			namespace: "default",
			cluster:   cluster,
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					assert.Equal(t, "kubectl", cli)
					assert.Contains(t, args, "--context=test-ctx")
					assert.Contains(t, args, "--kubeconfig=/path/to/config")
					assert.Contains(t, args, "apply")
					assert.Contains(t, args, "-f")
					assert.Contains(t, args, "manifest.yaml")
					assert.Contains(t, args, "-n")
					assert.Contains(t, args, "default")
					return nil
				}
			},
		},
		{
			name:      "successful apply without cluster",
			fileName:  "manifest.yaml",
			namespace: "default",
			cluster:   nil,
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					assert.Equal(t, "kubectl", cli)
					assert.NotContains(t, args, "--context=test-ctx")
					return nil
				}
			},
		},
		{
			name:      "command failure triggers fatal",
			fileName:  "manifest.yaml",
			namespace: "default",
			cluster:   cluster,
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("connection refused")
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

			ApplyKubectlManifest(tt.fileName, tt.namespace, tt.cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fakeOutput.FatalCalls[0], tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
			}
		})
	}
}

func TestGetKubectlResources(t *testing.T) {
	cluster := &Cluster{ContextName: "ctx", KubeConfigPath: "conf"}

	cleanup := util.NewTestEnvironment()
	defer cleanup()

	fakeExec := util.CommandExecutor.(*util.FakeExecutor)

	// Test with resource name
	fakeExec.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
		assert.Contains(t, args, "get")
		assert.Contains(t, args, "pods")
		assert.Contains(t, args, "my-pod")
		assert.Contains(t, args, "-o")
		assert.Contains(t, args, "yaml")
		return nil
	}
	GetKubectlResources("pods", "my-pod", "default", cluster, "yaml")

	fakeExec.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
		assert.Contains(t, args, "get")
		assert.Contains(t, args, "pods")
		assert.NotContains(t, args, "my-pod")
		return nil
	}
	GetKubectlResources("pods", "", "default", cluster, "")
}

func TestDeleteKubectlResources(t *testing.T) {
	cluster := &Cluster{ContextName: "ctx", KubeConfigPath: "conf"}
	cleanup := util.NewTestEnvironment()
	defer cleanup()

	fakeExec := util.CommandExecutor.(*util.FakeExecutor)
	fakeExec.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
		assert.Equal(t, "kubectl", cli)
		assert.Contains(t, args, "delete")
		assert.Contains(t, args, "pods")
		return nil
	}

	DeleteKubectlResources("pods", "my-pod", "default", cluster)
}

func TestDescribeKubectlResources(t *testing.T) {
	cluster := &Cluster{ContextName: "ctx", KubeConfigPath: "conf"}
	cleanup := util.NewTestEnvironment()
	defer cleanup()

	fakeExec := util.CommandExecutor.(*util.FakeExecutor)
	fakeExec.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
		assert.Equal(t, "kubectl", cli)
		assert.Contains(t, args, "describe")
		return nil
	}

	DescribeKubectlResources("pods", "my-pod", "default", cluster)
}

func TestVerifyPods(t *testing.T) {
	cluster := Cluster{
		ContextName:    "test-ctx",
		KubeConfigPath: "/path/to/config",
	}

	tests := []struct {
		name           string
		mockOutput     string
		mockError      error
		expectedStatus PodVerificationStatus
	}{
		{
			name: "all pods ready",
			mockOutput: `NAME                      READY   STATUS    RESTARTS   AGE
pod-1                     2/2     Running   0          1m
pod-2                     1/1     Running   0          1m`,
			expectedStatus: PodVerificationStatusSuccess,
		},
		{
			name: "pods still starting",
			mockOutput: `NAME                      READY   STATUS    RESTARTS   AGE
pod-1                     1/2     Running   0          1m
pod-2                     0/1     Pending   0          1m`,
			expectedStatus: PodVerificationStatusInProgress,
		},
		{
			name: "pod in error state",
			mockOutput: `NAME                      READY   STATUS             RESTARTS   AGE
pod-1                     0/1     ImagePullBackOff   0          1m`,
			expectedStatus: PodVerificationStatusFailed,
		},
		{
			name: "pod crashed",
			mockOutput: `NAME                      READY   STATUS             RESTARTS   AGE
pod-1                     0/1     CrashLoopBackOff     3          1m`,
			expectedStatus: PodVerificationStatusFailed,
		},
		{
			name: "completed pods ignored",
			mockOutput: `NAME                      READY   STATUS      RESTARTS   AGE
pod-1                     0/1     Completed   0          1m
pod-2                     1/1     Running     0          1m`,
			expectedStatus: PodVerificationStatusSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeExec.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
				stdout.Write([]byte(tt.mockOutput))
				return tt.mockError
			}

			status, _ := verifyPods(cluster, "default")
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestFetchLicenseSecret(t *testing.T) {
	cluster := Cluster{
		ContextName:    "test-ctx",
		KubeConfigPath: "/path/to/config",
	}

	tests := []struct {
		name        string
		secretName  string
		mockOutput  string
		mockError   error
		expectError bool
	}{
		{
			name:        "secret found",
			secretName:  "license",
			mockOutput:  "NAME      TYPE     DATA\nlicense   Opaque   1",
			expectError: false,
		},
		{
			name:        "secret not found",
			secretName:  "license",
			mockOutput:  "No resources found",
			expectError: true,
		},
		{
			name:        "kubectl command fails",
			secretName:  "license",
			mockError:   errors.New("connection refused"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeExec.ExecuteWithOutputFunc = func(cli string, stdout, stderr io.Writer, args ...string) error {
				if tt.mockError != nil {
					return tt.mockError
				}
				stdout.Write([]byte(tt.mockOutput))
				return nil
			}

			err := fetchLicenseSecret(tt.secretName, cluster, "default")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
