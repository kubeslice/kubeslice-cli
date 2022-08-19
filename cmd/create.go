package cmd

import (
	"github.com/kubeslice/slicectl/pkg"
	"github.com/kubeslice/slicectl/util"
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
			util.Fatalf("Namespace is required")
		}
		filename, _ := cmd.Flags().GetString("filename")
		config, _ := cmd.Flags().GetString("config")
		workerList, _ := cmd.Flags().GetStringSlice("set-worker")
		keyList, _ := cmd.Flags().GetStringSlice("set")
		if len(args) > 1 {
			objectName = args[1]
		}

		pkg.SetCliOptions(pkg.CliParams{Config: config, Namespace: ns, ObjectName: objectName, ObjectType: args[0], FileName: filename, Worker: workerList, Key: keyList})
		switch args[0] {
		case "project":
		//	pkg.CreateProject()
		case "sliceConfig":
			pkg.CreateSliceConfig(filename, workerList, keyList)
		default:
			util.Fatalf("Invalid object type")
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("namespace", "n", "", "namespace")
	createCmd.Flags().StringP("filename", "f", "", "Filename, directory, or URL to file to use to create the resource")
	createCmd.Flags().StringSliceP("set-worker", "w", nil, "WorkerList to install slice config on")
	//createCmd.Flags()
	createCmd.Flags().StringSliceP("set", "s", nil, "set key:values")
}
