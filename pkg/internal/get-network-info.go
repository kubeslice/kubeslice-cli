package internal

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

var (
	getAllClustersFunc = getAllClusters
)

func GatherNetworkInformation(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nFetching Network Address for Clusters...")

	if ApplicationConfiguration.Configuration.ClusterConfiguration.Profile == "" && ApplicationConfiguration.Configuration.ClusterConfiguration.ClusterType != "kind" {
		setControlPlaneAddress(&ApplicationConfiguration.Configuration.ClusterConfiguration)
		setNodeIP(&ApplicationConfiguration.Configuration.ClusterConfiguration)
	} else {
		setNodeIPForKindClusters(&ApplicationConfiguration.Configuration.ClusterConfiguration)
	}

	util.Printf("Successfully fetched network addresses for clusters.")
}

func setNodeIPForKindClusters(clusterConfig *ClusterConfiguration) {
	clusters := getAllClustersFunc(clusterConfig)

	for _, cluster := range clusters {
		ip := runDockerInspectForNodeIP(cluster.Name)
		cluster.NodeIP = ip
		cluster.ControlPlaneAddress = "https://" + ip + ":6443"
		util.Printf("%s Fetched Network Address for %s : %s", util.Tick, cluster.Name, ip)
		util.Sleep(200 * time.Millisecond)
	}
}

func runDockerInspectForNodeIP(clusterName string) string {
	var outB, errB bytes.Buffer
	err := util.CommandExecutor.ExecuteWithOutput("docker", &outB, &errB, "inspect", "--format={{.NetworkSettings.Networks.kind.IPAddress}}", fmt.Sprintf("%s-control-plane", clusterName))
	if err != nil {
		util.Fatalf("%s Failed to run command\nOutput: %s\nError: %s %v", util.Cross, outB.String(), errB.String(), err)
	}
	return strings.TrimSpace(outB.String())
}

func setControlPlaneAddress(clusterConfig *ClusterConfiguration) {
	for _, cluster := range getAllClustersFunc(clusterConfig) {
		if cluster.ControlPlaneAddress == "" {
			ip := _getControlPlaneAddress(cluster)
			cluster.ControlPlaneAddress = ip
			util.Printf("%s Control Plane Address fetched %s for %s", util.Tick, cluster.ControlPlaneAddress, cluster.Name)
		}
	}
}

func _getControlPlaneAddress(cluster *Cluster) string {
	var outB, errB bytes.Buffer
	err := util.CommandExecutor.ExecuteWithOutput("kubectl", &outB, &errB, "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "config", "view", "--minify=true", "-o", "jsonpath={.clusters[0].cluster.server}")
	if err != nil {
		util.Fatalf("%s Failed to run command\nOutput: %s\nError: %s %v", util.Cross, outB.String(), errB.String(), err)
	}
	return outB.String()
}

func setNodeIP(clusterConfig *ClusterConfiguration) {
	for _, cluster := range getAllClustersFunc(clusterConfig) {
		if cluster.NodeIP == "" {
			ip := _getNodeIP(cluster)
			cluster.NodeIP = ip
			util.Printf("%s Node IP fetched %s for %s", util.Tick, cluster.NodeIP, cluster.Name)
		}
	}
}

func _getNodeIP(cluster *Cluster) string {
	var outB, errB bytes.Buffer
	err := util.CommandExecutor.ExecuteWithOutput("kubectl", &outB, &errB, "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "get", "nodes", "-o", "jsonpath={\"ExternalIP=\"}{.items[0].status.addresses[?(@.type==\"ExternalIP\")].address}{\"\\n\"}{\"InternalIP=\"}{.items[0].status.addresses[?(@.type==\"InternalIP\")].address}")
	if err != nil {
		util.Fatalf("%s Failed to run command\nOutput: %s\nError: %s %v", util.Cross, outB.String(), errB.String(), err)
	}

	for _, s := range strings.Split(outB.String(), "\n") {
		splits := strings.Split(s, "=")
		if len(splits) > 1 && strings.TrimSpace(splits[1]) != "" {
			return strings.TrimSpace(splits[1])
		}
	}
	return ""
}
