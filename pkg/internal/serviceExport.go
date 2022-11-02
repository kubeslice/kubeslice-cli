package internal

import (
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const (
	serviceExportConfigFileName = "serviceExportConfig.yaml"
)

func CreateServiceExportConfig(namespace string, controllerCluster *Cluster, filename string) {
	ApplyFile(filename, namespace, controllerCluster)
	util.Printf("\nSuccessfully Applied Slice Configuration.")
}

func GetServiceExportConfig(serviceExportConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nFetching KubeSlice serviceExportConfig...")
	GetKubectlResources(ServiceExportConfigObject, serviceExportConfigName, namespace, controllerCluster, "")
	time.Sleep(200 * time.Millisecond)
}
func generateServiceExportConfigManifest(serviceExportConfigName string) {
	//util.DumpFile(fmt.Sprintf(ServiceExportConfigTemplate, serviceExportConfigName), kubesliceDirectory+"/"+serviceExportConfigFileName)
}

func DeleteServiceExportConfig(serviceExportConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDeleting KubeSlice serviceExportConfig...")
	DeleteKubectlResources(ServiceExportConfigObject, serviceExportConfigName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}

func EditServiceExportConfig(serviceExportConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nEditing KubeSlice serviceExportConfig...")
	EditKubectlResources(ServiceExportConfigObject, serviceExportConfigName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}

func DescribeServiceExportConfig(serviceExportConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDescribe KubeSlice serviceExportConfig...")
	DescribeKubectlResources(ServiceExportConfigObject, serviceExportConfigName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}
