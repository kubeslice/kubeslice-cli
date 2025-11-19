package internal

import (
	"bytes"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const (
	calicoOperatorURL       = "https://raw.githubusercontent.com/projectcalico/calico/v3.24.0/manifests/tigera-operator.yaml"
	calicoCustomResourceURL = "https://raw.githubusercontent.com/projectcalico/calico/v3.24.0/manifests/custom-resources.yaml"
	calicoNamespace         = "calico-system"
)

// Function variables for testing
var (
	podVerificationFuncCalico = PodVerification
)

func InstallCalico(clusterConfig *ClusterConfiguration) {
	util.Printf("\nInstalling Calico Networking...")

	clusters := getAllClusters(clusterConfig)
	for _, cluster := range clusters {
		if !calicoAlreadyInstalled(cluster) {
			util.Printf("Installing on Cluster %s", cluster.Name)
			installCalicoOperatorPrerequisites(cluster)
			util.Printf("%s Successfully applied Calico Operator Prerequisites on Cluster %s", util.Tick, cluster.Name)
			util.Sleep(200 * time.Millisecond)

			createCalicoOperator(cluster)
			util.Printf("%s Successfully installed Calico Operator on Cluster %s", util.Tick, cluster.Name)
			util.Sleep(200 * time.Millisecond)

			util.Printf("%s Waiting for Calico Pods to be Healthy on Cluster %s...", util.Wait, cluster.Name)
			podVerificationFuncCalico("Waiting for Calico Pods to be Healthy", *cluster, calicoNamespace)
		}
	}

	util.Printf("%s Successfully installed Calico Networking", util.Tick)
}

func calicoAlreadyInstalled(cluster *Cluster) bool {
	var outB, errB bytes.Buffer
	err := util.CommandExecutor.ExecuteWithOutput("kubectl", &outB, &errB, "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "get", "namespace", calicoNamespace)

	if err != nil {
		if strings.Contains(errB.String(), "NotFound") {
			return false
		}
	}

	podVerificationFuncCalico("Waiting for Calico Pods to be Healthy", *cluster, calicoNamespace)
	util.Printf("%s Calico Networking already present on cluster %s", util.Tick, cluster.Name)
	return true
}

func installCalicoOperatorPrerequisites(cluster *Cluster) {
	err := util.CommandExecutor.Execute("kubectl", "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "create", "-f", calicoOperatorURL)
	if err != nil {
		util.Fatalf("Process failed %v", err)
	}
}

func createCalicoOperator(cluster *Cluster) {
	err := util.CommandExecutor.Execute("kubectl", "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "create", "-f", calicoCustomResourceURL)
	if err != nil {
		util.Fatalf("Process failed %v", err)
	}
}
