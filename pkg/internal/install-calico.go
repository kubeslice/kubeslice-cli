package internal

import (
	"log"
	"time"

	"github.com/kubeslice/slicectl/util"
)

func InstallCalico(clusterConfig ClusterConfiguration) {
	util.Printf("\nInstalling Calico Networking...")

	clusters := getAllClusters(clusterConfig)
	for _, cluster := range clusters {
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

	util.Printf("%s Successfully installed Calico Networking", util.Tick)
}

func installCalicoOperatorPrerequisites(cluster *Cluster) {
	err := util.RunCommand("kubectl", "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "apply", "-f", "https://projectcalico.docs.tigera.io/manifests/tigera-operator.yaml")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func createCalicoOperator(cluster *Cluster) {
	err := util.RunCommand("kubectl", "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "apply", "-f", "https://projectcalico.docs.tigera.io/manifests/custom-resources.yaml")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}
