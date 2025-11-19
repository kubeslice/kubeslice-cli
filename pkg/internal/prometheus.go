package internal

import (
	"fmt"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const (
	PrometheusValuesFileName = "helm-values-Prometheus.yaml"
	PrometheusNamespace      = "monitoring"
)

// Function variables for testing
var (
	podVerificationFuncPrometheus = PodVerification
)

func InstallPrometheus(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nInstalling Prometheus...")

	wc := ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters
	cc := ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster
	hc := ApplicationConfiguration.Configuration.HelmChartConfiguration
	generatePrometheusValuesFile(hc)
	util.Printf("%s Generated Helm Values file for Prometheus Installation %s", util.Tick, PrometheusValuesFileName)
	util.SystemClock.Sleep(200 * time.Millisecond)

	installPrometheus(wc, &cc, hc, PrometheusValuesFileName)
	util.Printf("%s Successfully installed Prometheus on Worker clusters.", util.Tick)
	util.SystemClock.Sleep(200 * time.Millisecond)

	util.Printf("%s Setting Prometheus endpoint in cluster objects...", util.Wait)
	projectNamespace := fmt.Sprintf("kubeslice-%s", ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName)
	patchClusterObjectInControllerCluster(wc, &cc, projectNamespace)
}

func patchClusterObjectInControllerCluster(wc []Cluster, cc *Cluster, projectNS string) {
	for _, cluster := range wc {
		err := util.CommandExecutor.Execute("kubectl", "--context", cc.ContextName, "--kubeconfig", cc.KubeConfigPath, "patch", ClusterObject, cluster.Name, "-n", projectNS, "--type", "merge", "-p", fmt.Sprintf("{\"spec\":{\"clusterProperty\":{\"telemetry\":{\"enabled\":true,\"endpoint\":\"http://%s:32700\",\"telemetryProvider\":\"prometheus\"}}}}", cluster.NodeIP))
		if err != nil {
			util.Fatalf("Process failed %v", err)
		}
		util.Printf("%s Successfully set prometheus endpoint in %s", util.Tick, cluster.Name)
	}
}

func generatePrometheusValuesFile(hcConfig HelmChartConfiguration) {
	err := generateValuesFileFunc(kubesliceDirectory+"/"+PrometheusValuesFileName, &hcConfig.PrometheusChart, "")
	if err != nil {
		util.Fatalf("%s %s", util.Cross, err)
	}
}

func installPrometheus(clusters []Cluster, cc *Cluster, hc HelmChartConfiguration, filename string) {
	for _, cluster := range clusters {
		args := make([]string, 0)
		args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "upgrade", "-i", hc.PrometheusChart.ChartName, fmt.Sprintf("%s/%s", hc.RepoAlias, hc.PrometheusChart.ChartName), "--namespace", PrometheusNamespace, "--create-namespace", "-f", kubesliceDirectory+"/"+filename)
		if hc.PrometheusChart.Version != "" {
			args = append(args, "--version", hc.PrometheusChart.Version)
		}
		err := util.CommandExecutor.Execute("helm", args...)
		if err != nil {
			util.Fatalf("Process failed %v", err)
		}
		util.Printf("%s Successfully installed helm chart %s/%s on cluster %s", util.Tick, hc.RepoAlias, hc.PrometheusChart.ChartName, cluster.Name)
		util.SystemClock.Sleep(200 * time.Millisecond)
		util.Printf("%s Waiting for Prometheus Pods to be Healthy...", util.Wait)
		podVerificationFuncPrometheus("Waiting for Prometheus Pods to be Healthy", cluster, PrometheusNamespace)
		// Patch cluster object in controller cluster
	}

}
