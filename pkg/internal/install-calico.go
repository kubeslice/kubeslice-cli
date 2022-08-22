package internal

import (
	"bytes"
	"log"
	"strings"
	"time"

	"github.com/kubeslice/slicectl/util"
)

func InstallCalico(clusterConfig *ClusterConfiguration) {
	util.Printf("\nInstalling Calico Networking...")

	clusters := getAllClusters(clusterConfig)
	for _, cluster := range clusters {
		if !calicoAlreadyInstalled(cluster) {
			util.Printf("Installing on Cluster %s", cluster.Name)
			installCalicoOperatorPrerequisites(cluster)
			util.Printf("%s Successfully applied Calico Operator Prerequisites on Cluster %s", util.Tick, cluster.Name)
			time.Sleep(200 * time.Millisecond)

			createCalicoOperator(cluster)
			util.Printf("%s Successfully installed Calico Operator on Cluster %s", util.Tick, cluster.Name)
			time.Sleep(200 * time.Millisecond)

			util.Printf("%s Waiting for Calico Pods to be Healthy on Cluster %s...", util.Wait, cluster.Name)
			PodVerification("Waiting for Calico Pods to be Healthy", *cluster, "calico-system")
		}
	}

	util.Printf("%s Successfully installed Calico Networking", util.Tick)
}

func calicoAlreadyInstalled(cluster *Cluster) bool {
	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, true, "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "get", "namespace", "calico-system")
	if err != nil {
		if strings.Contains(errB.String(), "NotFound") {
			return false
		}
	}
	PodVerification("Waiting for Calico Pods to be Healthy", *cluster, "calico-system")
	util.Printf("%s Calico Networking already present on cluster %s", util.Tick, cluster.Name)
	return true
}

func installCalicoOperatorPrerequisites(cluster *Cluster) {
	err := util.RunCommand("kubectl", "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "create", "-f", "https://raw.githubusercontent.com/projectcalico/calico/v3.24.0/manifests/tigera-operator.yaml")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func createCalicoOperator(cluster *Cluster) {
	err := util.RunCommand("kubectl", "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "create", "-f", "https://raw.githubusercontent.com/projectcalico/calico/v3.24.0/manifests/custom-resources.yaml")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}
