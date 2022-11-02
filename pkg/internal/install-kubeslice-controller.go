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
	generateControllerValuesFile(cc.ControllerCluster, ApplicationConfiguration.Configuration.HelmChartConfiguration.ImagePullSecret)
	util.Printf("%s Generated Helm Values file for Controller Installation %s", util.Tick, controllerValuesFileName)
	time.Sleep(200 * time.Millisecond)

	installKubeSliceController(cc.ControllerCluster, hc)
	util.Printf("%s Successfully installed helm chart %s/%s", util.Tick, hc.RepoAlias, hc.ControllerChart.ChartName)
	time.Sleep(200 * time.Millisecond)

	util.Printf("%s Waiting for KubeSlice Controller Pods to be Healthy...", util.Wait)
	PodVerification("Waiting for KubeSlice Controller Pods to be Healthy", cc.ControllerCluster, "kubeslice-controller")

	util.Printf("%s Successfully installed KubeSlice Controller.\n", util.Tick)

}

func generateControllerValuesFile(cluster Cluster, imagePullSecret ImagePullSecrets) {

	util.DumpFile(fmt.Sprintf(controllerValuesTemplate+generateImagePullSecretsValue(imagePullSecret), cluster.ControlPlaneAddress), kubesliceDirectory+"/"+controllerValuesFileName)
}

func installKubeSliceController(cluster Cluster, hc HelmChartConfiguration) {
	args := make([]string, 0)
	args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "upgrade", "-i", "kubeslice-controller", fmt.Sprintf("%s/%s", hc.RepoAlias, hc.ControllerChart.ChartName), "--namespace", "kubeslice-controller", "--create-namespace", "-f", kubesliceDirectory+"/"+controllerValuesFileName)
	if hc.ControllerChart.Version != "" {
		args = append(args, "--version", hc.ControllerChart.Version)
	}
	err := util.RunCommand("helm", args...)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}
