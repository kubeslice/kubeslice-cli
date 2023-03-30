package pkg

import (
	"time"

	"github.com/kubeslice/kubeslice-cli/pkg/internal"
	"github.com/kubeslice/kubeslice-cli/util"
)

func Install(skipSteps map[string]string) {
	basicInstall(skipSteps)
	if _, skipDemo := skipSteps[internal.Demo_Component]; !skipDemo {
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

func basicInstall(skipSteps map[string]string) {
	internal.VerifyExecutables(ApplicationConfiguration)

	_, skipKind := skipSteps[internal.Kind_Component]
	_, skipCalico := skipSteps[internal.Calico_Component]
	_, skipController := skipSteps[internal.Controller_Component]
	_, skipWorker := skipSteps[internal.Worker_Component]
	_, skipWorker_registration := skipSteps[internal.Worker_registration_Component]
	_, skipUI := skipSteps[internal.UI_install_Component]
	_, skipCertManager := skipSteps[internal.CertManager_Component]

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
	}
	if (ApplicationConfiguration.Configuration.ClusterConfiguration.Profile != "" || ApplicationConfiguration.Configuration.ClusterConfiguration.ClusterType == "kind") && !skipCalico {
		internal.InstallCalico(&ApplicationConfiguration.Configuration.ClusterConfiguration)
	}
	internal.GatherNetworkInformation(ApplicationConfiguration)
	internal.AddHelmCharts(ApplicationConfiguration)
	if !skipController {
		if !skipCertManager {
			internal.InstallCertManager(ApplicationConfiguration)
		}
		internal.InstallKubeSliceController(ApplicationConfiguration)
		internal.CreateKubeSliceProject(ApplicationConfiguration, nil)
	}
	if !skipWorker_registration {
		internal.RegisterWorkerClusters(ApplicationConfiguration, nil)
	}
	if !skipWorker {
		internal.InstallKubeSliceWorker(ApplicationConfiguration)
	}
	if !skipUI {
		internal.InstallKubeSliceUI(ApplicationConfiguration)
	}
}

func Uninstall(componentsToUninstall, workersToUninstall map[string]string) {

	internal.VerifyExecutables(ApplicationConfiguration)

	// Custom topology passed
	if ApplicationConfiguration.Configuration.ClusterConfiguration.Profile == "" {
		_, uninstallController := componentsToUninstall[internal.Controller_Component]
		_, uninstallCertManager := componentsToUninstall[internal.CertManager_Component]
		_, uninstallWorker := componentsToUninstall[internal.Worker_Component]
		_, uninstallUI := componentsToUninstall[internal.UI_install_Component]

		if uninstallUI {
			internal.UninstallKubeSliceUI(ApplicationConfiguration)
		}
		if uninstallWorker {
			internal.UninstallKubeSliceWorker(ApplicationConfiguration, workersToUninstall)
		}
		if uninstallController {
			internal.UninstallKubeSliceController(ApplicationConfiguration)
			if uninstallCertManager {
				internal.UninstallCertManager(ApplicationConfiguration)
			}
		}
		return
	}
	// Cleanup setup of Minimal/Full Demo.
	internal.SetKubeConfigPath()
	internal.DeleteKindClusters(ApplicationConfiguration)
}
