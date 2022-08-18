package cmd

import (
	"github.com/kubeslice/slicectl/pkg"
	"github.com/kubeslice/slicectl/util"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d"},
	Short:   "delete kubeslice CRD objects.",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var objectName string
		ns, _ := cmd.Flags().GetString("namespace")
		if ns == "" {
			util.Fatalf("Namespace is required")
		}
		config, _ := cmd.Flags().GetString("config")

		if len(args) > 1 {
			objectName = args[1]
		}

		pkg.SetCliOptions(config, ns, args[0], objectName)
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
