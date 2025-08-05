package internal

import (
	"fmt"
	"log"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const (
	controllerValuesFileName = "helm-values-controller.yaml"
)

const controllerValuesTemplate = `
kubeslice:
  controller:
    loglevel: info
    rbacResourcePrefix: kubeslice-rbac
    projectnsPrefix: kubeslice
    endpoint: %s
`

func InstallKubeSliceController(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nInstalling KubeSlice Controller...")

	cc := ApplicationConfiguration.Configuration.ClusterConfiguration
	hc := ApplicationConfiguration.Configuration.HelmChartConfiguration
	generateControllerValuesFile(cc.ControllerCluster, ApplicationConfiguration.Configuration.HelmChartConfiguration)
	util.Printf("%s Generated Helm Values file for Controller Installation %s", util.Tick, controllerValuesFileName)
	time.Sleep(200 * time.Millisecond)

	installKubeSliceController(cc.ControllerCluster, hc)
	util.Printf("%s Successfully installed helm chart %s/%s", util.Tick, hc.RepoAlias, hc.ControllerChart.ChartName)
	time.Sleep(2 * time.Second)

	util.Printf("%s Waiting for KubeSlice Controller Pods to be Healthy...", util.Wait)
	PodVerification("Waiting for KubeSlice Controller Pods to be Healthy", cc.ControllerCluster, KUBESLICE_CONTROLLER_NAMESPACE)

	if ApplicationConfiguration.Configuration.ClusterConfiguration.Profile != "" && ApplicationConfiguration.Configuration.ClusterConfiguration.Profile == ProfileEntDemo {
		util.Printf("%s Waiting for KubeSlice Trial License to be Ready...", util.Wait)
		LicenseVerification("Waiting for KubeSlice Trial License to be Ready", cc.ControllerCluster, KUBESLICE_CONTROLLER_NAMESPACE)
	}

	util.Printf("%s Successfully installed KubeSlice Controller.\n", util.Tick)

}

func UninstallKubeSliceController(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nUninstalling KubeSlice Controller...")
	cc := ApplicationConfiguration.Configuration.ClusterConfiguration
	time.Sleep(200 * time.Millisecond)
	uninstallKubeSliceController(cc.ControllerCluster)
	time.Sleep(200 * time.Millisecond)
	util.Printf("%s Successfully uninstalled KubeSlice Controller", util.Tick)
	// wait for pods to be cleaned up.
	// util.Printf("%s Waiting for KubeSlice Manager Pods to be removed...", util.Wait)
}

func generateControllerValuesFile(cluster Cluster, hcConfig HelmChartConfiguration) {
	err := GenerateValuesFile(kubesliceDirectory+"/"+controllerValuesFileName, &hcConfig.ControllerChart, fmt.Sprintf(controllerValuesTemplate+generateImagePullSecretsValue(hcConfig.ImagePullSecret), cluster.ControlPlaneAddress))
	if err != nil {
		log.Fatalf("%s %s", util.Cross, err)
	}
}

func installKubeSliceController(cluster Cluster, hc HelmChartConfiguration) {
	args := make([]string, 0)
	args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "upgrade", "-i", KUBESLICE_CONTROLLER_NAMESPACE, fmt.Sprintf("%s/%s", hc.RepoAlias, hc.ControllerChart.ChartName), "--namespace", KUBESLICE_CONTROLLER_NAMESPACE, "--create-namespace", "-f", kubesliceDirectory+"/"+controllerValuesFileName)
	if hc.ControllerChart.Version != "" {
		args = append(args, "--version", hc.ControllerChart.Version)
	}
	err := util.RunCommand("helm", args...)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func uninstallKubeSliceController(cluster Cluster) {
	args := make([]string, 0)
	args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "uninstall", KUBESLICE_CONTROLLER_NAMESPACE, "--namespace", KUBESLICE_CONTROLLER_NAMESPACE)
	err := util.RunCommand("helm", args...)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}
