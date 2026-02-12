package cmd

import (
	"github.com/kubeslice/kubeslice-cli/pkg"
	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/spf13/cobra"
)

var showHealthCmd = &cobra.Command{
	Use:     "show-health",
	Aliases: []string{"health"},
	Short:   "Show health status of KubeSlice resources.",
	Args:    cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var objectName string
		ns, _ := cmd.Flags().GetString("namespace")
		if ns == "" {
			util.Fatalf("Namespace is required")
		}
		all, _ := cmd.Flags().GetBool("all")
		
		if len(args) > 1 {
			objectName = args[1]
		}

		pkg.SetCliOptions(pkg.CliParams{Config: Config, Namespace: ns, ObjectName: objectName, ObjectType: args[0], OutputFormat: outputFormat})
		switch args[0] {
		case "slice":
			if all {
				pkg.ShowAllSliceHealth()
			} else {
				if objectName == "" {
					util.Fatalf("Slice name is required when not using --all flag")
				}
				pkg.ShowSliceHealth()
			}
		default:
			util.Fatalf("Invalid object type. Supported types: slice")
		}
	},
}

func init() {
	rootCmd.AddCommand(showHealthCmd)
	showHealthCmd.Flags().StringP("namespace", "n", "", "namespace")
	showHealthCmd.Flags().BoolP("all", "A", false, "show health for all slices")
	showHealthCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "supported values json, yaml")
} 