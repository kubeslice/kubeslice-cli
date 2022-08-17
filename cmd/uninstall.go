package cmd

import (
	"github.com/kubeslice/slicectl/pkg"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Aliases: []string{"cleanup"},
	Short:   "Deletes the Kind Clusters used for the demo.",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		config, _ := cmd.Flags().GetString("config")
		pkg.ReadAndValidateConfiguration(config)
		pkg.Uninstall()
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
