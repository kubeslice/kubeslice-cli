package internal

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kubeslice/slicectl/util"
)

func GatherNetworkInformation(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nFetching Network Address for Clusters...")

	if ApplicationConfiguration.Configuration.ClusterConfiguration.Profile == "" {
		setControlPlaneAddress(&ApplicationConfiguration.Configuration.ClusterConfiguration)
		setNodeIP(&ApplicationConfiguration.Configuration.ClusterConfiguration)
	} else {
		setNodeIPForKindClusters(&ApplicationConfiguration.Configuration.ClusterConfiguration)
	}

	util.Printf("Successfully fetched network addresses for clusters.")
}

func setNodeIPForKindClusters(clusterConfig *ClusterConfiguration) {
	clusters := getAllClusters(clusterConfig)
	for _, cluster := range clusters {
		ip := runDockerInspectForNodeIP(cluster.Name)
		cluster.NodeIP = ip
		cluster.ControlPlaneAddress = "https://" + ip + ":6443"
		util.Printf("%s Fetched Network Address for %s : %s", util.Tick, cluster.Name, ip)
		time.Sleep(200 * time.Millisecond)

	}
}

func runDockerInspectForNodeIP(clusterName string) string {
	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("docker", &outB, &errB, true, "inspect", "--format={{.NetworkSettings.Networks.kind.IPAddress}}", fmt.Sprintf("%s-control-plane", clusterName))
	if err != nil {
		util.Printf("%s Failed to run command\nOutput: %s\nError: %s %v", util.Cross, outB.String(), errB.String(), err)
		os.Exit(1)
	}
	return strings.TrimSpace(outB.String())
}

func setControlPlaneAddress(clusterConfig *ClusterConfiguration) {
	for _, cluster := range getAllClusters(clusterConfig) {
		if cluster.ControlPlaneAddress == "" {
			ip := _getControlPlaneAddress(cluster)
			cluster.ControlPlaneAddress = ip
			util.Printf("%s Control Plane Address fetched %s for %s", util.Tick, cluster.ControlPlaneAddress, cluster.Name)
		}
	}
}

func _getControlPlaneAddress(cluster *Cluster) string {
	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, true, "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "config", "view", "--minify=true", "-o", "jsonpath={.clusters[0].cluster.server}")
	if err != nil {
		util.Printf("%s Failed to run command\nOutput: %s\nError: %s %v", util.Cross, outB.String(), errB.String(), err)
		os.Exit(1)
	}
	return outB.String()
}

func setNodeIP(clusterConfig *ClusterConfiguration) {
	for _, cluster := range getAllClusters(clusterConfig) {
		if cluster.NodeIP == "" {
			ip := _getNodeIP(cluster)
			cluster.NodeIP = ip
			util.Printf("%s Node IP fetched %s for %s", util.Tick, cluster.NodeIP, cluster.Name)
		}
	}
}

func _getNodeIP(cluster *Cluster) string {
	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, true, "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "get", "nodes", "-o", "jsonpath={\"ExternalIP=\"}{.items[0].status.addresses[?(@.type==\"ExternalIP\")].address}{\"\\n\"}{\"InternalIP=\"}{.items[0].status.addresses[?(@.type==\"InternalIP\")].address}")
	if err != nil {
		util.Printf("%s Failed to run command\nOutput: %s\nError: %s %v", util.Cross, outB.String(), errB.String(), err)
		os.Exit(1)
	}
	for _, s := range strings.Split(outB.String(), "\n") {
		splits := strings.Split(s, "=")
		if strings.TrimSpace(splits[1]) != "" {
			return strings.TrimSpace(splits[1])
		}
	}
	return ""
}
