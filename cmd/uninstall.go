package cmd

import (
	"github.com/kubeslice/kubeslice-cli/pkg"
	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/spf13/cobra"
)

var (
	uninstallAll          bool
	uninstallController   bool
	uninstallUI           bool
	uninstallCertManager  bool
	uninstallWorker       = []string{}
	workersToUninstall    map[string]string
	componentsToUninstall map[string]string
)
var uninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Aliases: []string{"cleanup"},
	Short:   "Performs cleanup of Kubeslice components.",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.ReadAndValidateConfiguration(Config, "")
		// if --all flag is passed, other flags should not be allowed
		if uninstallAll && uninstallUI {
			cmd.Help()
			util.Fatalf("\n %v Cannot use other options if --all is passed", util.Cross)
		}

		// if no flags are passed, set uninstallAll true
		if !uninstallAll && !uninstallUI {
			uninstallAll = true
		}

		componentsToUninstall = make(map[string]string)
		workersToUninstall = make(map[string]string)

		if uninstallAll {
			uninstallController = true
			uninstallUI = true
			uninstallWorker = []string{"*"}
		}
		if uninstallController {
			componentsToUninstall["controller"] = ""
		}
		if uninstallUI {
			componentsToUninstall["ui"] = ""
		}
		if uninstallCertManager {
			componentsToUninstall["cert-manager"] = ""
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
	uninstallCmd.Flags().BoolVarP(&uninstallAll, "all", "a", false, `Uninstalls all components (Worker, Controller, UI)`)
	uninstallCmd.Flags().BoolVarP(&uninstallUI, "ui", "u", false, `Uninstalls enterprise UI components (Kubeslice-Manager)`)
	// TODO: update the controller version after release
	uninstallCmd.Flags().BoolVarP(&uninstallCertManager, "cert-manager", "", false, `Uninstalls Cert Manager (required for controller version < 0.7.0)`)
	// TODO: A discussion is needed for graceful cleanup of worker clusters
	// uninstallCmd.Flags().StringSliceVarP(&uninstallWorker, "worker", "", []string{}, `Uninstalls worker clusters`)
	// uninstallCmd.Flags().Lookup("worker").NoOptDefVal = "*"
}
