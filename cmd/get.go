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
		if ns == "" {
			util.Fatalf("Namespace is required")
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
		default:
			util.Fatalf("Invalid object type")
		}

	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringP("namespace", "n", "", "namespace")
	getCmd.Flags().StringP("worker", "w", "", "worker")
	getCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "supported values json, yaml")
}
