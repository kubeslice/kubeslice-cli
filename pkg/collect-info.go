package pkg

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-cli/pkg/internal"
	"github.com/kubeslice/kubeslice-cli/util"
)

// CollectInfo orchestrates the collection of debug information from KubeSlice components
func CollectInfo(namespace, kubeconfig, outputFile string, components []string, collectLogs, collectConfigs, collectEvents, collectMetrics bool) {
	util.Printf("[INFO] Starting KubeSlice debug information collection")
	util.Printf("[INFO] Target namespace: %s", namespace)
	util.Printf("[INFO] Output file: %s", outputFile)

	// Create temporary directory for collection
	tempDir := fmt.Sprintf("kubeslice-debug-%s", time.Now().Format("20060102-150405"))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		util.Fatalf("[ERROR] Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	util.Printf("[SUCCESS] Created temporary directory: %s", tempDir)

	// Initialize collection context
	ctx := &internal.CollectionContext{
		Namespace:      namespace,
		Kubeconfig:     kubeconfig,
		TempDir:        tempDir,
		Components:     components,
		CollectLogs:    collectLogs,
		CollectConfigs: collectConfigs,
		CollectEvents:  collectEvents,
		CollectMetrics: collectMetrics,
	}

	// Collect information for each component
	for _, component := range components {
		switch strings.ToLower(component) {
		case "controller", "ctrl":
			util.Printf("[INFO] Collecting Controller information")
			internal.CollectControllerInfo(ctx)
		case "worker", "wrk":
			util.Printf("[INFO] Collecting Worker information")
			internal.CollectWorkerInfo(ctx)
		case "manager", "mgr":
			util.Printf("[INFO] Collecting Manager information")
			internal.CollectManagerInfo(ctx)
		case "calico":
			util.Printf("[INFO] Collecting Calico information")
			internal.CollectCalicoInfo(ctx)
		case "prometheus", "prom":
			util.Printf("[INFO] Collecting Prometheus information")
			internal.CollectPrometheusInfo(ctx)
		case "resources", "res":
			util.Printf("[INFO] Collecting Kubernetes resources")
			internal.CollectKubernetesResources(ctx)
		case "all":
			util.Printf("[INFO] Collecting all component information")
			internal.CollectAllComponents(ctx)
		default:
			util.Printf("[WARN] Unknown component: %s, skipping", component)
		}
	}

	// Create the final archive
	util.Printf("[INFO] Creating debug information archive")
	if err := internal.CreateArchive(tempDir, outputFile); err != nil {
		util.Fatalf("[ERROR] Failed to create archive: %v", err)
	}

	// Verify the output file
	if info, err := os.Stat(outputFile); err == nil {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		util.Printf("[SUCCESS] Successfully created debug archive: %s (%.2f MB)", outputFile, sizeMB)
	} else {
		util.Printf("[WARN] Archive created but could not verify size: %s", outputFile)
	}

	util.Printf("[SUCCESS] Debug information collection completed successfully!")
	util.Printf("[INFO] You can now share the archive file: %s", outputFile)
}
