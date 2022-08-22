package cmd

import (
	"github.com/kubeslice/slicectl/pkg"
	"github.com/kubeslice/slicectl/util"
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe Kubeslice resources.",
	Long:  "Show details of a specific Kubeslice resource or group of resources.",
	Args:  cobra.MinimumNArgs(1),
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
			pkg.DescribeProject()
		case "sliceConfig":
			pkg.DescribeSliceConfig()
		default:
			util.Fatalf("Invalid object type")
		}
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)
	describeCmd.Flags().StringP("namespace", "n", "", "namespace")
}
