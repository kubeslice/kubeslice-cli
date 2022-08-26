package cmd

import (
	"github.com/kubeslice/slicectl/pkg"
	"github.com/kubeslice/slicectl/util"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Get Kubeslice resources.",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var objectName string
		ns, _ := cmd.Flags().GetString("namespace")
		if ns == "" {
			util.Fatalf("Namespace is required")
		}

		if len(args) > 1 {
			objectName = args[1]
		}

		pkg.SetCliOptions(pkg.CliParams{Config: Config, Namespace: ns, ObjectName: objectName, ObjectType: args[0]})
		switch args[0] {
		case "project":
			pkg.GetProject()
		case "sliceConfig":
			pkg.GetSliceConfig()
		case "serviceExportConfig":
			pkg.GetServiceExportConfig()
		default:
			util.Fatalf("Invalid object type")
		}

	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringP("namespace", "n", "", "namespace")
}
