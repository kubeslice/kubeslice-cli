package cmd

import (
	"github.com/kubeslice/kubeslice-cli/pkg"
	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/spf13/cobra"
)

var (
	uninstallAll          bool
	uninstallController   bool
	uninstallEnterprise   bool
	uninstallWorker       = []string{}
	workersToUninstall    map[string]string
	componentsToUninstall map[string]string
)
var uninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Aliases: []string{"cleanup"},
	Short:   "Deletes the Kind Clusters used for the demo.",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.ReadAndValidateConfiguration(Config)
		// if --all flag is passed, other flags should not be allowed
		util.Printf("current values: \nuninstallAll: %v \nuninstallController: %v \nuninstallEnterprise: %v \nuninstallWorker: %v",
			uninstallAll, uninstallController, uninstallEnterprise, len(uninstallWorker) > 1)
		if uninstallAll && (uninstallController || uninstallEnterprise || len(uninstallWorker) > 0) {
			cmd.Help()
			util.Fatalf("\n %v Cannot use other options if --all is passed", util.Cross)
		}

		// if no flags are passed, set uninstallAll true
		if !uninstallAll || !uninstallController || !uninstallEnterprise || (len(uninstallWorker) < 1) {
			uninstallAll = true
		}

		componentsToUninstall = make(map[string]string)
		workersToUninstall = make(map[string]string)

		if uninstallAll {
			uninstallController = true
			uninstallEnterprise = true
			uninstallWorker = []string{"*"}
		}
		if uninstallController {
			componentsToUninstall["controller"] = ""
		}
		if uninstallEnterprise {
			componentsToUninstall["enterprise"] = ""
		}
		if len(uninstallWorker) > 0 {
			componentsToUninstall["worker"] = ""
			workersToUninstall = mapFromSlice(uninstallWorker)
		}

		pkg.Uninstall(componentsToUninstall, workersToUninstall)
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolVarP(&uninstallAll, "all", "A", false, `Uninstalls all components (Worker, Controller, UI)`)
	uninstallCmd.Flags().BoolVarP(&uninstallController, "controller", "", false, `Uninstalls the Controller`)
	uninstallCmd.Flags().BoolVarP(&uninstallEnterprise, "enterprise", "", false, `Uninstalls enterprise components (Kubeslice-Manager)`)
	uninstallCmd.Flags().StringSliceVarP(&uninstallWorker, "worker", "", []string{}, `Uninstalls UI (Kubeslice-Manager)`)
	uninstallCmd.Flags().Lookup("worker").NoOptDefVal = "*"
}
