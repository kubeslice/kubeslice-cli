package internal

import (
	"fmt"
	"log"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const (
	PrometheusValuesFileName = "helm-values-Prometheus.yaml"
	PrometheusNamespace      = "monitoring"
)

func InstallPrometheus(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nInstalling Prometheus...")

	wc := ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters
	cc := ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster
	hc := ApplicationConfiguration.Configuration.HelmChartConfiguration
	generatePrometheusValuesFile(hc)
	util.Printf("%s Generated Helm Values file for Prometheus Installation %s", util.Tick, PrometheusValuesFileName)
	time.Sleep(200 * time.Millisecond)
	installPrometheus(wc, &cc, hc, PrometheusValuesFileName)
	util.Printf("%s Successfully installed Prometheus on Worker clusters.", util.Tick)
	time.Sleep(200 * time.Millisecond)
	util.Printf("%s Setting Prometheus endpoint in cluster objects...", util.Wait)
	projectNamespce := fmt.Sprintf("kubeslice-%s", ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName)
	patchClusterObjectInControllerCluster(wc, &cc, projectNamespce)
}

func patchClusterObjectInControllerCluster(wc []Cluster, cc *Cluster, projectNS string) {
	for _, cluster := range wc {
		// Patch cluster object in controller cluster
		err := util.RunCommand("kubectl", "--context", cc.ContextName, "--kubeconfig", cc.KubeConfigPath, "patch", ClusterObject, cluster.Name, "-n", projectNS, "--type", "merge", "-p", fmt.Sprintf("{\"spec\":{\"clusterProperty\":{\"telemetry\":{\"enabled\":true,\"endpoint\":\"http://%s:32700\",\"telemetryProvider\":\"prometheus\"}}}}", cluster.NodeIP))
		if err != nil {
			log.Fatalf("Process failed %v", err)
		}
		util.Printf("%s Successfully set prometheus endpoint in %s", util.Tick, cluster.Name)
	}
}

func generatePrometheusValuesFile(hcConfig HelmChartConfiguration) {
	err := generateValuesFile(kubesliceDirectory+"/"+PrometheusValuesFileName, &hcConfig.PrometheusChart, "")
	if err != nil {
		log.Fatalf("%s %s", util.Cross, err)
	}
}

func installPrometheus(clusters []Cluster, cc *Cluster, hc HelmChartConfiguration, filename string) {
	for _, cluster := range clusters {
		args := make([]string, 0)
		args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "upgrade", "-i", hc.PrometheusChart.ChartName, fmt.Sprintf("%s/%s", hc.RepoAlias, hc.PrometheusChart.ChartName), "--namespace", PrometheusNamespace, "--create-namespace", "-f", kubesliceDirectory+"/"+filename)
		if hc.ControllerChart.Version != "" {
			args = append(args, "--version", hc.PrometheusChart.Version)
		}
		err := util.RunCommand("helm", args...)
		if err != nil {
			log.Fatalf("Process failed %v", err)
		}
		util.Printf("%s Successfully installed helm chart %s/%s on cluster %s", util.Tick, hc.RepoAlias, hc.PrometheusChart.ChartName, cluster.Name)
		time.Sleep(200 * time.Millisecond)
		util.Printf("%s Waiting for Prometheus Pods to be Healthy...", util.Wait)
		PodVerification("Waiting for Prometheus Pods to be Healthy", cluster, PrometheusNamespace)
		// Patch cluster object in controller cluster
	}

}
