package internal

import (
	"log"
	"time"

	"github.com/kubeslice/kubeslice-installer/util"
)

const (
	helmRepo = "https://kubeslice.github.io/charts/"
	helmRepoAlias = "kubeslice-demo"
	certManagerChartName = "cert-manager"
	controllerChartName = "kubeslice-controller"
	workerChartName = "kubeslice-worker"
)

func AddHelmCharts() {
	// helm repo add avesha https://kubeslice.github.io/charts/
	util.Printf("\nAdding KubeSlice Helm Charts...")

	addHelmChart()
	util.Printf("%s Successfully added helm repo %s : %s", util.Tick, helmRepoAlias, helmRepo)
	time.Sleep(200 * time.Millisecond)

	updateHelmChart()
	util.Printf("%s Successfully updated helm repo", util.Tick)
	time.Sleep(200 * time.Millisecond)

	//util.Printf("%s Listing helm repo for charts: ", util.Tick)
	//listHelmChart()
	//time.Sleep(200 * time.Millisecond)

	util.Printf("%s Successfully added helm charts.\n", util.Tick)
}

func addHelmChart() {
	err := util.RunCommand("helm", "repo", "add", helmRepoAlias, helmRepo, "--force-update")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func updateHelmChart() {
	err := util.RunCommand("helm", "repo", "update")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func listHelmChart() {
	err := util.RunCommandOnStdIO("helm", "search", "repo", helmRepoAlias)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}