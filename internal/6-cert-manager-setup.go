package internal

import (
	"fmt"
	"log"
	"time"

	"github.com/kubeslice/slicectl/util"
)

func InstallCertManager() {
	util.Printf("\nInstall Cert Manager to Controller Cluster...")

	installCertManager()
	util.Printf("%s Successfully installed helm chart %s/%s", util.Tick, helmRepoAlias, certManagerChartName)
	time.Sleep(200 * time.Millisecond)

	util.Printf("%s Waiting for Cert Manager Pods to be Healthy...", util.Wait)
	util.PodVerification("Waiting for Cert Manager Pods to be Healthy", controllerName, "cert-manager")

	util.Printf("%s Successfully installed cert manager.\n", util.Tick)

}

func installCertManager() {
	err := util.RunCommand("helm", "--kube-context", "kind-"+controllerName, "upgrade", "-i", "cert-manager", fmt.Sprintf("%s/%s", helmRepoAlias, certManagerChartName), "--namespace", "cert-manager", "--create-namespace", "--set", "installCRDs=true")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}
