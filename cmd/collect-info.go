package cmd

import (
	"github.com/kubeslice/kubeslice-cli/pkg"
	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/spf13/cobra"
)

var (
	collectNamespace  string
	collectKubeconfig string
	collectOutputFile string
	collectComponents []string
	collectLogs       bool
	collectConfigs    bool
	collectEvents     bool
	collectMetrics    bool
)

var collectInfoCmd = &cobra.Command{
	Use:     "collect-info",
	Aliases: []string{"ci", "collect"},
	Short:   "Collect debug information from KubeSlice components",
	Long: `Collect comprehensive debug information from KubeSlice components in the cluster.
	
This command gathers logs, configurations, events, and other debug information from:
- KubeSlice Controller pods
- KubeSlice Worker pods  
- KubeSlice Manager (enterprise)
- Calico networking components
- Prometheus monitoring (if installed)
- Kubernetes resources (CRDs, ConfigMaps, Secrets)

The output is packaged into a compressed tar.gz file for easy sharing and analysis.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate required flags
		if collectNamespace == "" {
			cmd.Help()
			util.Fatalf("\n [ERROR] Namespace is required. Use -n or --namespace flag")
		}

		if collectOutputFile == "" {
			cmd.Help()
			util.Fatalf("\n [ERROR] Output file is required. Use --out flag")
		}

		// Set default components if none specified
		if len(collectComponents) == 0 {
			collectComponents = []string{"controller", "worker", "manager", "calico", "prometheus", "resources"}
		}

		// Set default collection types if none specified
		if !collectLogs && !collectConfigs && !collectEvents && !collectMetrics {
			collectLogs = true
			collectConfigs = true
			collectEvents = true
		}

		// Execute collection
		pkg.CollectInfo(collectNamespace, collectKubeconfig, collectOutputFile, collectComponents, collectLogs, collectConfigs, collectEvents, collectMetrics)
	},
}

func init() {
	rootCmd.AddCommand(collectInfoCmd)

	// Required flags
	collectInfoCmd.Flags().StringVarP(&collectNamespace, "namespace", "n", "", "Namespace containing KubeSlice components (required)")
	collectInfoCmd.Flags().StringVar(&collectOutputFile, "out", "", "Output file path for the collected information archive (required)")

	// Optional flags
	collectInfoCmd.Flags().StringVar(&collectKubeconfig, "kubeconfig", "", "Path to kubeconfig file (uses default if not specified)")
	collectInfoCmd.Flags().StringSliceVar(&collectComponents, "components", []string{}, `Components to collect information from (comma-separated).
Supported values:
	- controller: KubeSlice Controller components
	- worker: KubeSlice Worker components
	- manager: KubeSlice Manager (enterprise)
	- calico: Calico networking components
	- prometheus: Prometheus monitoring components
	- resources: Kubernetes resources (CRDs, ConfigMaps, Secrets)
	- all: All components (default if none specified)`)

	collectInfoCmd.Flags().BoolVar(&collectLogs, "logs", false, "Collect pod logs (default: true if no collection type specified)")
	collectInfoCmd.Flags().BoolVar(&collectConfigs, "configs", false, "Collect configuration files and manifests (default: true if no collection type specified)")
	collectInfoCmd.Flags().BoolVar(&collectEvents, "events", false, "Collect Kubernetes events (default: true if no collection type specified)")
	collectInfoCmd.Flags().BoolVar(&collectMetrics, "metrics", false, "Collect metrics and status information (default: false)")

	// Mark required flags
	collectInfoCmd.MarkFlagRequired("namespace")
	collectInfoCmd.MarkFlagRequired("out")
}
