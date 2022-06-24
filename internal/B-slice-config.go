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
  name: demo
  namespace: kubeslice-%s
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

func GenerateSliceConfiguration() {
	util.Printf("\nGenerating Slice Configuration to %s directory", kubesliceDirectory)

	wc := ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters
	clusters := make([]string, 0, len(wc))
	for _, cluster := range wc {
		clusters = append(clusters, cluster.Name)
	}
	clusterString := strings.Join(clusters, ",")
	util.DumpFile(fmt.Sprintf(sliceTemplate, ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName, clusterString), kubesliceDirectory+"/"+sliceTemplateFileName)
	util.Printf("%s Generated %s", util.Tick, sliceTemplateFileName)
	time.Sleep(200 * time.Millisecond)

	util.Printf("Generated Slice Configuration")
}

func ApplySliceConfiguration() {
	util.Printf("\nApplying Slice Manifest %s to %s cluster", sliceTemplateFileName, ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster.Name)

	ApplyKubectlManifest(kubesliceDirectory+"/"+sliceTemplateFileName, "kubeslice-demo", ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster)

	util.Printf("\nSuccessfully Applied Slice Configuration.")
}
