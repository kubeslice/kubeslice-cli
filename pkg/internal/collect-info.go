package internal

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kubeslice/kubeslice-cli/util"
)

// CollectionContext holds the context for collecting debug information
type CollectionContext struct {
	Namespace      string
	Kubeconfig     string
	TempDir        string
	Components     []string
	CollectLogs    bool
	CollectConfigs bool
	CollectEvents  bool
	CollectMetrics bool
}

// findKubectlPath finds the path to kubectl executable
func findKubectlPath() string {
	// First try to find kubectl in PATH
	if path, err := exec.LookPath("kubectl"); err == nil {
		return path
	}

	// Fallback to common Windows paths
	if runtime.GOOS == "windows" {
		commonPaths := []string{
			"C:\\ProgramData\\chocolatey\\bin\\kubectl.exe",
			"C:\\Program Files\\Docker\\Docker\\resources\\bin\\kubectl.exe",
			"C:\\Program Files\\Kubernetes\\Minikube\\kubectl.exe",
		}

		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	// Default fallback
	return "kubectl"
}

// runKubectlCommand runs a kubectl command and returns the output
func runKubectlCommand(args []string) (string, string, error) {
	kubectlPath := findKubectlPath()
	util.Printf("[INFO] Using kubectl at: %s", kubectlPath)

	cmd := exec.Command(kubectlPath, args...)
	var outB, errB bytes.Buffer
	cmd.Stdout = &outB
	cmd.Stderr = &errB

	err := cmd.Run()
	return outB.String(), errB.String(), err
}

// CollectControllerInfo collects information from KubeSlice Controller components
func CollectControllerInfo(ctx *CollectionContext) {
	componentDir := filepath.Join(ctx.TempDir, "controller")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		util.Printf("[ERROR] Failed to create controller directory: %v", ctx.TempDir)
		return
	}

	// Collect controller pod logs
	if ctx.CollectLogs {
		collectPodLogs(ctx, componentDir, "control-plane=controller-manager")
	}

	// Collect controller configurations
	if ctx.CollectConfigs {
		collectPodConfigs(ctx, componentDir, "control-plane=controller-manager")
	}

	// Collect controller events
	if ctx.CollectEvents {
		collectPodEvents(ctx, componentDir, "control-plane=controller-manager")
	}

	util.Printf("[SUCCESS] Controller information collected in: %s", componentDir)
}

// CollectWorkerInfo collects information from KubeSlice Worker components
func CollectWorkerInfo(ctx *CollectionContext) {
	componentDir := filepath.Join(ctx.TempDir, "worker")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		util.Printf("[ERROR] Failed to create worker directory: %v", ctx.TempDir)
		return
	}

	// Collect worker pod logs
	if ctx.CollectLogs {
		collectPodLogs(ctx, componentDir, "app=kubeslice-worker")
	}

	// Collect worker configurations
	if ctx.CollectConfigs {
		collectPodConfigs(ctx, componentDir, "app=kubeslice-worker")
	}

	// Collect worker events
	if ctx.CollectEvents {
		collectPodEvents(ctx, componentDir, "app=kubeslice-worker")
	}

	util.Printf("[SUCCESS] Worker information collected in: %s", componentDir)
}

// CollectManagerInfo collects information from KubeSlice Manager (enterprise)
func CollectManagerInfo(ctx *CollectionContext) {
	componentDir := filepath.Join(ctx.TempDir, "manager")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		util.Printf("[ERROR] Failed to create manager directory: %v", ctx.TempDir)
		return
	}

	// Collect manager pod logs
	if ctx.CollectLogs {
		collectPodLogs(ctx, componentDir, "app=kubeslice-manager")
	}

	// Collect manager configurations
	if ctx.CollectConfigs {
		collectPodConfigs(ctx, componentDir, "app=kubeslice-manager")
	}

	util.Printf("[SUCCESS] Manager information collected in: %s", componentDir)
}

// CollectCalicoInfo collects information from Calico networking components
func CollectCalicoInfo(ctx *CollectionContext) {
	componentDir := filepath.Join(ctx.TempDir, "calico")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		util.Printf("[ERROR] Failed to create calico directory: %v", ctx.TempDir)
		return
	}

	// Collect calico pod logs
	if ctx.CollectLogs {
		collectPodLogs(ctx, componentDir, "k8s-app=calico-node")
	}

	// Collect calico configurations
	if ctx.CollectConfigs {
		collectPodConfigs(ctx, componentDir, "k8s-app=calico-node")
	}

	util.Printf("[SUCCESS] Calico information collected in: %s", componentDir)
}

// CollectPrometheusInfo collects information from Prometheus monitoring components
func CollectPrometheusInfo(ctx *CollectionContext) {
	componentDir := filepath.Join(ctx.TempDir, "prometheus")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		util.Printf("[ERROR] Failed to create prometheus directory: %v", ctx.TempDir)
		return
	}

	// Collect prometheus pod logs
	if ctx.CollectLogs {
		collectPodLogs(ctx, componentDir, "app=prometheus")
	}

	// Collect prometheus configurations
	if ctx.CollectConfigs {
		collectPodConfigs(ctx, componentDir, "app=prometheus")
	}

	util.Printf("[SUCCESS] Prometheus information collected in: %s", componentDir)
}

// CollectKubernetesResources collects Kubernetes resources information
func CollectKubernetesResources(ctx *CollectionContext) {
	componentDir := filepath.Join(ctx.TempDir, "resources")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		util.Printf("[ERROR] Failed to create resources directory: %v", ctx.TempDir)
		return
	}

	// Collect CRDs
	collectResource(ctx, componentDir, "crd", "kubeslice.io")

	// Collect ConfigMaps
	collectResource(ctx, componentDir, "configmap", "")

	// Collect Secrets (names only, not content)
	collectResource(ctx, componentDir, "secret", "")

	// Collect Services
	collectResource(ctx, componentDir, "service", "")

	// Collect Deployments
	collectResource(ctx, componentDir, "deployment", "")

	util.Printf("[SUCCESS] Kubernetes resources collected in: %s", componentDir)
}

// CollectAllComponents collects information from all components
func CollectAllComponents(ctx *CollectionContext) {
	allComponents := []string{"controller", "worker", "manager", "calico", "prometheus", "resources"}
	for _, component := range allComponents {
		switch component {
		case "controller":
			CollectControllerInfo(ctx)
		case "worker":
			CollectWorkerInfo(ctx)
		case "manager":
			CollectManagerInfo(ctx)
		case "calico":
			CollectCalicoInfo(ctx)
		case "prometheus":
			CollectPrometheusInfo(ctx)
		case "resources":
			CollectKubernetesResources(ctx)
		}
	}
}

// Helper functions for collection
func collectPodLogs(ctx *CollectionContext, componentDir, labelSelector string) {
	logsDir := filepath.Join(componentDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return
	}

	// Get pods with the label selector
	pods := getPodsByLabel(ctx, labelSelector)
	for _, pod := range pods {
		logFile := filepath.Join(logsDir, fmt.Sprintf("%s.logs", pod))
		collectPodLog(ctx, pod, logFile)
	}
}

func collectPodConfigs(ctx *CollectionContext, componentDir, labelSelector string) {
	configsDir := filepath.Join(componentDir, "configs")
	if err := os.MkdirAll(configsDir, 0755); err != nil {
		return
	}

	// Get pods with the label selector
	pods := getPodsByLabel(ctx, labelSelector)
	for _, pod := range pods {
		configFile := filepath.Join(configsDir, fmt.Sprintf("%s.yaml", pod))
		collectPodConfig(ctx, pod, configFile)
	}
}

func collectPodEvents(ctx *CollectionContext, componentDir, labelSelector string) {
	eventsDir := filepath.Join(componentDir, "events")
	if err := os.MkdirAll(eventsDir, 0755); err != nil {
		return
	}

	// Get pods with the label selector
	pods := getPodsByLabel(ctx, labelSelector)
	for _, pod := range pods {
		eventFile := filepath.Join(eventsDir, fmt.Sprintf("%s.events", pod))
		collectPodEvent(ctx, pod, eventFile)
	}
}

func collectResource(ctx *CollectionContext, componentDir, resourceType, labelSelector string) {
	resourceDir := filepath.Join(componentDir, resourceType)
	if err := os.MkdirAll(resourceDir, 0755); err != nil {
		return
	}

	outputFile := filepath.Join(resourceDir, fmt.Sprintf("%s.yaml", resourceType))
	collectKubernetesResource(ctx, resourceType, labelSelector, outputFile)
}

// CreateArchive creates a tar.gz archive from the temporary directory
func CreateArchive(tempDir, outputFile string) error {
	// Create output file
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	// Create gzip writer
	gw := gzip.NewWriter(file)
	defer gw.Close()

	// Create tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Walk through the temporary directory
	return filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create header
		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}

		// Update header name to be relative to tempDir
		relPath, err := filepath.Rel(tempDir, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// If it's a file, write the content
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}

		return nil
	})
}

// getPodsByLabel gets pods by label selector using kubectl
func getPodsByLabel(ctx *CollectionContext, labelSelector string) []string {
	args := []string{"get", "pods", "-n", ctx.Namespace, "-l", labelSelector, "-o", "jsonpath={.items[*].metadata.name}"}

	if ctx.Kubeconfig != "" {
		args = append([]string{"--kubeconfig=" + ctx.Kubeconfig}, args...)
	}

	output, stderr, err := runKubectlCommand(args)
	if err != nil {
		util.Printf("[ERROR] Failed to get pods with selector %s: %v", labelSelector, err)
		if stderr != "" {
			util.Printf("[WARN] Stderr: %s", stderr)
		}
		return []string{}
	}

	podNames := strings.Fields(output)
	if len(podNames) == 0 {
		util.Printf("[WARN] No pods found with selector: %s", labelSelector)
	}

	return podNames
}

// collectPodLog collects logs for a specific pod
func collectPodLog(ctx *CollectionContext, podName, outputFile string) {
	util.Printf("[INFO] Collecting logs for pod: %s", podName)

	args := []string{"logs", podName, "-n", ctx.Namespace, "--all-containers=true"}
	if ctx.Kubeconfig != "" {
		args = append([]string{"--kubeconfig=" + ctx.Kubeconfig}, args...)
	}

	// Create output file
	file, err := os.Create(outputFile)
	if err != nil {
		util.Printf("[ERROR] Failed to create log file %s: %v", outputFile, err)
		return
	}
	defer file.Close()

	// Run kubectl logs command
	output, stderr, err := runKubectlCommand(args)
	if err != nil {
		util.Printf("[ERROR] Failed to collect logs for pod %s: %v", podName, err)
		if stderr != "" {
			util.Printf("[WARN] Stderr: %s", stderr)
		}
		file.WriteString(fmt.Sprintf("Error collecting logs: %v\n", err))
		return
	}

	// Write logs to file
	file.WriteString(output)
	if stderr != "" {
		file.WriteString(fmt.Sprintf("\n--- Errors ---\n%s", stderr))
	}
}

// collectPodConfig collects configuration for a specific pod
func collectPodConfig(ctx *CollectionContext, podName, outputFile string) {
	util.Printf("[INFO] Collecting config for pod: %s", podName)

	args := []string{"get", "pod", podName, "-n", ctx.Namespace, "-o", "yaml"}
	if ctx.Kubeconfig != "" {
		args = append([]string{"--kubeconfig=" + ctx.Kubeconfig}, args...)
	}

	// Create output file
	file, err := os.Create(outputFile)
	if err != nil {
		util.Printf("[ERROR] Failed to create config file %s: %v", outputFile, err)
		return
	}
	defer file.Close()

	// Run kubectl get command
	output, stderr, err := runKubectlCommand(args)
	if err != nil {
		util.Printf("[ERROR] Failed to collect config for pod %s: %v", podName, err)
		if stderr != "" {
			util.Printf("[WARN] Stderr: %s", stderr)
		}
		file.WriteString(fmt.Sprintf("Error collecting config: %v\n", err))
		return
	}

	// Write config to file
	file.WriteString(output)
	if stderr != "" {
		file.WriteString(fmt.Sprintf("\n--- Errors ---\n%s", stderr))
	}
}

// collectPodEvent collects events for a specific pod
func collectPodEvent(ctx *CollectionContext, podName, outputFile string) {
	util.Printf("[INFO] Collecting events for pod: %s", podName)

	args := []string{"get", "events", "-n", ctx.Namespace, "--field-selector", fmt.Sprintf("involvedObject.name=%s", podName), "-o", "yaml"}
	if ctx.Kubeconfig != "" {
		args = append([]string{"--kubeconfig=" + ctx.Kubeconfig}, args...)
	}

	// Create output file
	file, err := os.Create(outputFile)
	if err != nil {
		util.Printf("[ERROR] Failed to create event file %s: %v", outputFile, err)
		return
	}
	defer file.Close()

	// Run kubectl get events command
	output, stderr, err := runKubectlCommand(args)
	if err != nil {
		util.Printf("[ERROR] Failed to collect events for pod %s: %v", podName, err)
		if stderr != "" {
			util.Printf("[WARN] Stderr: %s", stderr)
		}
		file.WriteString(fmt.Sprintf("Error collecting events: %v\n", err))
		return
	}

	// Write events to file
	file.WriteString(output)
	if stderr != "" {
		file.WriteString(fmt.Sprintf("\n--- Errors ---\n%s", stderr))
	}
}

// collectKubernetesResource collects Kubernetes resources
func collectKubernetesResource(ctx *CollectionContext, resourceType, labelSelector, outputFile string) {
	util.Printf("[INFO] Collecting %s resources", resourceType)

	args := []string{"get", resourceType, "-n", ctx.Namespace, "-o", "yaml"}
	if labelSelector != "" {
		args = append(args, "-l", labelSelector)
	}
	if ctx.Kubeconfig != "" {
		args = append([]string{"--kubeconfig=" + ctx.Kubeconfig}, args...)
	}

	// Create output file
	file, err := os.Create(outputFile)
	if err != nil {
		util.Printf("[ERROR] Failed to create resource file %s: %v", outputFile, err)
		return
	}
	defer file.Close()

	// Run kubectl get command
	output, stderr, err := runKubectlCommand(args)
	if err != nil {
		util.Printf("[ERROR] Failed to collect %s resources: %v", resourceType, err)
		if stderr != "" {
			util.Printf("[WARN] Stderr: %s", stderr)
		}
		file.WriteString(fmt.Sprintf("Error collecting %s resources: %v\n", resourceType, err))
		return
	}

	// Write resources to file
	file.WriteString(output)
	if stderr != "" {
		file.WriteString(fmt.Sprintf("\n--- Errors ---\n%s", stderr))
	}
}
