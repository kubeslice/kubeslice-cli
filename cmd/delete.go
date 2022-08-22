package cmd

import (
	"github.com/kubeslice/slicectl/pkg"
	"github.com/kubeslice/slicectl/util"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d"},
	Short:   "Delete Kubeslice resources.",
	Args:    cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var objectName string
		ns, _ := cmd.Flags().GetString("namespace")
		if ns == "" {
			util.Fatalf("Namespace is required")
		}

		objectName = args[1]

		pkg.SetCliOptions(pkg.CliParams{Config: Config, Namespace: ns, ObjectName: objectName, ObjectType: args[0]})
		switch args[0] {
		case "project":
			pkg.DeleteProject()
		case "sliceConfig":
			pkg.DeleteSliceConfig()
		default:
			util.Fatalf("Invalid object type")
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringP("namespace", "n", "", "namespace")
}
