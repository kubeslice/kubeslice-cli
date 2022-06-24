package internal

import (
	"fmt"
	"time"

	"github.com/kubeslice/slicectl/util"
)

const (
	clusterRegistrationFileName = "cluster-registration.yaml"
)

const clusterRegistrationTemplate = `
apiVersion: controller.kubeslice.io/v1alpha1
kind: Cluster
metadata:
  name: %s 
  namespace: kubeslice-demo
spec:
  networkInterface: eth0
---
apiVersion: controller.kubeslice.io/v1alpha1
kind: Cluster
metadata:
  name: %s
  namespace: kubeslice-demo
spec:
  networkInterface: eth0
`

func RegisterWorkerClusters() {
	util.Printf("\nRegistering Worker Clusters with Project...")

	generateClusterRegistrationManifest()
	util.Printf("%s Generated cluster registration manifest %s", util.Tick, clusterRegistrationFileName)
	time.Sleep(200 * time.Millisecond)

	util.ApplyKubectlManifest(kubesliceDirectory+"/"+clusterRegistrationFileName, "kubeslice-demo", controllerName)
	util.Printf("%s Applied %s", util.Tick, clusterRegistrationFileName)
	time.Sleep(200 * time.Millisecond)

	util.Printf("Created KubeSlice Project.")
}

func generateClusterRegistrationManifest() {
	util.DumpFile(fmt.Sprintf(clusterRegistrationTemplate, worker1Name, worker2Name), kubesliceDirectory+"/"+clusterRegistrationFileName)
}
