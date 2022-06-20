package main

import (
	"os"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-installer/internal"
	"github.com/kubeslice/kubeslice-installer/util"
)

func main() {
	args := os.Args
	if len(args) != 2 {
		util.Printf("Invalid arguments. Try running %s --help", args[0])
		return
	}
	switch strings.Trim(args[1], "-") {
	case "full-install":
		seamlessInstall()
	case "minimal-install":
		minimalInstall()
	case "uninstall":
		uninstall()
	case "cleanup":
		cleanup()
	case "help":
		printHelp()
	}
}

func seamlessInstall() {
	basicInstall()
	internal.GenerateSliceConfiguration()
	internal.ApplySliceConfiguration()
	util.Printf("%s Waiting for configuration propagation", util.Wait)
	time.Sleep(20 * time.Second)
	internal.GenerateIPerfManifests()
	internal.GenerateIPerfServiceExportManifest()
	internal.InstallIPerf()
	internal.ApplyIPerfServiceExportManifest()
	util.Printf("%s Waiting for configuration propagation", util.Wait)
	time.Sleep(20 * time.Second)
	internal.RolloutRestartIPerf()
	internal.PrintNextSteps(true)
}

func minimalInstall() {
	basicInstall()
	internal.GenerateSliceConfiguration()
	internal.GenerateIPerfManifests()
	internal.InstallIPerf()
	internal.GenerateIPerfServiceExportManifest()
	internal.PrintNextSteps(false)
}

func uninstall() {
	internal.VerifyExecutables()
	internal.SetKubeConfigPath()
	internal.DeleteKindClusters()
}

func cleanup() {
	uninstall()
	internal.DeleteKubeSliceDirectory()
}

func basicInstall() {
	internal.VerifyExecutables()
	internal.GenerateKubeSliceDirectory()
	internal.GenerateKindConfiguration()
	internal.CreateKubeConfig()
	internal.SetKubeConfigPath()
	internal.CreateKindClusters()
	internal.InstallCalico()
	internal.PopulateDockerNetworkMap()
	internal.AddHelmCharts()
	internal.InstallCertManager()
	internal.InstallKubeSliceController()
	internal.CreateKubeSliceProject()
	internal.RegisterWorkerClusters()
	internal.InstallKubeSliceWorker()
}

func printHelp() {
	util.Printf(`
KubeSlice CLI for KubeSlice Operations

Options:
  --help				
		Prints this help menu

  --full-install		
		Creates 3 Kind Clusters, sets-up KubeSlice Controller, KubeSlice Worker,
		a demo slice, and iperf example application

  --minimal-install	
		Creates 3 Kind Clusters, sets-up KubeSlice Controller, KubeSlice Worker,
		and iperf example application.
		Once the setup is done, prints the instructions on how to create a slice
		and verify the connectivity.

  --uninstall			
		Deletes the 3 Kind Clusters, but retains the kubeslice configuration directory.

  --cleanup
		Deletes the 3 Kind Clusters and kubeslice directory.
`)
}
