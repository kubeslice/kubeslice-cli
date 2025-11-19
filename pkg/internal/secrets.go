package internal

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

var (
	getSecretNameFunc              = GetSecretName
	getKubectlResourcesFuncSecrets = GetKubectlResources
)

func GetSecrets(workerName string, namespace string, controllerCluster *Cluster, outputFormat string) {
	if controllerCluster == nil {
		util.Printf("%s Controller cluster cannot be nil", util.Cross)
		return
	}

	util.Printf("\nFetching KubeSlice secret...")
	SecretName := getSecretNameFunc(workerName, namespace, controllerCluster)
	if SecretName == "" {
		util.Printf("%s No secret found for worker %s", util.Cross, workerName)
		return
	}
	getKubectlResourcesFuncSecrets(SecretObject, SecretName, namespace, controllerCluster, outputFormat)
	util.SystemClock.Sleep(200 * time.Millisecond)
}

func GetSecretName(workerName string, namespace string, controllerCluster *Cluster) string {
	if controllerCluster == nil {
		util.Printf("%s Controller cluster cannot be nil", util.Cross)
		return ""
	}

	jsonpath := fmt.Sprintf(`{.items[?(@.metadata.name contains "worker-%s")].metadata.name}`, workerName)
	var outB, errB bytes.Buffer
	err := util.CommandExecutor.ExecuteWithOutput(
		"kubectl",
		&outB,
		&errB,
		"--context="+controllerCluster.ContextName,
		"--kubeconfig="+controllerCluster.KubeConfigPath,
		"get", SecretObject,
		"-n", namespace,
		"-o", "jsonpath="+jsonpath,
	)

	if err != nil {
		util.Printf("%s Failed to get secret: %v", util.Cross, err)
		return ""
	}

	secretName := strings.TrimSpace(outB.String())
	if secretName == "" {
		util.Printf("%s No matching secret found for worker-%s", util.Cross, workerName)
	}
	return secretName
}
