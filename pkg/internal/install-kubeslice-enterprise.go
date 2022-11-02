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
	util.Printf("\nInstalling KubeSlice dashboard...")
	if ApplicationConfiguration.Configuration.HelmChartConfiguration.UIChart.ChartName == "" {
		util.Printf("%s UI Helm Chart not found. Update UI chart configuration in topology file.", util.Cross)
		return
	}
	cc := ApplicationConfiguration.Configuration.ClusterConfiguration
	hc := ApplicationConfiguration.Configuration.HelmChartConfiguration
	time.Sleep(200 * time.Millisecond)
	clusterType := ApplicationConfiguration.Configuration.ClusterConfiguration.ClusterType
	generateUIValuesFile(clusterType, cc.ControllerCluster, ApplicationConfiguration.Configuration.HelmChartConfiguration.ImagePullSecret)
	installKubeSliceUI(cc.ControllerCluster, hc)
	util.Printf("%s Successfully installed helm chart %s/%s", util.Tick, hc.RepoAlias, hc.UIChart.ChartName)
	time.Sleep(200 * time.Millisecond)
	util.Printf("%s Waiting for KubeSlice UI Pods to be Healthy...", util.Wait)
	PodVerification("Waiting for KubeSlice Ui Pods to be Healthy", cc.ControllerCluster, "kubernetes-dashboard")
	util.Printf("%s Successfully installed KubeSlice dashboard.\n", util.Tick)
}

func generateUIValuesFile(clusterType string, cluster Cluster, imagePullSecrets ImagePullSecrets) {
	serviceType := ""
	if clusterType == "kind" {
		serviceType = "NodePort"
	} else {
		serviceType = "LoadBalancer"
	}
	util.DumpFile(fmt.Sprintf(UIValuesTemplate+generateImagePullSecretsValue(imagePullSecrets), serviceType), kubesliceDirectory+"/"+uiValuesFileName)
}

func installKubeSliceUI(cluster Cluster, hc HelmChartConfiguration) {
	args := make([]string, 0)
	args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "upgrade", "-i", "kubeslice-ui", fmt.Sprintf("%s/%s", hc.RepoAlias, hc.UIChart.ChartName), "--namespace", "kubeslice-controller", "-f", kubesliceDirectory+"/"+uiValuesFileName)
	if hc.ControllerChart.Version != "" {
		args = append(args, "--version", hc.ControllerChart.Version)
	}
	err := util.RunCommand("helm", args...)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}
