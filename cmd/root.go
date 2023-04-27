package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	// "github.com/spf13/cobra/doc"
)

var version = "0.5.1"
var rootCmd = &cobra.Command{
	Use:     "kubeslice-cli",
	Version: version,
	Short:   "kubeslice-cli - a simple CLI for KubeSlice Operations",
	Long: `kubeslice-cli - a simple CLI for KubeSlice Operations
    
Use kubeslice-cli to install/uninstall required workloads to run KubeSlice Controller and KubeSlice Worker.
Additional example applications can also be installed in demo profiles to showcase the
KubeSlice functionality`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	rootCmd.PersistentFlags().StringVarP(&Config, "config", "c", "", `<path-to-topology-configuration-yaml-file>
	The yaml file with topology configuration. 
	Refer: https://github.com/kubeslice/kubeslice-cli/blob/master/samples/template.yaml`)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing kubeslice-cli '%s'", err)
		os.Exit(1)
	}
	//  Uncomment to generate docs for new commands/flags
	// doc.GenMarkdownTree(rootCmd, "doc")

}
