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
  namespace: kubeslice-%s
spec:
  networkInterface: eth0
---

`

func RegisterWorkerClusters(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nRegistering Worker Clusters with Project...")

	generateClusterRegistrationManifest(ApplicationConfiguration)
	util.Printf("%s Generated cluster registration manifest %s", util.Tick, clusterRegistrationFileName)
	time.Sleep(200 * time.Millisecond)

	ac := ApplicationConfiguration.Configuration
	ApplyKubectlManifest(kubesliceDirectory+"/"+clusterRegistrationFileName, "kubeslice-"+ac.KubeSliceConfiguration.ProjectName, ac.ClusterConfiguration.ControllerCluster)
	util.Printf("%s Applied %s", util.Tick, clusterRegistrationFileName)
	time.Sleep(200 * time.Millisecond)

	util.Printf("Created KubeSlice Project.")
}

func generateClusterRegistrationManifest(ApplicationConfiguration *ConfigurationSpecs) {
	var clusterRegistrationContent = ""
	for _, cluster := range ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters {
		clusterRegistrationContent = clusterRegistrationContent + fmt.Sprintf(clusterRegistrationTemplate, cluster.Name, ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName)
	}
	util.DumpFile(clusterRegistrationContent, kubesliceDirectory+"/"+clusterRegistrationFileName)
}
