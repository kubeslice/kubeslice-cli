package internal

import (
	"fmt"
	"log"
	"time"

	"github.com/kubeslice/slicectl/util"
)

func InstallKubeSliceUI(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nInstalling KubeSlice dashboard...")
	cc := ApplicationConfiguration.Configuration.ClusterConfiguration
	hc := ApplicationConfiguration.Configuration.HelmChartConfiguration
	time.Sleep(200 * time.Millisecond)

	installKubeSliceUI(cc.ControllerCluster, hc)
	util.Printf("%s Successfully installed helm chart %s/%s", util.Tick, hc.RepoAlias, hc.UIChart.ChartName)
	time.Sleep(200 * time.Millisecond)

	util.Printf("%s Waiting for KubeSlice UI Pods to be Healthy...", util.Wait)
	PodVerification("Waiting for KubeSlice Ui Pods to be Healthy", cc.ControllerCluster, "kubernetes-dashboard")

	util.Printf("%s Successfully installed KubeSlice dashboard.\n", util.Tick)

}

func installKubeSliceUI(cluster Cluster, hc HelmChartConfiguration) {
	args := make([]string, 0)
	fmt.Println("roshani", hc.UIChart.ChartName)
	args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "upgrade", "-i", "kubeslice-ui", fmt.Sprintf("%s/%s", hc.RepoAlias, hc.UIChart.ChartName), "--namespace", "kubeslice-controller")
	if hc.ControllerChart.Version != "" {
		args = append(args, "--version", hc.ControllerChart.Version)
	}
	err := util.RunCommand("helm", args...)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}
