package pkg

import (
	"time"

	"github.com/kubeslice/slicectl/internal"
	"github.com/kubeslice/slicectl/util"
)

func Install() {
	basicInstall()
	switch internal.ApplicationConfiguration.Configuration.ClusterConfiguration.Profile {
	case internal.ProfileFullDemo:
		fullDemo()
	case internal.ProfileMinimalDemo:
		minimalDemo()
	}
}

func fullDemo() {
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

func minimalDemo() {
	internal.GenerateSliceConfiguration()
	internal.GenerateIPerfManifests()
	internal.InstallIPerf()
	internal.GenerateIPerfServiceExportManifest()
	internal.PrintNextSteps(false)
}

func basicInstall() {
	internal.VerifyExecutables()
	internal.GenerateKubeSliceDirectory()
	if internal.ApplicationConfiguration.Configuration.ClusterConfiguration.Profile != "" {
		internal.GenerateKindConfiguration()
		internal.CreateKubeConfig()
		internal.SetKubeConfigPath()
		internal.CreateKindClusters()
		internal.InstallCalico()
	}
	internal.GatherNetworkInformation()
	internal.AddHelmCharts()
	internal.InstallCertManager()
	internal.InstallKubeSliceController()
	internal.CreateKubeSliceProject()
	internal.RegisterWorkerClusters()
	internal.InstallKubeSliceWorker()
}

func Uninstall() {
	if internal.ApplicationConfiguration.Configuration.ClusterConfiguration.Profile == "" {
		util.Fatalf("%s Uninstallation of topology is not yet supported")
	}
	internal.VerifyExecutables()
	internal.SetKubeConfigPath()
	internal.DeleteKindClusters()
}
