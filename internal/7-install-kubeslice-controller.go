package internal

import (
	"fmt"
	"log"
	"time"

	"github.com/kubeslice/kubeslice-installer/util"
)

const (
	controllerValuesFileName = "helm-values-controller.yaml"
)

const controllerValuesTemplate = `
kubeslice:
  controller:
    loglevel: info
    rbacResourcePrefix: kubeslice-rbac
    projectnsPrefix: kubeslice
    endpoint: https://%s:6443
`

func InstallKubeSliceController() {
	util.Printf("\nInstalling KubeSlice Controller...")

	generateControllerValuesFile()
	util.Printf("%s Generated Helm Values file for Controller Installation %s", util.Tick, controllerValuesFileName)
	time.Sleep(200 * time.Millisecond)

	installKubeSliceController()
	util.Printf("%s Successfully installed helm chart %s/%s", util.Tick, helmRepoAlias, controllerChartName)
	time.Sleep(200 * time.Millisecond)

	util.Printf("%s Waiting for KubeSlice Controller Pods to be Healthy...", util.Wait)
	util.PodVerification("Waiting for KubeSlice Controller Pods to be Healthy", controllerName, "kubeslice-controller")

	util.Printf("%s Successfully installed KubeSlice Controller.\n", util.Tick)

}

func generateControllerValuesFile() {
	util.DumpFile(fmt.Sprintf(controllerValuesTemplate, dockerNetworkMap[controllerName]), kubesliceDirectory+"/"+controllerValuesFileName)
}

func installKubeSliceController() {
	err := util.RunCommand("helm", "--kube-context", "kind-"+controllerName, "upgrade", "-i", "kubeslice-controller", fmt.Sprintf("%s/%s", helmRepoAlias, controllerChartName), "--namespace", "kubeslice-controller", "--create-namespace", "-f", kubesliceDirectory+"/"+controllerValuesFileName)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}
