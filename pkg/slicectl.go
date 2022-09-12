package pkg

import (
	"time"

	"github.com/kubeslice/slicectl/pkg/internal"
	"github.com/kubeslice/slicectl/util"
)

func Install(skipSteps map[string]string, ent bool) {
	basicInstall(skipSteps, ent)
	if _, skipDemo := skipSteps[internal.Demo_skipStep]; !skipDemo {
		switch ApplicationConfiguration.Configuration.ClusterConfiguration.Profile {
		case ProfileFullDemo:
			fullDemo()
		case ProfileMinimalDemo:
			minimalDemo()
		}
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

func basicInstall(skipSteps map[string]string, ent bool) {
	internal.VerifyExecutables(ApplicationConfiguration)

	_, skipKind := skipSteps[internal.Kind_skipStep]
	_, skipCalico := skipSteps[internal.Calico_skipStep]
	_, skipController := skipSteps[internal.Controller_skipStep]
	_, skipWorker := skipSteps[internal.Worker_skipStep]
	_, skipWorker_registration := skipSteps[internal.Worker_registration_skipStep]

	internal.GenerateKubeSliceDirectory()
	if ApplicationConfiguration.Configuration.ClusterConfiguration.Profile != "" {
		if !skipKind {
			internal.GenerateKindConfiguration(ApplicationConfiguration)
		}
		internal.CreateKubeConfig()
		internal.SetKubeConfigPath()
		if !skipKind {
			internal.CreateKindClusters(ApplicationConfiguration)
		}
		if !skipCalico {
			internal.InstallCalico(&ApplicationConfiguration.Configuration.ClusterConfiguration)
		}
	}
	internal.GatherNetworkInformation(ApplicationConfiguration)
	internal.AddHelmCharts(ApplicationConfiguration)
	if !skipController {
		internal.InstallCertManager(ApplicationConfiguration)
		internal.InstallKubeSliceController(ApplicationConfiguration)
		internal.CreateKubeSliceProject(ApplicationConfiguration, nil)
	}
	if !skipWorker_registration {
		internal.RegisterWorkerClusters(ApplicationConfiguration, nil)
	}
	if !skipWorker {
		internal.InstallKubeSliceWorker(ApplicationConfiguration)
	}
	if ent {
		internal.InstallKubeSliceUI(ApplicationConfiguration)
	}
}

func Uninstall() {
	if ApplicationConfiguration.Configuration.ClusterConfiguration.Profile == "" {
		util.Fatalf("%s Uninstallation of topology is not yet supported")
	}
	internal.VerifyExecutables(ApplicationConfiguration)
	internal.SetKubeConfigPath()
	internal.DeleteKindClusters(ApplicationConfiguration)
}
