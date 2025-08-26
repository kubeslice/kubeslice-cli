package internal

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

func GetSecrets(workerName string, namespace string, controllerCluster *Cluster, outputFormat string) {
	util.Printf("\nFetching KubeSlice secret...")
	SecretName := GetSecretName(workerName, namespace, controllerCluster)
	GetKubectlResources(SecretObject, SecretName, namespace, controllerCluster, outputFormat)
	time.Sleep(200 * time.Millisecond)
}

func GetSecretName(workerName string, namespace string, controllerCluster *Cluster) string {
	cmdArgs := []string{}
	cmdArgs = append(cmdArgs, "get", SecretObject, "-n", namespace)
	var outB bytes.Buffer
	kubectlPath := os.Getenv("KUBECTL_PATH")
	if kubectlPath == "" {
		kubectlPath = "/home/excellarate/.local/bin/kubectl"
	}
	c1 := exec.Command(kubectlPath, cmdArgs...)
	c2 := exec.Command("grep", "worker-"+workerName)
	c3 := exec.Command("awk", "{print $1}")
	c2.Stdin, _ = c1.StdoutPipe()
	c3.Stdin, _ = c2.StdoutPipe()
	c3.Stdout = &outB
	_ = c2.Start()
	_ = c3.Start()
	_ = c1.Run()
	_ = c2.Wait()
	_ = c3.Wait()
	s := outB.String()
	return strings.TrimSuffix(s, "\n")
}
