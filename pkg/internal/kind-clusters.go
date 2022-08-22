package internal

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kubeslice/slicectl/util"
)

const KubeconfigPath = kubesliceDirectory + "/kubeconfig.yaml"

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
			time.Sleep(200 * time.Millisecond)
		}
	}
	if !created {
		util.Printf("\nKind clusters already exist... Skipping\n")
	} else {
		util.Printf("Created required kind clusters")
	}
}

func SetKubeConfigPath() {
	os.Setenv("KUBECONFIG", KubeconfigPath)
}

func CreateKubeConfig() {
	if _, err := os.Stat(KubeconfigPath); errors.Is(err, os.ErrNotExist) {
		util.DumpFile("", KubeconfigPath)
		util.Printf("%s Created Empty KubeConfig file : %s", util.Tick, KubeconfigPath)
		time.Sleep(200 * time.Millisecond)
	}
}

func getExistingClusters(clusters []*Cluster) []bool {
	result := make([]bool, len(clusters), len(clusters))
	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("kind", &outB, &errB, true, "get", "clusters")
	if err != nil {
		log.Fatalf("Process failed %v", err)
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
	err := util.RunCommandOnStdIO("kind", "create", "cluster", fmt.Sprintf("--config=%s/%s/%s", kubesliceDirectory, kindSubDirectory, configFile))
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func DeleteKindClusters(ApplicationConfiguration *ConfigurationSpecs) {
	clusters := getAllClusters(&ApplicationConfiguration.Configuration.ClusterConfiguration)
	existingClusters := getExistingClusters(clusters)
	args := make([]string, 0, 0)
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
	err := util.RunCommand("kind", args...)
	if err != nil {
		log.Fatalf("Process failed %v", err)
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
