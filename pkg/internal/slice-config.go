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

const (
	defaultNodeIPWaitTime   = 5 * time.Second
	defaultNodeIPRetryLimit = 10
)

var (
	applyKubectlManifestFuncSlice     = ApplyKubectlManifest
	getKubectlResourcesFuncSlice      = GetKubectlResources
	deleteKubectlResourcesFuncSlice   = DeleteKubectlResources
	editKubectlResourcesFuncSlice     = EditKubectlResources
	describeKubectlResourcesFuncSlice = DescribeKubectlResources
	verifyNodeIPsInClustersFunc       = verifyNodeIPsInClusters
	applyFileFuncSlice                = ApplyFile
)

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
	util.SystemClock.Sleep(200 * time.Millisecond)
	util.Printf("Generated Slice Configuration")
}

func ApplySliceConfiguration(ApplicationConfiguration *ConfigurationSpecs) {
	verifyNodeIPsInClustersFunc(ApplicationConfiguration)
	util.Printf("\nApplying Slice Manifest %s to %s cluster", sliceTemplateFileName, ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster.Name)
	applyKubectlManifestFuncSlice(kubesliceDirectory+"/"+sliceTemplateFileName, "kubeslice-demo", &ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster)
	util.Printf("\nSuccessfully Applied Slice Configuration.")
}

func verifyNodeIPsInClusters(ApplicationConfiguration *ConfigurationSpecs) {
	cc := ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster
	wc := ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters
	projectNamespace := "kubeslice-" + ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName
	for _, cluster := range wc {
		util.Printf("%s Waiting for NodeIPs to be populated in %s...", util.Wait, cluster.Name)
		var nodeIPs string
		i := 1
		for nodeIPs == "" && i < defaultNodeIPRetryLimit+1 {
			var outB, errB bytes.Buffer
			util.CommandExecutor.ExecuteWithOutput("kubectl", &outB, &errB, "--context="+cc.ContextName, "--kubeconfig="+cc.KubeConfigPath, "get", ClusterObject, cluster.Name, "-n", projectNamespace, "-o", "jsonpath='{.status.nodeIPs}'")
			nodeIPs = outB.String()
			if nodeIPs == "" {
				util.SystemClock.Sleep(defaultNodeIPWaitTime)
				util.Printf("%s Waiting for NodeIPs to be populated in %s... %d seconds elapsed", util.Wait, cluster.Name, i*5)
				i++
			} else {
				util.Printf("%s NodeIPs populated in %s", util.Tick, cluster.Name)
			}
		}
		if nodeIPs == "" {
			util.Printf("%s Warning: NodeIPs not populated in %s after %d seconds", util.Warn, cluster.Name, defaultNodeIPRetryLimit*5)
		}
	}

}

func GetSliceConfig(sliceConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nFetching KubeSlice sliceConfig...")
	getKubectlResourcesFuncSlice(SliceConfigObject, sliceConfigName, namespace, controllerCluster, "")
	util.SystemClock.Sleep(200 * time.Millisecond)
}

func DeleteSliceConfig(sliceConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDeleting KubeSlice SliceConfig...")
	deleteKubectlResourcesFuncSlice(SliceConfigObject, sliceConfigName, namespace, controllerCluster)
	util.SystemClock.Sleep(200 * time.Millisecond)
}

func EditSliceConfig(sliceConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nEditing KubeSlice SliceConfig...")
	editKubectlResourcesFuncSlice(SliceConfigObject, sliceConfigName, namespace, controllerCluster)
	util.SystemClock.Sleep(200 * time.Millisecond)
}

func DescribeSliceConfig(sliceConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDescribing KubeSlice SliceConfig...")
	describeKubectlResourcesFuncSlice(SliceConfigObject, sliceConfigName, namespace, controllerCluster)
	util.SystemClock.Sleep(200 * time.Millisecond)
}

func CreateSliceConfig(namespace string, controllerCluster *Cluster, filename string) {
	applyFileFuncSlice(filename, namespace, controllerCluster)
	util.Printf("\nSuccessfully Applied Slice Configuration.")
}
