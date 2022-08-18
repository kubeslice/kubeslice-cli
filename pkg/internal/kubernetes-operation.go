package internal

import (
	"bytes"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/kubeslice/slicectl/util"
)

type PodVerificationStatus int

const (
	PodVerificationStatusSuccess PodVerificationStatus = iota
	PodVerificationStatusInProgress
	PodVerificationStatusFailed
)

func PodVerification(message string, cluster Cluster, namespace string) {
	var i = 0
	var backoffCount = 0
	var backoffLimit = 6
	for {
		i = i + 1
		time.Sleep(5 * time.Second)
		status, output := verifyPods(cluster, namespace)
		if status == PodVerificationStatusSuccess {
			break
		} else if status == PodVerificationStatusFailed {
			backoffCount = backoffCount + 1
			if backoffCount > backoffLimit {
				log.Fatalf("Pod(s) in error state\n%s", output)
			}
		}
		util.Printf("%s %s... %d seconds elapsed", util.Wait, message, i*5)
	}
}

func ApplyKubectlManifest(fileName, namespace string, cluster Cluster) {
	err := util.RunCommand("kubectl", "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "apply", "-f", fileName, "-n", namespace)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func GetKubectlResources(resourceType string, resourceName string, namespace string, cluster *Cluster) {
	cmdArgs := []string{}
	if cluster != nil {
		cmdArgs = append(cmdArgs, "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath)
	}
	if resourceName == "" {
		cmdArgs = append(cmdArgs, "get", resourceType, "-n", namespace)
	} else {
		cmdArgs = append(cmdArgs, "get", resourceType, resourceName, "-n", namespace)
	}
	err := util.RunCommandOnStdIO("kubectl", cmdArgs...)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func DeleteKubectlResources(resourceType string, resourceName string, namespace string, cluster *Cluster) {
	cmdArgs := []string{}
	if cluster != nil {
		cmdArgs = append(cmdArgs, "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath)
	}
	cmdArgs = append(cmdArgs, "delete", resourceType, resourceName, "-n", namespace)
	err := util.RunCommandOnStdIO("kubectl", cmdArgs...)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func verifyPods(cluster Cluster, namespace string) (PodVerificationStatus, string) {
	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, true, "--context="+cluster.ContextName, "--kubeconfig="+cluster.KubeConfigPath, "get", "pods", "-n", namespace)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
	var count = 0
	var lines = 0
	for _, line := range strings.Split(outB.String(), "\n") {
		if strings.Contains(line, "Error") || strings.Contains(line, "ImagePullBackOff") || strings.Contains(line, "CrashLoopBackOff") {
			return PodVerificationStatusFailed, outB.String()
		}
		if strings.Contains(line, "Completed") {
			continue
		}
		if strings.Contains(line, "/") {
			index := strings.Index(line, "/")
			upper, _ := strconv.Atoi(string(line[index+1]))
			lower, _ := strconv.Atoi(string(line[index-1]))
			if upper == lower {
				count = count + 1
			}
			lines = lines + 1
		}
	}
	if count == lines {
		return PodVerificationStatusSuccess, outB.String()
	}
	return PodVerificationStatusInProgress, outB.String()
}
