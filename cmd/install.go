package cmd

import (
	"github.com/kubeslice/slicectl/internal"
	"github.com/kubeslice/slicectl/pkg"
	"github.com/kubeslice/slicectl/util"
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
		config, _ := cmd.Flags().GetString("config")
		profile, _ := cmd.Flags().GetString("profile")
		// check if config and profile are both set, if so, error out
		if config != "" && profile != "" {
			util.Fatalf("Cannot use both -config and -profile options")
		}
		if profile != "" {
			switch profile {
			case internal.ProfileFullDemo:
			case internal.ProfileMinimalDemo:
			default:
				util.Fatalf("Unknown profile: %s. Possible values %s", profile, []string{internal.ProfileFullDemo, internal.ProfileMinimalDemo})
			}
			internal.ReadAndValidateConfiguration("")
			internal.ApplicationConfiguration.Configuration.ClusterConfiguration.Profile = profile
		} else {
			internal.ReadAndValidateConfiguration(config)
		}
		pkg.Install()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringP("profile", "p", "", `<profile-value>
The profile for installation/uninstallation.
Supported values:
	- full-demo:
		Showcases the KubeSlice inter-cluster connectivity by spawning
		3 Kind Clusters, including 1 KubeSlice Controller and 2 KubeSlice Workers, 
		and installing iPerf application to generate network traffic.
	- minimal-demo:
		Sets up 3 Kind Clusters, including 1 KubeSlice Controller and 2 KubeSlice Workers. 
		Generates the KubernetesManifests for user to manually apply, and verify 
		the functionality`)
}
