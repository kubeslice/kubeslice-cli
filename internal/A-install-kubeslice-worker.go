package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-installer/util"
)

const (
	worker1ValuesFileName = "helm-values-worker-1.yaml"
	worker2ValuesFileName = "helm-values-worker-2.yaml"
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

`

func InstallKubeSliceWorker() {
	util.Printf("\nInstalling KubeSlice Worker...")

	generateWorkerValuesFile(worker1Name, worker1ValuesFileName)
	util.Printf("%s Generated Helm Values file for Worker Installation %s", util.Tick, worker1ValuesFileName)
	time.Sleep(200 * time.Millisecond)

	installWorker(worker1Name, worker1ValuesFileName)

	generateWorkerValuesFile(worker2Name, worker2ValuesFileName)
	util.Printf("%s Generated Helm Values file for Worker Installation %s", util.Tick, worker1ValuesFileName)
	time.Sleep(200 * time.Millisecond)

	installWorker(worker2Name, worker2ValuesFileName)

	util.Printf("%s Successfully Installed Kubeslice Worker", util.Tick)
	time.Sleep(200 * time.Millisecond)
}

func generateWorkerValuesFile(clusterName, valuesFile string) {
	secrets := fetchSecret(clusterName)
	util.DumpFile(fmt.Sprintf(workerValuesTemplate, secrets["namespace"], secrets["controllerEndpoint"], secrets["ca.crt"], secrets["token"], clusterName, dockerNetworkMap[clusterName]), kubesliceDirectory+"/"+valuesFile)
}

func installWorker(clusterName, valuesName string) {
	installKubeSliceWorkerHelm(clusterName, valuesName)
	util.Printf("%s Successfully installed helm chart %s/%s on %s", util.Tick, helmRepoAlias, workerChartName, clusterName)
	time.Sleep(200 * time.Millisecond)

	util.Printf("%s Waiting for KubeSlice Worker Pods to be Healthy...", util.Wait)
	util.PodVerification("Waiting for KubeSlice Worker Pods to be Healthy", clusterName, "kubeslice-system")

	util.Printf("%s Successfully installed KubeSlice Worker %s.", util.Tick, clusterName)
}

func installKubeSliceWorkerHelm(clusterName, valuesFile string) {
	err := util.RunCommand("helm", "--kube-context", "kind-"+clusterName, "upgrade", "-i", "kubeslice-worker", fmt.Sprintf("%s/%s", helmRepoAlias, workerChartName), "--namespace", "kubeslice-system", "--create-namespace", "-f", kubesliceDirectory+"/"+valuesFile)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func fetchSecret(clusterName string) map[string]string {
	//kubectl get secrets -n kubeslice-demo -o name
	secret := findSecret(clusterName)
	//kubectl get secret/kubeslice-rbac-worker-kubeslice-worker-1-token-h99pc -n kubeslice-demo -o jsonpath={.data}
	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, false, "--context=kind-"+controllerName, "get", secret, "-n", "kubeslice-demo", "-o", "jsonpath={.data}")
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

func findSecret(clusterName string) string {
	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, false, "--context=kind-"+controllerName, "get", "secrets", "-n", "kubeslice-demo", "-o", "name")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}

	var secret string
	for _, line := range strings.Split(outB.String(), "\n") {
		if strings.Contains(line, "worker-"+clusterName) {
			secret = line
			break
		}
	}
	if secret == "" {
		log.Fatalf("failed to find secret for %s", clusterName)
	}
	return secret
}
