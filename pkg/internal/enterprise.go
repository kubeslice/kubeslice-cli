package internal

import (
	"fmt"
	"log"
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
