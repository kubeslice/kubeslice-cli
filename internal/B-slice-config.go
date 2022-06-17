package internal

import (
	"fmt"
	"time"

	"github.com/kubeslice/kubeslice-installer/util"
)

const (
	sliceTemplateFileName = "slice-demo.yaml"
)

const sliceTemplate = `
apiVersion: controller.kubeslice.io/v1alpha1
kind: SliceConfig
metadata:
  name: demo
  namespace: kubeslice-demo
spec:
  sliceSubnet: 10.1.0.0/16
  sliceType: Application
  sliceGatewayProvider:
    sliceGatewayType: OpenVPN
    sliceCaType: Local
  sliceIpamType: Local
  clusters:
    - %s
    - %s
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

	util.DumpFile(fmt.Sprintf(sliceTemplate, worker1Name, worker2Name), kubesliceDirectory+"/"+sliceTemplateFileName)
	util.Printf("%s Generated %s", util.Tick, sliceTemplateFileName)
	time.Sleep(200 * time.Millisecond)

	util.Printf("Generated Slice Configuration")
}

func ApplySliceConfiguration() {
	util.Printf("\nApplying Slice Manifest %s to %s cluster", sliceTemplateFileName, controllerName)

	util.ApplyKubectlManifest(kubesliceDirectory+"/"+sliceTemplateFileName, "kubeslice-demo", controllerName)

	util.Printf("\nSuccessfully Applied Slice Configuration.")
}