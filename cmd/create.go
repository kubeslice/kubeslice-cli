package cmd

import (
	"github.com/kubeslice/kubeslice-cli/pkg"
	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Kubeslice resources.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var objectName string
		ns, _ := cmd.Flags().GetString("namespace")
		if ns == "" {
			util.PrintCliError(&util.CliError{
				Msg:        "Namespace is required",
				Context:    "create command",
				Suggestion: "Use the --namespace or -n flag to specify the namespace.",
			})
		}
		filename, _ := cmd.Flags().GetString("filename")
		workerList, _ := cmd.Flags().GetStringSlice("setWorker")
		if len(args) > 1 {
			objectName = args[1]
		}
		pkg.SetCliOptions(pkg.CliParams{Config: Config, Namespace: ns, ObjectName: objectName, ObjectType: args[0], FileName: filename})
		switch args[0] {
		case "project":
			pkg.CreateProject()
		case "sliceConfig":
			pkg.CreateSliceConfig(workerList)
		case "serviceExportConfig":
			pkg.CreateServiceExportConfig(filename)
		default:
			util.PrintCliError(&util.CliError{
				Msg:        "Invalid object type",
				Context:    "create command",
				Suggestion: "Valid types: project, sliceConfig, serviceExportConfig.",
			})
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("namespace", "n", "", "namespace")
	createCmd.Flags().StringP("filename", "f", "", "Filename, directory, or URL to file to use to create the resource")
	createCmd.Flags().StringSliceP("setWorker", "w", nil, "List of Worker Clusters to be registered in the SliceConfig")
}
