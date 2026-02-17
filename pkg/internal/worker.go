package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const workerValuesTemplate = `
controllerSecret:
  namespace: %s 
  endpoint: %s
  ca.crt: %s
  token: %s
metrics:
  insecure: %t

cluster:
  name: %s
  endpoint: %s

`

func InstallKubeSliceWorker(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nInstalling KubeSlice Worker...")

	cc := ApplicationConfiguration.Configuration.ClusterConfiguration
	for _, cluster := range cc.WorkerClusters {
		filename := "helm-values-" + cluster.Name + ".yaml"
		insecureMetrics := ApplicationConfiguration.Configuration.ClusterConfiguration.ClusterType == Kind_Component
		generateWorkerValuesFile(cluster,
			filename,
			ApplicationConfiguration.Configuration,
			insecureMetrics,
		)

		util.Printf("%s Generated Helm Values file for Worker Installation %s", util.Tick, filename)
		time.Sleep(200 * time.Millisecond)

		installWorker(cluster, filename, ApplicationConfiguration.Configuration.HelmChartConfiguration)
	}

	util.Printf("%s Successfully Installed Kubeslice Worker", util.Tick)
	time.Sleep(200 * time.Millisecond)
}

func UninstallKubeSliceWorker(ApplicationConfiguration *ConfigurationSpecs, workersToUninstall map[string]string) {
	util.Printf("\nUninstalling KubeSlice Worker...")

	_, uninstallAllWorker := workersToUninstall["*"]

	cc := ApplicationConfiguration.Configuration.ClusterConfiguration
	for _, cluster := range cc.WorkerClusters {
		_, found := workersToUninstall[cluster.Name]
		if found || uninstallAllWorker {
			uninstallKubeSliceWorkerHelm(cluster)
			time.Sleep(200 * time.Millisecond)
		}
	}

	// util.Printf("%s Successfully Installed Kubeslice Worker", util.Tick)
	time.Sleep(200 * time.Millisecond)
}

// Retry tries to execute the funtion, If failed reattempts till backoffLimit
func Retry(backoffLimit int, sleep time.Duration, f func() error) (err error) {
	start := time.Now()
	for i := 0; i < backoffLimit; i++ {
		if i > 0 {
			time.Sleep(sleep)
			sleep *= 2
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	elapsed := time.Since(start)
	return fmt.Errorf("retry failed after %d attempts (took %d seconds), last error: %s", backoffLimit, int(elapsed.Seconds()), err)
}

// CreateWorkerSpecificHelmChart merges global worker chart values with cluster-specific values
func CreateWorkerSpecificHelmChart(globalChart HelmChart, cluster Cluster) HelmChart {
	// Start with a copy of the global chart
	workerChart := HelmChart{
		ChartName: globalChart.ChartName,
		Version:   globalChart.Version,
		Values:    make(map[string]interface{}),
	}

	// Copy global values first
	for k, v := range globalChart.Values {
		workerChart.Values[k] = v
	}

	// Override with cluster-specific values if they exist
	if cluster.HelmValues != nil {
		for k, v := range cluster.HelmValues {
			workerChart.Values[k] = v
		}
	}

	return workerChart
}

func generateWorkerValuesFile(cluster Cluster, valuesFile string, config Configuration, insecureMetrics bool) {
	var secrets map[string]string
	err := Retry(3, 1*time.Second, func() (err error) {
		secrets = fetchSecret(cluster.Name, config.ClusterConfiguration.ControllerCluster, config.KubeSliceConfiguration.ProjectName)
		if secrets["namespace"] == "" || secrets["controllerEndpoint"] == "" || secrets["ca.crt"] == "" || secrets["token"] == "" {
			return fmt.Errorf("secret is empty")
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Unable to fetch secrets\n%s", err)
	}

	// Create a worker-specific helm chart configuration
	workerChart := CreateWorkerSpecificHelmChart(config.HelmChartConfiguration.WorkerChart, cluster)
	
	err = generateValuesFile(kubesliceDirectory+"/"+valuesFile, &workerChart, fmt.Sprintf(workerValuesTemplate+generateImagePullSecretsValue(config.HelmChartConfiguration.ImagePullSecret), secrets["namespace"], secrets["controllerEndpoint"], secrets["ca.crt"], secrets["token"], insecureMetrics, cluster.Name, cluster.ControlPlaneAddress))
	if err != nil {
		log.Fatalf("%s %s", util.Cross, err)
	}
}

func installWorker(cluster Cluster, valuesName string, helmChartConfig HelmChartConfiguration) {
	hc := helmChartConfig
	installKubeSliceWorkerHelm(cluster, valuesName, hc)
	util.Printf("%s Successfully installed helm chart %s/%s on %s", util.Tick, hc.RepoAlias, hc.WorkerChart.ChartName, cluster.Name)
	time.Sleep(200 * time.Millisecond)

	util.Printf("%s Waiting for KubeSlice Worker Pods to be Healthy...", util.Wait)
	PodVerification("Waiting for KubeSlice Worker Pods to be Healthy", cluster, "kubeslice-system")

	util.Printf("%s Successfully installed KubeSlice Worker %s.", util.Tick, cluster.Name)
}

func installKubeSliceWorkerHelm(cluster Cluster, valuesFile string, hc HelmChartConfiguration) {
	args := make([]string, 0)
	args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "upgrade", "-i", "kubeslice-worker", fmt.Sprintf("%s/%s", hc.RepoAlias, hc.WorkerChart.ChartName), "--namespace", "kubeslice-system", "--create-namespace", "-f", kubesliceDirectory+"/"+valuesFile)
	if hc.WorkerChart.Version != "" {
		args = append(args, "--version", hc.WorkerChart.Version)
	}
	err := util.RunCommand("helm", args...)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func fetchSecret(clusterName string, cc Cluster, projectName string) map[string]string {
	//kubectl get secrets -n kubeslice-demo -o name
	secret := findSecret(clusterName, projectName, cc)
	//kubectl get secret/kubeslice-rbac-worker-kubeslice-worker-1-token-h99pc -n kubeslice-demo -o jsonpath={.data}
	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, true, "--context="+cc.ContextName, "--kubeconfig="+cc.KubeConfigPath, "get", secret, "-n", "kubeslice-"+projectName, "-o", "jsonpath={.data}")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
	x := map[string]string{}
	err = json.Unmarshal(outB.Bytes(), &x)
	if err != nil {
		log.Fatalf("failed to read secret %s", secret)
	}
	return x
}

func findSecret(workerName string, projectName string, cc Cluster) string {
	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, true, "--context="+cc.ContextName, "--kubeconfig="+cc.KubeConfigPath, "get", "sa", "-n", "kubeslice-"+projectName, "-o", "name")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}

	var secret string
	for _, line := range strings.Split(outB.String(), "\n") {
		if strings.Contains(line, "rbac-worker-"+workerName) {
			secret = fmt.Sprintf("secrets/%s", strings.TrimPrefix(line, "serviceaccount/"))
			break
		}
	}
	if secret == "" {
		log.Fatalf("failed to find secret for %s", workerName)
	}
	return secret
}

func uninstallKubeSliceWorkerHelm(cluster Cluster) {
	args := make([]string, 0)
	args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "uninstall", "kubeslice-worker", "--namespace", "kubeslice-system")

	err := util.RunCommand("helm", args...)
	if err != nil {
		util.Printf("%s Uninstall failed. %v", util.Cross, err)
	}
	util.Printf("%s Successfully uninstalled KubeSlice Worker %s.", util.Tick, cluster.Name)
}
