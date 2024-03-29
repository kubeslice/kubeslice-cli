package internal

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
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
	util.Printf("%s Generated %s", util.Tick, "slice-"+sliceConfigName+".yaml")
	time.Sleep(200 * time.Millisecond)

	util.Printf("Generated Slice Configuration")
}

func ApplySliceConfiguration(ApplicationConfiguration *ConfigurationSpecs) {
	verifyNodeIPsInClusters(ApplicationConfiguration)
	util.Printf("\nApplying Slice Manifest %s to %s cluster", sliceTemplateFileName, ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster.Name)

	ApplyKubectlManifest(kubesliceDirectory+"/"+sliceTemplateFileName, "kubeslice-demo", &ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster)

	util.Printf("\nSuccessfully Applied Slice Configuration.")
}

func verifyNodeIPsInClusters(ApplicationConfiguration *ConfigurationSpecs) {
	var outB, errB bytes.Buffer
	cc := ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster
	wc := ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters
	projectNamespace := "kubeslice-" + ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName
	for _, cluster := range wc {
		util.Printf("%s Waiting for NodeIPs to be populated in %s...", util.Wait, cluster.Name)
		var nodeIPs string
		i := 1 // retry for 50 seconds
		for nodeIPs == "" && i < 11 {
			util.RunCommandCustomIO("kubectl", &outB, &errB, true, "--context="+cc.ContextName, "--kubeconfig="+cc.KubeConfigPath, "get", ClusterObject, cluster.Name, "-n", projectNamespace, "-o", "jsonpath='{.status.nodeIPs}'")
			nodeIPs = outB.String()
			if nodeIPs == "" {
				time.Sleep(5 * time.Second)
				util.Printf("%s Waiting for NodeIPs to be populated in %s... %d seconds elapsed", util.Wait, cluster.Name, i*5)
				i++
			} else {
				util.Printf("%s NodeIPs populated in %s", util.Tick, cluster.Name)
			}
		}
	}

}

func GetSliceConfig(sliceConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nFetching KubeSlice sliceConfig...")
	GetKubectlResources(SliceConfigObject, sliceConfigName, namespace, controllerCluster, "")
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

func CreateSliceConfig(namespace string, controllerCluster *Cluster, filename string) {
	ApplyFile(filename, namespace, controllerCluster)
	util.Printf("\nSuccessfully Applied Slice Configuration.")
}
