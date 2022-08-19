package pkg

import (
	"time"

	"github.com/kubeslice/slicectl/pkg/internal"
	"github.com/kubeslice/slicectl/util"
)

func Install() {
	basicInstall()
	switch ApplicationConfiguration.Configuration.ClusterConfiguration.Profile {
	case ProfileFullDemo:
		fullDemo()
	case ProfileMinimalDemo:
		minimalDemo()
	}
}

func fullDemo() {
	internal.GenerateSliceConfiguration(ApplicationConfiguration, nil, "", "")
	internal.ApplySliceConfiguration(ApplicationConfiguration)
	util.Printf("%s Waiting for configuration propagation", util.Wait)
	time.Sleep(20 * time.Second)
	internal.GenerateIPerfManifests()
	internal.GenerateIPerfServiceExportManifest(ApplicationConfiguration)
	internal.InstallIPerf(ApplicationConfiguration)
	internal.ApplyIPerfServiceExportManifest(ApplicationConfiguration)
	util.Printf("%s Waiting for configuration propagation", util.Wait)
	time.Sleep(20 * time.Second)
	internal.RolloutRestartIPerf(ApplicationConfiguration)
	internal.PrintNextSteps(true, ApplicationConfiguration)
}

func minimalDemo() {
	internal.GenerateSliceConfiguration(ApplicationConfiguration, nil, "", "")
	internal.GenerateIPerfManifests()
	internal.InstallIPerf(ApplicationConfiguration)
	internal.GenerateIPerfServiceExportManifest(ApplicationConfiguration)
	internal.PrintNextSteps(false, ApplicationConfiguration)
}

func basicInstall() {
	internal.VerifyExecutables(ApplicationConfiguration)
	internal.GenerateKubeSliceDirectory()
	if ApplicationConfiguration.Configuration.ClusterConfiguration.Profile != "" {
		internal.GenerateKindConfiguration(ApplicationConfiguration)
		internal.CreateKubeConfig()
		internal.SetKubeConfigPath()
		internal.CreateKindClusters(ApplicationConfiguration)
		internal.InstallCalico(ApplicationConfiguration.Configuration.ClusterConfiguration)
	}
	internal.GatherNetworkInformation(ApplicationConfiguration)
	internal.AddHelmCharts(ApplicationConfiguration)
	internal.InstallCertManager(ApplicationConfiguration)
	internal.InstallKubeSliceController(ApplicationConfiguration)
	internal.CreateKubeSliceProject(ApplicationConfiguration)
	internal.RegisterWorkerClusters(ApplicationConfiguration)
	internal.InstallKubeSliceWorker(ApplicationConfiguration)
}

func Uninstall() {
	if ApplicationConfiguration.Configuration.ClusterConfiguration.Profile == "" {
		util.Fatalf("%s Uninstallation of topology is not yet supported")
	}
	internal.VerifyExecutables(ApplicationConfiguration)
	internal.SetKubeConfigPath()
	internal.DeleteKindClusters(ApplicationConfiguration)
}
