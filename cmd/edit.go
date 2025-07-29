package cmd

import (
	"github.com/kubeslice/kubeslice-cli/pkg"
	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:     "edit",
	Aliases: []string{"e"},
	Short:   "Edit Kubeslice resources.",
	Long: `The edit command allows you to directly edit any Kubeslice resource you can retrieve via the command line tools. It will open the editor defined by your KUBE_EDITOR, or EDITOR environment variables, or fall back to ‘vi’ for Linux or ‘notepad’ for Windows. You can edit multiple objects, although changes are applied one at a time. The command accepts filenames as well as command line arguments, although the files you point to must be previously saved versions of resources.
	The default format is YAML.
	In the event an error occurs while updating, a temporary file will be created on disk that contains your unapplied changes. The most common error when updating a resource is another editor changing the resource on the server. When this occurs, you will have to apply your changes to the newer version of the resource, or update your temporary saved copy to include the latest resource version.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var objectName string
		ns, _ := cmd.Flags().GetString("namespace")
		if ns == "" {
			util.PrintCliError(&util.CliError{
				Msg:        "Namespace is required",
				Context:    "edit command",
				Suggestion: "Use the --namespace or -n flag to specify the namespace.",
			})
		}
		filename, _ := cmd.Flags().GetString("filename")

		if len(args) > 1 {
			objectName = args[1]
		}

		pkg.SetCliOptions(pkg.CliParams{Config: Config, Namespace: ns, ObjectName: objectName, ObjectType: args[0], FileName: filename})
		switch args[0] {
		case "project":
			pkg.EditProject()
		case "sliceConfig":
			pkg.EditSliceConfig()
		case "serviceExportConfig":
			pkg.EditServiceExportConfig()
		case "worker":
			pkg.EditWorker()
		default:
			util.PrintCliError(&util.CliError{
				Msg:        "Invalid object type",
				Context:    "edit command",
				Suggestion: "Valid types: project, sliceConfig, serviceExportConfig, worker.",
			})
		}
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.Flags().StringP("namespace", "n", "", "namespace")
	editCmd.Flags().StringP("filename", "f", "", "Filename, directory, or URL to file to use to create the resource")
}
