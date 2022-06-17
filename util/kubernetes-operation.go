package util

import (
	"bytes"
	"log"
	"strconv"
	"strings"
	"time"
)

type PodVerificationStatus int
const (
	PodVerificationStatusSuccess PodVerificationStatus = iota
	PodVerificationStatusInProgress
	PodVerificationStatusFailed
)

func PodVerification(message, clusterName, namespace string) {
	var i = 0
	var backoffCount = 0
	var backoffLimit = 6
	for {
		i = i + 1
		time.Sleep(5 * time.Second)
		status, output := verifyPods(clusterName, namespace)
		if status == PodVerificationStatusSuccess {
			break
		} else if status == PodVerificationStatusFailed {
			backoffCount = backoffCount + 1
			if backoffCount > backoffLimit {
				log.Fatalf("Pod(s) in error state\n%s", output)
			}
		}
		Printf("%s %s... %d seconds elapsed", Wait, message, i*5)
	}
}

func ApplyKubectlManifest(fileName, namespace, clusterName string) {
	err := RunCommand("kubectl", "--context=kind-"+clusterName, "apply", "-f", fileName, "-n", namespace)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func verifyPods(clusterName, namespace string) (PodVerificationStatus, string) {
	var outB, errB bytes.Buffer
	err := RunCommandCustomIO("kubectl", &outB, &errB, true, "--context=kind-"+clusterName, "get", "pods", "-n", namespace)
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
