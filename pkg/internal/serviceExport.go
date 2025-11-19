package internal

import (
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const (
	serviceExportConfigFileName = "serviceExportConfig.yaml"
)

// Function variables for testing
var (
	applyFileFuncServiceExport                = ApplyFile
	getKubectlResourcesFuncServiceExport      = GetKubectlResources
	deleteKubectlResourcesFuncServiceExport   = DeleteKubectlResources
	editKubectlResourcesFuncServiceExport     = EditKubectlResources
	describeKubectlResourcesFuncServiceExport = DescribeKubectlResources
)

func CreateServiceExportConfig(namespace string, controllerCluster *Cluster, filename string) {
	applyFileFuncServiceExport(filename, namespace, controllerCluster)
	util.Printf("\nSuccessfully Applied Service Export Configuration.")
}

func GetServiceExportConfig(serviceExportConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nFetching KubeSlice serviceExportConfig...")
	getKubectlResourcesFuncServiceExport(ServiceExportConfigObject, serviceExportConfigName, namespace, controllerCluster, "")
	util.SystemClock.Sleep(200 * time.Millisecond)
}

func generateServiceExportConfigManifest(serviceExportConfigName string) {
	//util.DumpFile(fmt.Sprintf(ServiceExportConfigTemplate, serviceExportConfigName), kubesliceDirectory+"/"+serviceExportConfigFileName)
}

func DeleteServiceExportConfig(serviceExportConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDeleting KubeSlice serviceExportConfig...")
	deleteKubectlResourcesFuncServiceExport(ServiceExportConfigObject, serviceExportConfigName, namespace, controllerCluster)
	util.SystemClock.Sleep(200 * time.Millisecond)
}

func EditServiceExportConfig(serviceExportConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nEditing KubeSlice serviceExportConfig...")
	editKubectlResourcesFuncServiceExport(ServiceExportConfigObject, serviceExportConfigName, namespace, controllerCluster)
	util.SystemClock.Sleep(200 * time.Millisecond)
}

func DescribeServiceExportConfig(serviceExportConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDescribe KubeSlice serviceExportConfig...")
	describeKubectlResourcesFuncServiceExport(ServiceExportConfigObject, serviceExportConfigName, namespace, controllerCluster)
	util.SystemClock.Sleep(200 * time.Millisecond)
}
