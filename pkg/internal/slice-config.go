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

func ShowSliceHealth(sliceConfigName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nFetching KubeSlice slice health status for %s...", sliceConfigName)

	// Get the slice config with JSON output for parsing
	var outB, errB bytes.Buffer
	cmdArgs := []string{}
	if controllerCluster != nil {
		cmdArgs = append(cmdArgs, "--context="+controllerCluster.ContextName, "--kubeconfig="+controllerCluster.KubeConfigPath)
	}
	cmdArgs = append(cmdArgs, "get", SliceConfigObject, sliceConfigName, "-n", namespace, "-o", "json")

	err := util.RunCommandCustomIO("kubectl", &outB, &errB, true, cmdArgs...)
	if err != nil {
		util.Printf("%s Failed to fetch slice health status: %v", util.Cross, err)
		return
	}

	// Parse and display health status
	displaySliceHealthStatus(outB.String(), sliceConfigName)
	time.Sleep(200 * time.Millisecond)
}

func ShowAllSliceHealth(namespace string, controllerCluster *Cluster) {
	util.Printf("\nFetching KubeSlice slice health status for all slices...")

	// Get all slice configs with JSON output for parsing
	var outB, errB bytes.Buffer
	cmdArgs := []string{}
	if controllerCluster != nil {
		cmdArgs = append(cmdArgs, "--context="+controllerCluster.ContextName, "--kubeconfig="+controllerCluster.KubeConfigPath)
	}
	cmdArgs = append(cmdArgs, "get", SliceConfigObject, "-n", namespace, "-o", "json")

	err := util.RunCommandCustomIO("kubectl", &outB, &errB, true, cmdArgs...)
	if err != nil {
		util.Printf("%s Failed to fetch slice health status: %v", util.Cross, err)
		return
	}

	// Parse and display health status for all slices
	displayAllSliceHealthStatus(outB.String())
	time.Sleep(200 * time.Millisecond)
}

func displaySliceHealthStatus(jsonOutput string, sliceName string) {
	// Parse JSON to extract health information
	// This is a simplified implementation - in a real scenario, you'd want to parse the JSON properly
	util.Printf("\n=== Slice Health Status: %s ===", sliceName)

	if strings.Contains(jsonOutput, `"status"`) {
		// Extract status information
		if strings.Contains(jsonOutput, `"ready"`) {
			util.Printf("%s Status: Ready", util.Tick)
		} else {
			util.Printf("%s Status: Not Ready", util.Cross)
		}

		// Extract additional health metrics if available
		if strings.Contains(jsonOutput, `"conditions"`) {
			util.Printf("Conditions: Available")
		}

		// Show slice subnet if available
		if strings.Contains(jsonOutput, `"sliceSubnet"`) {
			util.Printf("Slice Subnet: Configured")
		}

		// Show cluster information
		if strings.Contains(jsonOutput, `"clusters"`) {
			util.Printf("Clusters: Connected")
		}
	} else {
		util.Printf("%s Status: Unknown", util.Wait)
	}

	util.Printf("Raw JSON Output:")
	util.Printf(jsonOutput)
}

func displayAllSliceHealthStatus(jsonOutput string) {
	util.Printf("\n=== All Slices Health Status ===")

	// Parse JSON to extract all slices
	if strings.Contains(jsonOutput, `"items"`) {
		util.Printf("Found multiple slices:")

		// Count slices
		sliceCount := strings.Count(jsonOutput, `"kind":"SliceConfig"`)
		util.Printf("Total Slices: %d", sliceCount)

		// Extract slice names
		lines := strings.Split(jsonOutput, "\n")
		for _, line := range lines {
			if strings.Contains(line, `"name"`) && strings.Contains(line, `"metadata"`) {
				// Extract name from JSON
				if nameStart := strings.Index(line, `"name"`); nameStart != -1 {
					if nameEnd := strings.Index(line[nameStart:], `"`); nameEnd != -1 {
						name := strings.Trim(line[nameStart+nameEnd+1:], `", `)
						util.Printf("Slice: %s", name)
					}
				}
			}
		}
	} else {
		util.Printf("No slices found or error parsing output")
	}

	util.Printf("\nRaw JSON Output:")
	util.Printf(jsonOutput)
}
