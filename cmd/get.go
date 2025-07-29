package cmd

import (
	"github.com/kubeslice/kubeslice-cli/pkg"
	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Get Kubeslice resources.",
	Args:    cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var objectName string
		ns, _ := cmd.Flags().GetString("namespace")
		if ns == "" && args[0] != "ui-endpoint" {
			util.PrintCliError(&util.CliError{
				Msg:        "Namespace is required",
				Context:    "get command",
				Suggestion: "Use the --namespace or -n flag to specify the namespace.",
			})
		}
		worker, _ := cmd.Flags().GetString("worker")
		if len(args) > 1 {
			objectName = args[1]
		}

		pkg.SetCliOptions(pkg.CliParams{Config: Config, Namespace: ns, ObjectName: objectName, ObjectType: args[0], OutputFormat: outputFormat})
		switch args[0] {
		case "project":
			pkg.GetProject()
		case "sliceConfig":
			pkg.GetSliceConfig()
		case "serviceExportConfig":
			pkg.GetServiceExportConfig()
		case "secrets":
			pkg.GetSecrets(worker)
		case "worker":
			pkg.GetWorker()
		case "ui-endpoint":
			pkg.GetUIEndpoint()
		default:
			util.PrintCliError(&util.CliError{
				Msg:        "Invalid object type",
				Context:    "get command",
				Suggestion: "Valid types: project, sliceConfig, serviceExportConfig, secrets, worker, ui-endpoint.",
			})
		}

	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringP("namespace", "n", "", "namespace")
	getCmd.Flags().StringP("worker", "w", "", "worker")
	getCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "supported values json, yaml")
}
