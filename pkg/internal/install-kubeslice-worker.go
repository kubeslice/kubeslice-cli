package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kubeslice/slicectl/util"
)

const workerValuesTemplate = `
controllerSecret:
  namespace: %s 
  endpoint: %s
  ca.crt: %s
  token: %s

cluster:
  name: %s
  nodeIp: %s
  endpoint: %s

`

func InstallKubeSliceWorker(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nInstalling KubeSlice Worker...")

	cc := ApplicationConfiguration.Configuration.ClusterConfiguration
	for _, cluster := range cc.WorkerClusters {
		filename := "helm-values-" + cluster.Name + ".yaml"
		generateWorkerValuesFile(cluster,
			filename,
			ApplicationConfiguration.Configuration.HelmChartConfiguration.ImagePullSecret,
			ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster,
			ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName)

		util.Printf("%s Generated Helm Values file for Worker Installation %s", util.Tick, filename)
		time.Sleep(200 * time.Millisecond)

		installWorker(cluster, filename, ApplicationConfiguration.Configuration.HelmChartConfiguration)
	}

	util.Printf("%s Successfully Installed Kubeslice Worker", util.Tick)
	time.Sleep(200 * time.Millisecond)
}

func generateWorkerValuesFile(cluster Cluster, valuesFile string, imagePullSecrets ImagePullSecrets, cc Cluster, projectName string) {
	secrets := fetchSecret(cluster.Name, cc, projectName)
	util.DumpFile(fmt.Sprintf(workerValuesTemplate+generateImagePullSecretsValue(imagePullSecrets), secrets["namespace"], secrets["controllerEndpoint"], secrets["ca.crt"], secrets["token"], cluster.Name, cluster.NodeIP, cluster.ControlPlaneAddress), kubesliceDirectory+"/"+valuesFile)
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
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, false, "--context="+cc.ContextName, "--kubeconfig="+cc.KubeConfigPath, "get", secret, "-n", "kubeslice-"+projectName, "-o", "jsonpath={.data}")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
	x := map[string]string{}
	err = json.Unmarshal([]byte(outB.String()), &x)
	if err != nil {
		log.Fatalf("failed to read secret %s", secret)
	}
	return x
}

func findSecret(workerName string, projectName string, cc Cluster) string {
	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, false, "--context="+cc.ContextName, "--kubeconfig="+cc.KubeConfigPath, "get", "secrets", "-n", "kubeslice-"+projectName, "-o", "name")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}

	var secret string
	for _, line := range strings.Split(outB.String(), "\n") {
		if strings.Contains(line, "worker-"+workerName) {
			secret = line
			break
		}
	}
	if secret == "" {
		log.Fatalf("failed to find secret for %s", workerName)
	}
	return secret
}
