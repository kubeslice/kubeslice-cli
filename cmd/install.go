package cmd

import (
	"github.com/kubeslice/kubeslice-cli/pkg"
	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"i"},
	Short:   "Installs workloads to run KubeSlice",
	Long: `Installs the required workloads to run KubeSlice Controller and KubeSlice Worker.
	Additional example applications are also installed in demo profiles to showcase the
	KubeSlice functionality`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// check if config and profile are both set, if so, error out
		if Config != "" && profile != "" {
			cmd.Help()
			util.Fatalf("\n %v Cannot use both --config and --profile options", util.Cross)
		}
		// check if config and profile are both not set, if so, error out
		if Config == "" && profile == "" {
			cmd.Help()
			util.Fatalf("\n %v Please pass either --config or --profile option", util.Cross)
		}
		if profile != "" {
			switch profile {
			case pkg.ProfileFullDemo:
			case pkg.ProfileMinimalDemo:
			default:
				util.Fatalf("%v Unknown profile: %s. Possible values %s", util.Cross, profile, []string{pkg.ProfileFullDemo, pkg.ProfileMinimalDemo})
			}
			pkg.ReadAndValidateConfiguration("")
			pkg.ApplicationConfiguration.Configuration.ClusterConfiguration.Profile = profile
		} else {
			pkg.ReadAndValidateConfiguration(Config)
		}
		stepsToSkipMap := mapFromSlice(skipSteps)
		pkg.Install(stepsToSkipMap)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringVarP(&profile, "profile", "p", "", `<profile-value>
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
Cannot be used with --config flag.`)
	installCmd.Flags().StringSliceVarP(&skipSteps, "skip", "s", []string{}, `Skips the installation steps (comma-seperated). 
Supported values:
	- kind: Skips the creation of kind clusters
	- calico: Skips the installation of Calico
	- controller: Skips the installation of KubeSlice Controller
	- worker-registration: Skips the registration of KubeSlice Workers on the Controller
	- worker: Skips the installation of KubeSlice Worker
	- demo: Skips the installation of additional example applications
	- ui: Skips the installtion of enterprise UI components (Kubeslice-Manager)`)

}
