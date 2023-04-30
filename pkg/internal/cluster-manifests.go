package internal

import (
	"fmt"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const (
	clusterRegistrationFileName = "cluster-registration.yaml"
)

const clusterRegistrationTemplate = `
apiVersion: controller.kubeslice.io/v1alpha1
kind: Cluster
metadata:
  name: %s 
  namespace: %s
spec:
  clusterProperty: %s
---

`
const regionTemplate1 = `
    geoLocation:
      cloudProvider: GCP
      cloudRegion: custom
      latitude: "36.7783"
      longitude: "-119.4179"
`
const regionTemplate2 = `
    geoLocation:
      cloudProvider: DATACENTER
      cloudRegion: custom
      latitude: "40.6976633"
      longitude: "-74.1201054"
`

var regionTemplates = map[string]string{
	"ks-w-1": regionTemplate1,
	"ks-w-2": regionTemplate2,
}

func RegisterWorkerClusters(ApplicationConfiguration *ConfigurationSpecs, cliOptions *CliOptionsStruct) {
	util.Printf("\nRegistering Worker Clusters with Project...")

	if cliOptions != nil {
		if cliOptions.FileName == "" {
			cliOptions.FileName = kubesliceDirectory + "/" + "custom-" + clusterRegistrationFileName
			generateClusterRegistrationManifest(ApplicationConfiguration, cliOptions.FileName, cliOptions.Namespace)
		}
		util.Printf("%s Generated cluster registration manifest %s", util.Tick, cliOptions.FileName)
		time.Sleep(200 * time.Millisecond)
		ApplyKubectlManifest(cliOptions.FileName, cliOptions.Namespace, cliOptions.Cluster)
		util.Printf("%s Applied %s", util.Tick, cliOptions.FileName)
		time.Sleep(200 * time.Millisecond)
	} else {
		ac := ApplicationConfiguration.Configuration
		generateClusterRegistrationManifest(ApplicationConfiguration, kubesliceDirectory+"/"+clusterRegistrationFileName, "kubeslice-"+ac.KubeSliceConfiguration.ProjectName)
		util.Printf("%s Generated cluster registration manifest %s", util.Tick, clusterRegistrationFileName)
		time.Sleep(200 * time.Millisecond)

		ApplyKubectlManifest(kubesliceDirectory+"/"+clusterRegistrationFileName, "kubeslice-"+ac.KubeSliceConfiguration.ProjectName, &ac.ClusterConfiguration.ControllerCluster)
		util.Printf("%s Applied %s", util.Tick, clusterRegistrationFileName)
		time.Sleep(200 * time.Millisecond)
	}
	util.Printf("Registered Worker Clusters with Project.")
}

func generateClusterRegistrationManifest(ApplicationConfiguration *ConfigurationSpecs, filename string, namespace string) {
	var clusterRegistrationContent = ""
	var regionTemplate = "{}"
	if namespace == "" {
		namespace = "kubeslice-" + ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName
	}
	for _, cluster := range ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters {
		if ApplicationConfiguration.Configuration.ClusterConfiguration.Profile == ProfileEntDemo {
			regionTemplate = regionTemplates[cluster.Name]
		}
		clusterRegistrationContent = clusterRegistrationContent + fmt.Sprintf(clusterRegistrationTemplate, cluster.Name, namespace, regionTemplate)
	}
	util.DumpFile(clusterRegistrationContent, filename)
}

func GetKubeSliceCluster(clusterName string, namespace string, controllerCluster *Cluster, outputFormat string) {
	util.Printf("\nFetching KubeSlice Worker...")
	GetKubectlResources(ClusterObject, clusterName, namespace, controllerCluster, outputFormat)
	time.Sleep(200 * time.Millisecond)
}

func DeleteKubeSliceCluster(clusterName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDeleting KubeSlice Worker...")
	DeleteKubectlResources(ClusterObject, clusterName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}

func EditKubeSliceCluster(clusterName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nEditing KubeSlice Worker...")
	EditKubectlResources(ClusterObject, clusterName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}

func DescribeKubeSliceCluster(clusterName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDescribe KubeSlice Worker...")
	DescribeKubectlResources(ClusterObject, clusterName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}
