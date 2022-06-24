package internal

import (
	"log"
	"time"

	"github.com/kubeslice/slicectl/util"
)

func InstallCalico() {
	util.Printf("\nInstalling Calico Networking...")

	for _, clusterName := range []string{controllerName, worker1Name, worker2Name} {
		util.Printf("Installing on Cluster %s", clusterName)
		installCalicoOperatorPrerequisites(clusterName)
		util.Printf("%s Successfully applied Calico Operator Prerequisites on Cluster %s", util.Tick, clusterName)
		time.Sleep(200 * time.Millisecond)

		createCalicoOperator(clusterName)
		util.Printf("%s Successfully installed Calico Operator on Cluster %s", util.Tick, clusterName)
		time.Sleep(200 * time.Millisecond)

		util.Printf("%s Waiting for Calico Pods to be Healthy on Cluster %s...", util.Wait, clusterName)
		util.PodVerification("Waiting for Calico Pods to be Healthy", clusterName, "calico-system")
	}

	util.Printf("%s Successfully installed Calico Networking", util.Tick)
}

func installCalicoOperatorPrerequisites(clusterName string) {
	err := util.RunCommand("kubectl", "--context=kind-"+clusterName, "apply", "-f", "https://projectcalico.docs.tigera.io/manifests/tigera-operator.yaml")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func createCalicoOperator(clusterName string) {
	err := util.RunCommand("kubectl", "--context=kind-"+clusterName, "apply", "-f", "https://projectcalico.docs.tigera.io/manifests/custom-resources.yaml")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}
