package internal

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kubeslice/slicectl/util"
)

const kubeconfigPath = kubesliceDirectory + "/kubeconfig.yaml"

func CreateKindClusters() {

	controllerExist, worker1Exist, worker2Exist := getExistingClusters()

	if controllerExist && worker1Exist && worker2Exist {
		util.Printf("\nKind clusters already exist... Skipping\n")
		return
	}
	util.Printf("\nCreating Kind Clusters...")

	if !controllerExist {
		createKindCluster(controllerFilename)
		util.Printf("%s Created Kind Cluster : %s", util.Tick, controllerName)
		time.Sleep(200 * time.Millisecond)
	}

	if !worker1Exist {
		createKindCluster(worker1Filename)
		util.Printf("%s Created Kind Cluster : %s", util.Tick, worker1Name)
		time.Sleep(200 * time.Millisecond)
	}

	if !worker2Exist {
		createKindCluster(worker2Filename)
		util.Printf("%s Created Kind Cluster : %s", util.Tick, worker2Name)
		time.Sleep(200 * time.Millisecond)
	}

	util.Printf("Created required kind clusters")
}

func SetKubeConfigPath() {
	os.Setenv("KUBECONFIG", kubeconfigPath)
}

func CreateKubeConfig() {
	if _, err := os.Stat(kubeconfigPath); errors.Is(err, os.ErrNotExist) {
		util.DumpFile("", kubeconfigPath)
		util.Printf("%s Created Empty KubeConfig file : %s", util.Tick, kubeconfigPath)
		time.Sleep(200 * time.Millisecond)
	}
}

func getExistingClusters() (bool, bool, bool) {
	controllerExist := false
	worker1Exist := false
	worker2Exist := false

	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("kind", &outB, &errB, true, "get", "clusters")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
	for _, line := range strings.Split(outB.String(), "\n") {
		if strings.Contains(line, controllerName) {
			controllerExist = true
		}
		if strings.Contains(line, worker1Name) {
			worker1Exist = true
		}
		if strings.Contains(line, worker2Name) {
			worker2Exist = true
		}
	}
	return controllerExist, worker1Exist, worker2Exist
}

func createKindCluster(configFile string) {
	err := util.RunCommandOnStdIO("kind", "create", "cluster", fmt.Sprintf("--config=%s/%s/%s", kubesliceDirectory, kindSubDirectory, configFile))
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func DeleteKindClusters() {
	controllerExist, worker1Exist, worker2Exist := getExistingClusters()
	args := make([]string, 0, 0)
	args = append(args, "delete", "clusters")
	if !controllerExist && !worker1Exist && !worker2Exist {
		util.Printf("No Kind Clusters found for deletion")
		return
	}
	if controllerExist {
		args = append(args, controllerName)
	}
	if worker1Exist {
		args = append(args, worker1Name)
	}
	if worker2Exist {
		args = append(args, worker2Name)
	}
	err := util.RunCommand("kind", args...)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}
