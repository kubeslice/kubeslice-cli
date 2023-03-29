package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const (
	uiValuesFileName = "helm-values-ui.yaml"
)

const UIValuesTemplate = `
kubeslice:
  uiproxy:
    service: 
      type: %s
`

func InstallKubeSliceUI(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nInstalling KubeSlice Manager...")
	if ApplicationConfiguration.Configuration.HelmChartConfiguration.UIChart.ChartName == "" {
		util.Printf("%s Skipping Kubeslice Manager installaition. UI Helm Chart not found in topology file.", util.Warn)
		return
	}
	cc := ApplicationConfiguration.Configuration.ClusterConfiguration
	hc := ApplicationConfiguration.Configuration.HelmChartConfiguration
	time.Sleep(200 * time.Millisecond)
	clusterType := ApplicationConfiguration.Configuration.ClusterConfiguration.ClusterType
	generateUIValuesFile(clusterType, cc.ControllerCluster, ApplicationConfiguration.Configuration.HelmChartConfiguration)
	installKubeSliceUI(cc.ControllerCluster, hc)
	util.Printf("%s Successfully installed helm chart %s/%s", util.Tick, hc.RepoAlias, hc.UIChart.ChartName)
	time.Sleep(200 * time.Millisecond)
	util.Printf("%s Waiting for KubeSlice Manager Pods to be Healthy...", util.Wait)
	PodVerification("Waiting for KubeSlice Manager Pods to be Healthy", cc.ControllerCluster, "kubernetes-dashboard")
	util.Printf("%s Successfully installed KubeSlice Manager.\n", util.Tick)
}

func UninstallKubeSliceUI(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nUninstalling KubeSlice Manager...")
	cc := ApplicationConfiguration.Configuration.ClusterConfiguration
	time.Sleep(200 * time.Millisecond)
	ok, err := uninstallKubeSliceUI(cc.ControllerCluster)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
	if ok {
		time.Sleep(200 * time.Millisecond)
		util.Printf("%s Successfully uninstalled KubeSlice Manager", util.Tick)
	}
}

func generateUIValuesFile(clusterType string, cluster Cluster, hcConfig HelmChartConfiguration) {
	serviceType := ""
	if clusterType == "kind" {
		serviceType = "NodePort"
	} else {
		serviceType = "LoadBalancer"
	}
	err := generateValuesFile(kubesliceDirectory+"/"+uiValuesFileName, &hcConfig.UIChart, fmt.Sprintf(UIValuesTemplate+generateImagePullSecretsValue(hcConfig.ImagePullSecret), serviceType))
	if err != nil {
		log.Fatalf("%s %s", util.Cross, err)
	}
}

func installKubeSliceUI(cluster Cluster, hc HelmChartConfiguration) {
	args := make([]string, 0)
	args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "upgrade", "-i", "kubeslice-ui", fmt.Sprintf("%s/%s", hc.RepoAlias, hc.UIChart.ChartName), "--namespace", KUBESLICE_CONTROLLER_NAMESPACE, "-f", kubesliceDirectory+"/"+uiValuesFileName)
	if hc.UIChart.Version != "" {
		args = append(args, "--version", hc.UIChart.Version)
	}
	err := util.RunCommand("helm", args...)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func uninstallKubeSliceUI(cluster Cluster) (bool, error) {
	args := make([]string, 0)
	// fetching UI release
	args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "status", "kubeslice-ui", "--namespace", KUBESLICE_CONTROLLER_NAMESPACE)
	err := util.RunCommandWithoutPrint("helm", args...)
	if err != nil {
		util.Printf("%s KubeSlice Manager not installed, skipping uninstall.", util.Cross)
		return false, nil
	} else {
		args = make([]string, 0)
		args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "uninstall", "kubeslice-ui", "--namespace", KUBESLICE_CONTROLLER_NAMESPACE)
		err = util.RunCommand("helm", args...)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func GetUIEndpoint(cc *Cluster) {
	util.Printf("\nFetching KubeSlice Manager Endpoint...")
	ep := ""

	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, true, "--context="+cc.ContextName, "--kubeconfig="+cc.KubeConfigPath, "get", "services", "kubeslice-ui-proxy", "-n", KUBESLICE_CONTROLLER_NAMESPACE, "-o", "jsonpath='{.spec}'")
	if err == nil {
		jsonMap := make(map[string]interface{})
		err = json.Unmarshal(outB.Bytes()[1:len(outB.Bytes())-1], &jsonMap)
		if err != nil {
			util.Printf("%s Unable to parse. Err: %v", util.Cross, err)
		}
		switch jsonMap["type"] {
		case "NodePort":
			ports := jsonMap["ports"].([]interface{})
			for _, port := range ports {
				portMap := port.(map[string]interface{})
				if portMap["name"] == "http" { // Assuming that http is the name of the port that you want to use
					nodePort := int(portMap["nodePort"].(float64))
					nodeIP, err := getNodeIP(cc)
					if err == nil {
						ep = fmt.Sprintf("https://%s:%d", strings.Trim(nodeIP, "'"), nodePort)
					} else {
						util.Printf("%s Unable to get node IP. Err: %v", util.Cross, err)
					}
					break
				}
			}
		case "LoadBalancer":
			if jsonMap["externalIPs"] != nil {
				lbIP := jsonMap["externalIPs"].([]interface{})[0].(string)
				ports := jsonMap["ports"].([]interface{})
				for _, port := range ports {
					portMap := port.(map[string]interface{})
					if portMap["name"] == "http" { // Assuming that http is the name of the port that you want to use
						nodePort := int(portMap["port"].(float64))
						ep = fmt.Sprintf("https://%s:%d", lbIP, nodePort)
						break
					}
				}
			}

		default:
			util.Printf("%s Unsupported service type: %s", util.Cross, jsonMap["type"])
		}
	}
	if err != nil || ep == "" {
		util.Printf("%s Unable to find the endpoint.", util.Cross)
	} else {
		util.Printf("%s Visit %v from your browser to access the Kubeslice Manager.", util.Tick, ep)
	}
}

func getNodeIP(cc *Cluster) (string, error) {
	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, true, "--context="+cc.ContextName, "--kubeconfig="+cc.KubeConfigPath, "get", "nodes", "-o", "jsonpath='{.items[*].status.addresses[?(@.type==\"InternalIP\")].address}'")
	if err == nil {
		nodeIPs := strings.FieldsFunc(outB.String(), func(c rune) bool { return c == ' ' || c == '\n' })
		if len(nodeIPs) > 0 {
			return nodeIPs[0], nil
		}
		return "", errors.New("No nodes found")
	}
	return "", err
}
