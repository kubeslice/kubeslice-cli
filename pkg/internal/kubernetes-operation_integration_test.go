//go:build integration
// +build integration

package internal

import (
	"os"
	"os/exec"
	"testing"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyKubectlManifestIntegration(t *testing.T) {
	if !isBinaryAvailable(t, "kind") || !isBinaryAvailable(t, "kubectl") {
		t.SkipNow()
	}

	clusterName := "kubeslice-itest-kubectl"
	createTestKindClusters(t, []string{clusterName})

	originalFileSystem := util.FileSystem
	originalSetEnv := setEnvFunc
	setupKubeconfigForCluster(t, clusterName)

	originalExecutor := util.CommandExecutor
	cleanupEnv := util.NewTestEnvironment()
	util.CommandExecutor = originalExecutor
	fakeOutput := util.Output.(*util.FakeOutput)

	defer func() {
		cleanupEnv()
		util.FileSystem = originalFileSystem
		setEnvFunc = originalSetEnv
		cleanupKindClusters(t, []string{clusterName})
		os.RemoveAll("test-manifest.yaml")
		os.RemoveAll(kubesliceDirectory)
		os.Unsetenv("KUBECONFIG")
	}()

	contextName := "kind-" + clusterName
	cluster := &Cluster{
		Name:           clusterName,
		ContextName:    contextName,
		KubeConfigPath: KubeconfigPath,
	}

	manifest := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value`

	err := os.WriteFile("test-manifest.yaml", []byte(manifest), 0644)
	require.NoError(t, err)

	ApplyKubectlManifest("test-manifest.yaml", "default", cluster)

	require.Empty(t, fakeOutput.FatalCalls)

	cmd := exec.Command(util.ExecutablePaths["kubectl"], "--context", contextName,
		"--kubeconfig", KubeconfigPath, "get", "configmap", "test-config", "-n", "default")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "ConfigMap should exist after apply")
	assert.Contains(t, string(output), "test-config")
}

func TestGetKubectlResourcesIntegration(t *testing.T) {
	if !isBinaryAvailable(t, "kind") || !isBinaryAvailable(t, "kubectl") {
		t.SkipNow()
	}

	clusterName := "kubeslice-itest-get"
	createTestKindClusters(t, []string{clusterName})

	originalFileSystem := util.FileSystem
	originalSetEnv := setEnvFunc
	setupKubeconfigForCluster(t, clusterName)

	originalExecutor := util.CommandExecutor
	cleanupEnv := util.NewTestEnvironment()
	util.CommandExecutor = originalExecutor
	fakeOutput := util.Output.(*util.FakeOutput)

	defer func() {
		cleanupEnv()
		util.FileSystem = originalFileSystem
		setEnvFunc = originalSetEnv
		cleanupKindClusters(t, []string{clusterName})
		os.RemoveAll(kubesliceDirectory)
		os.Unsetenv("KUBECONFIG")
	}()

	contextName := "kind-" + clusterName
	cluster := &Cluster{
		Name:           clusterName,
		ContextName:    contextName,
		KubeConfigPath: KubeconfigPath,
	}

	GetKubectlResources("pods", "", "kube-system", cluster, "")

	require.Empty(t, fakeOutput.FatalCalls)
}

func TestDeleteKubectlResourcesIntegration(t *testing.T) {
	if !isBinaryAvailable(t, "kind") || !isBinaryAvailable(t, "kubectl") {
		t.SkipNow()
	}

	clusterName := "kubeslice-itest-delete"
	createTestKindClusters(t, []string{clusterName})

	originalFileSystem := util.FileSystem
	originalSetEnv := setEnvFunc
	setupKubeconfigForCluster(t, clusterName)

	originalExecutor := util.CommandExecutor
	cleanupEnv := util.NewTestEnvironment()
	util.CommandExecutor = originalExecutor
	fakeOutput := util.Output.(*util.FakeOutput)

	defer func() {
		cleanupEnv()
		util.FileSystem = originalFileSystem
		setEnvFunc = originalSetEnv
		cleanupKindClusters(t, []string{clusterName})
		os.RemoveAll("test-delete.yaml")
		os.RemoveAll(kubesliceDirectory)
		os.Unsetenv("KUBECONFIG")
	}()

	contextName := "kind-" + clusterName
	cluster := &Cluster{
		Name:           clusterName,
		ContextName:    contextName,
		KubeConfigPath: KubeconfigPath,
	}

	manifest := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-delete-config
data:
  key: value`

	err := os.WriteFile("test-delete.yaml", []byte(manifest), 0644)
	require.NoError(t, err)

	ApplyKubectlManifest("test-delete.yaml", "default", cluster)
	require.Empty(t, fakeOutput.FatalCalls)

	DeleteKubectlResources("configmap", "test-delete-config", "default", cluster)
	require.Empty(t, fakeOutput.FatalCalls)

	cmd := exec.Command(util.ExecutablePaths["kubectl"], "--context", contextName,
		"--kubeconfig", KubeconfigPath, "get", "configmap", "test-delete-config", "-n", "default")
	output, err := cmd.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(output), "not found")
}

func TestDescribeKubectlResourcesIntegration(t *testing.T) {
	if !isBinaryAvailable(t, "kind") || !isBinaryAvailable(t, "kubectl") {
		t.SkipNow()
	}

	clusterName := "kubeslice-itest-describe"
	createTestKindClusters(t, []string{clusterName})

	originalFileSystem := util.FileSystem
	originalSetEnv := setEnvFunc
	setupKubeconfigForCluster(t, clusterName)

	originalExecutor := util.CommandExecutor
	cleanupEnv := util.NewTestEnvironment()
	util.CommandExecutor = originalExecutor
	fakeOutput := util.Output.(*util.FakeOutput)

	defer func() {
		cleanupEnv()
		util.FileSystem = originalFileSystem
		setEnvFunc = originalSetEnv
		cleanupKindClusters(t, []string{clusterName})
		os.RemoveAll(kubesliceDirectory)
		os.Unsetenv("KUBECONFIG")
	}()

	contextName := "kind-" + clusterName
	cluster := &Cluster{
		Name:           clusterName,
		ContextName:    contextName,
		KubeConfigPath: KubeconfigPath,
	}

	DescribeKubectlResources("namespace", "kube-system", "", cluster)

	require.Empty(t, fakeOutput.FatalCalls)
}
