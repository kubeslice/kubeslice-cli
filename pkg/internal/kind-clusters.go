package internal

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const KubeconfigPath = kubesliceDirectory + "/kubeconfig.yaml"

// Function variables for testing
var (
	setEnvFunc = os.Setenv
)

func CreateKindClusters(ApplicationConfiguration *ConfigurationSpecs) {

	clusters := getAllClusters(&ApplicationConfiguration.Configuration.ClusterConfiguration)
	existingClusters := getExistingClusters(clusters)
	created := false
	util.Printf("\nCreating Kind Clusters...")
	for i, cluster := range clusters {
		if !existingClusters[i] {
			created = true
			createKindCluster(cluster.Name + ".yaml")
			util.Printf("%s Created Kind Cluster : %s", util.Tick, cluster.Name)
			util.Sleep(200 * time.Millisecond)
		}
	}
	if !created {
		util.Printf("\nKind clusters already exist... Skipping\n")
	} else {
		util.Printf("Created required kind clusters")
	}
}

func SetKubeConfigPath() {
	err := setEnvFunc("KUBECONFIG", KubeconfigPath)
	if err != nil {
		util.Printf("%s Warning: Failed to set KUBECONFIG environment variable: %v", util.Warn, err)
	}
}

func CreateKubeConfig() {
	_, err := util.FileSystem.Stat(KubeconfigPath)
	if errors.Is(err, os.ErrNotExist) {
		util.DumpFile("", KubeconfigPath)
		util.Printf("%s Created Empty KubeConfig file : %s", util.Tick, KubeconfigPath)
		util.Sleep(200 * time.Millisecond)
	}
}

func getExistingClusters(clusters []*Cluster) []bool {
	result := make([]bool, len(clusters))

	var outB, errB bytes.Buffer
	err := util.CommandExecutor.ExecuteWithOutput("kind", &outB, &errB, "get", "clusters")
	if err != nil {
		util.Fatalf("Process failed %v", err)
	}
	for i, cluster := range clusters {
		for _, line := range strings.Split(outB.String(), "\n") {
			if strings.Contains(line, cluster.Name) {
				result[i] = true
			}
		}
	}

	return result
}

func createKindCluster(configFile string) {
	err := util.CommandExecutor.Execute("kind", "create", "cluster", fmt.Sprintf("--config=%s/%s/%s", kubesliceDirectory, kindSubDirectory, configFile))
	if err != nil {
		util.Fatalf("Process failed %v", err)
	}
}

func DeleteKindClusters(ApplicationConfiguration *ConfigurationSpecs) {
	clusters := getAllClusters(&ApplicationConfiguration.Configuration.ClusterConfiguration)
	existingClusters := getExistingClusters(clusters)

	args := make([]string, 0)
	args = append(args, "delete", "clusters")
	cNames := make([]string, 0)
	for i, cluster := range clusters {
		if existingClusters[i] {
			cNames = append(cNames, cluster.Name)
		}
	}
	if len(cNames) == 0 {
		util.Printf("No Kind Clusters found for deletion")
		return
	}
	args = append(args, cNames...)
	err := util.CommandExecutor.Execute("kind", args...)
	if err != nil {
		util.Fatalf("Process failed %v", err)
	}
}

func getAllClusters(clusterConfig *ClusterConfiguration) []*Cluster {
	cc := clusterConfig
	clusters := make([]*Cluster, 0, len(cc.WorkerClusters)+1)
	clusters = append(clusters, &cc.ControllerCluster)
	for i := 0; i < len(cc.WorkerClusters); i++ {
		clusters = append(clusters, &cc.WorkerClusters[i])
	}
	return clusters
}

func getControllerCluster(clusterConfig ClusterConfiguration) *Cluster {
	cc := clusterConfig
	return &cc.ControllerCluster
}
