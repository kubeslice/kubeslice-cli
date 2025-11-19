//go:build integration
// +build integration

package internal

import (
	"os"
	"strings"
	"testing"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateKindClustersIntegration(t *testing.T) {
	if !isBinaryAvailable(t, "kind") {
		t.SkipNow()
	}
	tests := []struct {
		name         string
		config       *ConfigurationSpecs
		clusterNames []string
	}{
		{
			name: "Create controller cluster only",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "kubeslice-itest-ctrl"},
						WorkerClusters:    []Cluster{},
					},
				},
			},
			clusterNames: []string{"kubeslice-itest-ctrl"},
		},
		{
			name: "Create controller and worker clusters",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "kubeslice-itest-ctrl-2"},
						WorkerClusters: []Cluster{
							{Name: "kubeslice-itest-w1"},
							{Name: "kubeslice-itest-w2"},
						},
					},
				},
			},
			clusterNames: []string{"kubeslice-itest-ctrl-2", "kubeslice-itest-w1", "kubeslice-itest-w2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalExecutor := util.CommandExecutor
			originalFileSystem := util.FileSystem
			cleanupEnv := util.NewTestEnvironment()
			util.CommandExecutor = originalExecutor
			util.FileSystem = originalFileSystem
			fakeOutput := util.Output.(*util.FakeOutput)
			defer cleanupEnv()

			t.Cleanup(func() {
				cleanupKindClusters(t, tt.clusterNames)
				os.RemoveAll(kubesliceDirectory)
			})

			util.CreateDirectoryPath(kubesliceDirectory + "/" + kindSubDirectory)
			for _, name := range tt.clusterNames {
				kindConfig := `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: ` + name
				util.DumpFile(kindConfig, kubesliceDirectory+"/"+kindSubDirectory+"/"+name+".yaml")
			}

			CreateKindClusters(tt.config)
			require.Empty(t, fakeOutput.FatalCalls)
			existingClusters := getKindClusterList(t)
			for _, name := range tt.clusterNames {
				assert.Contains(t, existingClusters, name)
			}
		})
	}
}

func TestCreateKindClustersIdempotencyIntegration(t *testing.T) {
	if !isBinaryAvailable(t, "kind") {
		t.SkipNow()
	}

	clusterName := "kubeslice-itest-idempotent"
	config := &ConfigurationSpecs{
		Configuration: Configuration{
			ClusterConfiguration: ClusterConfiguration{
				ControllerCluster: Cluster{Name: clusterName},
			},
		},
	}

	originalExecutor := util.CommandExecutor
	originalFileSystem := util.FileSystem
	cleanupEnv := util.NewTestEnvironment()
	util.CommandExecutor = originalExecutor
	util.FileSystem = originalFileSystem
	fakeOutput := util.Output.(*util.FakeOutput)
	defer cleanupEnv()

	t.Cleanup(func() {
		cleanupKindClusters(t, []string{clusterName})
		os.RemoveAll(kubesliceDirectory)
	})

	util.CreateDirectoryPath(kubesliceDirectory + "/" + kindSubDirectory)
	kindConfig := `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: ` + clusterName
	util.DumpFile(kindConfig, kubesliceDirectory+"/"+kindSubDirectory+"/"+clusterName+".yaml")

	CreateKindClusters(config)
	require.Empty(t, fakeOutput.FatalCalls)
	require.Contains(t, getKindClusterList(t), clusterName)

	fakeOutput.InfoCalls = []string{}

	CreateKindClusters(config)

	require.Empty(t, fakeOutput.FatalCalls)
	infoOutput := strings.Join(fakeOutput.InfoCalls, "\n")
	assert.Contains(t, infoOutput, "Kind clusters already exist... Skipping")
}

func TestDeleteKindClustersIntegration(t *testing.T) {
	if !isBinaryAvailable(t, "kind") {
		t.SkipNow()
	}

	t.Run("Delete existing clusters", func(t *testing.T) {
		clusterNames := []string{"kubeslice-itest-del-1", "kubeslice-itest-del-2"}

		originalExecutor := util.CommandExecutor
		cleanupEnv := util.NewTestEnvironment()
		util.CommandExecutor = originalExecutor
		fakeOutput := util.Output.(*util.FakeOutput)
		defer cleanupEnv()

		config := &ConfigurationSpecs{
			Configuration: Configuration{
				ClusterConfiguration: ClusterConfiguration{
					ControllerCluster: Cluster{Name: clusterNames[0]},
					WorkerClusters: []Cluster{
						{Name: clusterNames[1]},
					},
				},
			},
		}

		createTestKindClusters(t, clusterNames)
		t.Cleanup(func() {
			cleanupKindClusters(t, clusterNames)
		})

		DeleteKindClusters(config)
		require.Empty(t, fakeOutput.FatalCalls)
		existingAfter := getKindClusterList(t)
		for _, name := range clusterNames {
			assert.NotContains(t, existingAfter, name)
		}
	})

	t.Run("Delete non-existent clusters", func(t *testing.T) {
		originalExecutor := util.CommandExecutor
		cleanupEnv := util.NewTestEnvironment()
		util.CommandExecutor = originalExecutor
		fakeOutput := util.Output.(*util.FakeOutput)
		defer cleanupEnv()

		config := &ConfigurationSpecs{
			Configuration: Configuration{
				ClusterConfiguration: ClusterConfiguration{
					ControllerCluster: Cluster{Name: "non-existent-cluster"},
				},
			},
		}

		DeleteKindClusters(config)

		require.Empty(t, fakeOutput.FatalCalls)
		allOutput := strings.Join(fakeOutput.InfoCalls, " ")
		assert.Contains(t, allOutput, "No Kind Clusters found for deletion")
	})
}

func TestCreateKubeConfigIntegration(t *testing.T) {
	originalFileSystem := util.FileSystem
	cleanupEnv := util.NewTestEnvironment()
	util.FileSystem = originalFileSystem
	defer cleanupEnv()

	t.Cleanup(func() {
		os.RemoveAll(kubesliceDirectory)
	})

	util.CreateDirectoryPath(kubesliceDirectory)

	CreateKubeConfig()

	_, err := os.Stat(KubeconfigPath)
	assert.NoError(t, err)
}

func TestSetKubeConfigPathIntegration(t *testing.T) {
	originalSetEnv := setEnvFunc
	cleanupEnv := util.NewTestEnvironment()
	setEnvFunc = os.Setenv
	defer func() {
		cleanupEnv()
		setEnvFunc = originalSetEnv
		os.Unsetenv("KUBECONFIG")
	}()

	SetKubeConfigPath()

	assert.Equal(t, KubeconfigPath, os.Getenv("KUBECONFIG"))
}
