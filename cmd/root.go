package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.0.1"
var rootCmd = &cobra.Command{
	Use:     "slicectl",
	Version: version,
	Short:   "slicectl - a simple CLI for KubeSlice Operations",
	Long: `slicectl - a simple CLI for KubeSlice Operations
    
Use slicectl to install/uninstall required workloads to run KubeSlice Controller and KubeSlice Worker.
Additional example applications can also be installed in demo profiles to showcase the
KubeSlice functionality`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	rootCmd.PersistentFlags().StringP("config", "c", "", `<path-to-topology-configuration-yaml-file>
	The yaml file with topology configuration. 
	Refer: https://github.com/kubeslice/slicectl/blob/master/samples/template.yaml`)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing slicectl '%s'", err)
		os.Exit(1)
	}

}
