package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/kubeslice/slicectl/internal"
	"github.com/kubeslice/slicectl/util"
)

func main() {
	configPtr := flag.String("config", "", "The path to topology configuration yaml. Cannot be used with -profile option")
	profile := flag.String("profile", "", "The profile for the installation. Cannot be used with -config option")

	flag.Usage = printHelp
	flag.Parse()

	osArgs := os.Args
	args := flag.Args()

	if len(args) > 0 && args[0] == "help" {
		printHelp()
		return
	}

	if len(args) > 1 {
		util.Fatalf("Invalid arguments %s. Exactly one command is required. Try running `%s help` for more information", args, osArgs[0])
	}

	if *profile != "" {
		switch *profile {
		case internal.ProfileFullDemo:
		case internal.ProfileMinimalDemo:
		default:
			util.Fatalf("Unknown profile: %s. Possible values %s", *profile, []string{internal.ProfileFullDemo, internal.ProfileMinimalDemo})
		}
		internal.ReadAndValidateConfiguration("")
		internal.ApplicationConfiguration.Configuration.ClusterConfiguration.Profile = *profile
	} else {
		internal.ReadAndValidateConfiguration(*configPtr)
	}

	if len(args) == 0 {
		util.Fatalf("Command is required. Try %s help", osArgs[0])
	}
	switch strings.TrimSpace(args[0]) {
	case "install":
		install()
	case "uninstall":
		uninstall()
	case "help":
		printHelp()
	default:
		util.Fatalf("Unknown Command %s", args[0])
	}
}

func install() {
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

func uninstall() {
	if internal.ApplicationConfiguration.Configuration.ClusterConfiguration.Profile == "" {
		util.Fatalf("%s Uninstallation of topology is not yet supported")
	}
	internal.VerifyExecutables()
	internal.SetKubeConfigPath()
	internal.DeleteKindClusters()
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
		internal.PopulateDockerNetworkMap()
	}
	internal.AddHelmCharts()
	internal.InstallCertManager()
	internal.InstallKubeSliceController()
	internal.CreateKubeSliceProject()
	internal.RegisterWorkerClusters()
	internal.InstallKubeSliceWorker()
}

func printHelp() {
	util.Printf(`
slicectl for KubeSlice Operations

Usage:
slicectl <options> <command>

Options:
  --profile=<profile-value>
      The profile for installation/uninstallation.
      Supported values:
        - full-demo:
            Showcases the KubeSlice inter-cluster connectivity by spawning
            3 Kind Clusters, including 1 KubeSlice Controller and 2 KubeSlice Workers, 
            and installing iPerf application to generate network traffic.
        - minimal-demo:
            Sets up 3 Kind Clusters, including 1 KubeSlice Controller and 2 KubeSlice Workers. 
            Generates the KubernetesManifests for user to manually apply, and verify 
            the functionality

  --config=<path-to-topology-configuration-yaml>
      The yaml file with topology configuration. 
      Refer: https://github.com/kubeslice/slicectl/blob/master/samples/template.yaml

Commands:
  install
      Creates 3 Kind Clusters, sets-up KubeSlice Controller, KubeSlice Worker,
      and iperf example application.
      Once the setup is done, prints the instructions on how to create a slice
      and verify the connectivity.

  uninstall
      Deletes the Kind Clusters used for the demo.

  help
      Prints this help menu.
`)
}
