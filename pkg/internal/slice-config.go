package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/kubeslice/slicectl/util"
)

const (
	sliceTemplateFileName = "slice-demo.yaml"
)

const sliceTemplate = `
apiVersion: controller.kubeslice.io/v1alpha1
kind: SliceConfig
metadata:
  name: %s
  namespace: %s
spec:
  sliceSubnet: 10.1.0.0/16
  sliceType: Application
  sliceGatewayProvider:
    sliceGatewayType: OpenVPN
    sliceCaType: Local
  sliceIpamType: Local
  clusters: [%s]
  qosProfileDetails:
    queueType: HTB
    priority: 1
    tcType: BANDWIDTH_CONTROL
    bandwidthCeilingKbps: 5120
    bandwidthGuaranteedKbps: 2560
    dscpClass: AF11
  namespaceIsolationProfile:
   applicationNamespaces:
    - namespace: iperf
      clusters:
      - '*'
`

func GenerateSliceConfiguration(ApplicationConfiguration *ConfigurationSpecs, worker []string, sliceConfigName string, namespace string) {
	util.Printf("\nGenerating Slice Configuration to %s directory", kubesliceDirectory)
	clusters := make([]string, 0)
	if len(worker) != 0 {
		clusters = append(clusters, worker...)
	} else {
		wc := ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters
		for _, cluster := range wc {
			clusters = append(clusters, cluster.Name)
		}
	}
	clusterString := strings.Join(clusters, ",")
	if len(sliceConfigName) == 0 {
		sliceConfigName = "demo"
	}
	projectNamespace := "kubeslice-" + ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName
	if len(namespace) != 0 {
		projectNamespace = namespace
	}
	util.DumpFile(fmt.Sprintf(sliceTemplate, sliceConfigName, projectNamespace, clusterString), kubesliceDirectory+"/"+"slice-"+sliceConfigName+".yaml")
	util.Printf("%s Generated %s", util.Tick, sliceTemplateFileName)
	time.Sleep(200 * time.Millisecond)

	util.Printf("Generated Slice Configuration")
}

func ApplySliceConfiguration(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nApplying Slice Manifest %s to %s cluster", sliceTemplateFileName, ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster.Name)

	ApplyKubectlManifest(kubesliceDirectory+"/"+sliceTemplateFileName, "kubeslice-demo", &ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster)

	util.Printf("\nSuccessfully Applied Slice Configuration.")
}

func GetSliceConfig(sliceConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nFetching KubeSlice sliceConfig...")
	GetKubectlResources(SliceConfigObject, sliceConfigName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}

func DeleteSliceConfig(sliceConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDeleting KubeSlice SliceConfig...")
	DeleteKubectlResources(SliceConfigObject, sliceConfigName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}

func EditSliceConfig(sliceConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nEditing KubeSlice SliceConfig...")
	EditKubectlResources(SliceConfigObject, sliceConfigName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}

func DescribeSliceConfig(sliceConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDescribing KubeSlice SliceConfig...")
	DescribeKubectlResources(SliceConfigObject, sliceConfigName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}

func CreateSliceConfig(sliceConfigName string, namespace string, controllerCluster *Cluster, filename string) {
	ApplyFile(filename, namespace, controllerCluster)
	util.Printf("\nSuccessfully Applied Slice Configuration.")
}
