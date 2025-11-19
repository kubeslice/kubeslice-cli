//go:build integration
// +build integration

package internal

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherNetworkInformation_KindPath(t *testing.T) {
	if !isBinaryAvailable(t, "kind") || !isBinaryAvailable(t, "docker") {
		t.SkipNow()
	}

	const clusterName = "kubeslice-itest-net-kind"

	createTestKindClusters(t, []string{clusterName})
	t.Cleanup(func() { cleanupKindClusters(t, []string{clusterName}) })

	config := &ConfigurationSpecs{
		Configuration: Configuration{
			ClusterConfiguration: ClusterConfiguration{
				ClusterType:       "kind",
				ControllerCluster: Cluster{Name: clusterName},
			},
		},
	}

	originalExecutor := util.CommandExecutor
	cleanupEnv := util.NewTestEnvironment()
	util.CommandExecutor = originalExecutor
	fakeOutput := util.Output.(*util.FakeOutput)
	defer cleanupEnv()

	GatherNetworkInformation(config)

	require.Empty(t, fakeOutput.FatalCalls, "GatherNetworkInformation should not fail")

	ip := config.Configuration.ClusterConfiguration.ControllerCluster.NodeIP
	addr := config.Configuration.ClusterConfiguration.ControllerCluster.ControlPlaneAddress

	assert.NotEmpty(t, ip, "NodeIP should be populated")
	assert.True(t, strings.HasPrefix(ip, "172.") || strings.HasPrefix(ip, "10."), "Kind IP should be valid (172.x or 10.x), got: %s", ip)
	assert.Equal(t, "https://"+ip+":6443", addr, "ControlPlaneAddress should be derived from NodeIP")
}

func TestGatherNetworkInformation_NonKindPath(t *testing.T) {
	if !isBinaryAvailable(t, "kind") || !isBinaryAvailable(t, "kubectl") {
		t.SkipNow()
	}

	const clusterName = "kubeslice-itest-net-nonkind"
	const contextName = "kind-" + clusterName

	createTestKindClusters(t, []string{clusterName})

	originalExecutor := util.CommandExecutor
	originalFileSystem := util.FileSystem
	originalSetEnv := setEnvFunc

	cleanupEnv := util.NewTestEnvironment()
	fakeOutput := util.Output.(*util.FakeOutput)

	util.CommandExecutor = originalExecutor
	util.FileSystem = originalFileSystem
	setEnvFunc = os.Setenv

	t.Cleanup(func() {
		cleanupEnv()
		setEnvFunc = originalSetEnv
		os.Unsetenv("KUBECONFIG")
		os.RemoveAll(kubesliceDirectory)
		cleanupKindClusters(t, []string{clusterName})
	})

	util.CreateDirectoryPath(kubesliceDirectory)
	CreateKubeConfig()
	SetKubeConfigPath()

	cmd := exec.Command(util.ExecutablePaths["kind"], "get", "kubeconfig", "--name", clusterName)
	kubeconfigBytes, err := cmd.Output()
	require.NoError(t, err, "Failed to get kubeconfig from kind")

	err = os.WriteFile(KubeconfigPath, kubeconfigBytes, 0644)
	require.NoError(t, err, "Failed to write real kubeconfig to test file")

	config := &ConfigurationSpecs{
		Configuration: Configuration{
			ClusterConfiguration: ClusterConfiguration{
				ClusterType: "eks",
				ControllerCluster: Cluster{
					Name:           "controller",
					ContextName:    contextName,
					KubeConfigPath: KubeconfigPath,
				},
			},
		},
	}

	GatherNetworkInformation(config)

	require.Empty(t, fakeOutput.FatalCalls, "GatherNetworkInformation should not fail")

	ip := config.Configuration.ClusterConfiguration.ControllerCluster.NodeIP
	addr := config.Configuration.ClusterConfiguration.ControllerCluster.ControlPlaneAddress

	assert.NotEmpty(t, ip, "NodeIP should be populated by kubectl")
	assert.True(t, strings.HasPrefix(ip, "172.") || strings.HasPrefix(ip, "10."), "Kind Node IP should be valid, got: %s", ip)

	assert.NotEmpty(t, addr, "ControlPlaneAddress should be populated by kubectl")
	assert.Contains(t, addr, "https://127.0.0.1:", "Kind Control Plane Address should be populated")
}

func TestGatherNetworkInformation_KindPath_Failure(t *testing.T) {
	if !isBinaryAvailable(t, "docker") {
		t.SkipNow()
	}

	const clusterName = "cluster-does-not-exist"
	config := &ConfigurationSpecs{
		Configuration: Configuration{
			ClusterConfiguration: ClusterConfiguration{
				ClusterType:       "kind",
				ControllerCluster: Cluster{Name: clusterName},
			},
		},
	}

	originalExecutor := util.CommandExecutor
	cleanupEnv := util.NewTestEnvironment()
	util.CommandExecutor = originalExecutor
	fakeOutput := util.Output.(*util.FakeOutput)
	defer cleanupEnv()

	GatherNetworkInformation(config)

	require.NotEmpty(t, fakeOutput.FatalCalls, "GatherNetworkInformation should call Fatalf")
	assert.Contains(t, fakeOutput.FatalCalls[0], "Failed to run command", "Should fail on docker inspect")
	assert.Contains(t, fakeOutput.FatalCalls[0], "Error:", "Error should be from docker")
}
