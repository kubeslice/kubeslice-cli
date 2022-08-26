package cmd

import (
	"github.com/kubeslice/slicectl/pkg"
	"github.com/kubeslice/slicectl/util"
	"github.com/spf13/cobra"
)

var (
	profile string
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
			util.Fatalf("Cannot use both -config and -profile options")
		}
		if profile != "" {
			switch profile {
			case pkg.ProfileFullDemo:
			case pkg.ProfileMinimalDemo:
			default:
				util.Fatalf("Unknown profile: %s. Possible values %s", profile, []string{pkg.ProfileFullDemo, pkg.ProfileMinimalDemo})
			}
			pkg.ReadAndValidateConfiguration("")
			pkg.ApplicationConfiguration.Configuration.ClusterConfiguration.Profile = profile
		} else {
			pkg.ReadAndValidateConfiguration(Config)
		}
		pkg.Install()
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
		the functionality`)
}
