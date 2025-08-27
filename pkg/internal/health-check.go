package internal

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const (
	HealthStatusHealthy   = "Healthy"
	HealthStatusUnhealthy = "Unhealthy"
	HealthStatusUnknown   = "Unknown"
)

type ClusterHealth struct {
	ClusterName    string             `json:"cluster_name"`
	ClusterType    string             `json:"cluster_type"`
	OverallStatus  string             `json:"overall_status"`
	NodeStatus     NodeHealthStatus   `json:"node_status"`
	PodStatus      PodHealthStatus    `json:"pod_status"`
	ComponentStatus ComponentHealthStatus `json:"component_status"`
	Timestamp      string             `json:"timestamp"`
}

type NodeHealthStatus struct {
	Status      string `json:"status"`
	ReadyNodes  int    `json:"ready_nodes"`
	TotalNodes  int    `json:"total_nodes"`
	Details     string `json:"details"`
}

type PodHealthStatus struct {
	Status          string `json:"status"`
	RunningPods     int    `json:"running_pods"`
	TotalPods       int    `json:"total_pods"`
	NamespaceStatus map[string]string `json:"namespace_status"`
	Details         string `json:"details"`
}

type ComponentHealthStatus struct {
	Status     string            `json:"status"`
	Components map[string]string `json:"components"`
	Details    string            `json:"details"`
}

func ShowClusterHealth(clusterName string, config *ConfigurationSpecs, options *CliOptionsStruct) {
	var cluster *Cluster
	var clusterType string

	// Find the cluster in configuration
	if config.Configuration.ClusterConfiguration.ControllerCluster.Name == clusterName {
		cluster = &config.Configuration.ClusterConfiguration.ControllerCluster
		clusterType = "controller"
	} else {
		for _, worker := range config.Configuration.ClusterConfiguration.WorkerClusters {
			if worker.Name == clusterName {
				cluster = &worker
				clusterType = "worker"
				break
			}
		}
	}

	if cluster == nil {
		util.Fatalf("Cluster '%s' not found in configuration", clusterName)
	}

	util.Printf("\n%s Checking health for cluster: %s (%s)", util.Info, clusterName, clusterType)
	
	health := checkClusterHealth(*cluster, clusterType)
	displayClusterHealth(health, options.OutputFormat)
}

func ShowAllClustersHealth(config *ConfigurationSpecs, options *CliOptionsStruct) {
	util.Printf("\n%s Checking health for all clusters", util.Info)
	
	var allHealth []ClusterHealth

	// Check controller cluster
	controllerHealth := checkClusterHealth(config.Configuration.ClusterConfiguration.ControllerCluster, "controller")
	allHealth = append(allHealth, controllerHealth)

	// Check worker clusters
	for _, worker := range config.Configuration.ClusterConfiguration.WorkerClusters {
		workerHealth := checkClusterHealth(worker, "worker")
		allHealth = append(allHealth, workerHealth)
	}

	displayAllClustersHealth(allHealth, options.OutputFormat)
}

func checkClusterHealth(cluster Cluster, clusterType string) ClusterHealth {
	health := ClusterHealth{
		ClusterName: cluster.Name,
		ClusterType: clusterType,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	// Check node status
	health.NodeStatus = checkNodeHealth(cluster)
	
	// Check pod status
	health.PodStatus = checkPodHealth(cluster)
	
	// Check component status
	health.ComponentStatus = checkComponentHealth(cluster, clusterType)

	// Determine overall status
	health.OverallStatus = determineOverallStatus(health.NodeStatus.Status, health.PodStatus.Status, health.ComponentStatus.Status)

	return health
}

func checkNodeHealth(cluster Cluster) NodeHealthStatus {
	var outB, errB bytes.Buffer
	status := NodeHealthStatus{
		Status: HealthStatusUnknown,
	}

	// Get node status
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, true, 
		"--context="+cluster.ContextName, 
		"--kubeconfig="+cluster.KubeConfigPath, 
		"get", "nodes", 
		"-o", "jsonpath={.items[*].status.conditions[?(@.type==\"Ready\")].status}")
	
	if err != nil {
		status.Details = fmt.Sprintf("Failed to get node status: %v", err)
		status.Status = HealthStatusUnhealthy
		return status
	}

	readyStatuses := strings.Fields(outB.String())
	status.TotalNodes = len(readyStatuses)
	
	for _, readyStatus := range readyStatuses {
		if readyStatus == "True" {
			status.ReadyNodes++
		}
	}

	if status.ReadyNodes == status.TotalNodes && status.TotalNodes > 0 {
		status.Status = HealthStatusHealthy
		status.Details = fmt.Sprintf("All %d nodes are ready", status.TotalNodes)
	} else if status.ReadyNodes > 0 {
		status.Status = HealthStatusUnhealthy
		status.Details = fmt.Sprintf("%d out of %d nodes are ready", status.ReadyNodes, status.TotalNodes)
	} else {
		status.Status = HealthStatusUnhealthy
		status.Details = "No nodes are ready"
	}

	return status
}

func checkPodHealth(cluster Cluster) PodHealthStatus {
	status := PodHealthStatus{
		Status:          HealthStatusUnknown,
		NamespaceStatus: make(map[string]string),
	}

	// Check pods in kubeslice-related namespaces
	namespaces := []string{"kubeslice-controller", "kubeslice-system"}
	
	totalRunning := 0
	totalPods := 0
	allHealthy := true

	for _, namespace := range namespaces {
		nsStatus := checkNamespacePods(cluster, namespace)
		status.NamespaceStatus[namespace] = nsStatus.Status
		
		if nsStatus.Status != HealthStatusHealthy {
			allHealthy = false
		}
		
		totalRunning += nsStatus.RunningPods
		totalPods += nsStatus.TotalPods
	}

	status.RunningPods = totalRunning
	status.TotalPods = totalPods

	if allHealthy && totalPods > 0 {
		status.Status = HealthStatusHealthy
		status.Details = fmt.Sprintf("All %d pods are running across kubeslice namespaces", totalRunning)
	} else if totalRunning > 0 {
		status.Status = HealthStatusUnhealthy
		status.Details = fmt.Sprintf("%d out of %d pods are running", totalRunning, totalPods)
	} else {
		status.Status = HealthStatusUnhealthy
		status.Details = "No pods are running in kubeslice namespaces"
	}

	return status
}

func checkNamespacePods(cluster Cluster, namespace string) PodHealthStatus {
	var outB, errB bytes.Buffer
	status := PodHealthStatus{
		Status: HealthStatusUnknown,
	}

	// Check if namespace exists first
	err := util.RunCommandCustomIO("kubectl", &outB, &errB, true,
		"--context="+cluster.ContextName,
		"--kubeconfig="+cluster.KubeConfigPath,
		"get", "namespace", namespace)
	
	if err != nil {
		// Namespace doesn't exist, which might be normal for some clusters
		status.Status = HealthStatusHealthy
		status.Details = fmt.Sprintf("Namespace %s does not exist", namespace)
		return status
	}

	// Get pod status in the namespace
	outB.Reset()
	errB.Reset()
	err = util.RunCommandCustomIO("kubectl", &outB, &errB, true,
		"--context="+cluster.ContextName,
		"--kubeconfig="+cluster.KubeConfigPath,
		"get", "pods", "-n", namespace,
		"-o", "jsonpath={.items[*].status.phase}")
	
	if err != nil {
		status.Details = fmt.Sprintf("Failed to get pod status in namespace %s: %v", namespace, err)
		status.Status = HealthStatusUnhealthy
		return status
	}

	if outB.String() == "" {
		// No pods in namespace
		status.Status = HealthStatusHealthy
		status.Details = fmt.Sprintf("No pods in namespace %s", namespace)
		return status
	}

	phases := strings.Fields(outB.String())
	status.TotalPods = len(phases)
	
	for _, phase := range phases {
		if phase == "Running" {
			status.RunningPods++
		}
	}

	if status.RunningPods == status.TotalPods {
		status.Status = HealthStatusHealthy
	} else {
		status.Status = HealthStatusUnhealthy
	}

	return status
}

func checkComponentHealth(cluster Cluster, clusterType string) ComponentHealthStatus {
	status := ComponentHealthStatus{
		Status:     HealthStatusUnknown,
		Components: make(map[string]string),
	}

	var components []string
	if clusterType == "controller" {
		components = []string{"kubeslice-controller", "cert-manager"}
	} else {
		components = []string{"kubeslice-operator", "spire-server", "spire-agent"}
	}

	allHealthy := true
	for _, component := range components {
		componentStatus := checkDeploymentStatus(cluster, component)
		status.Components[component] = componentStatus
		if componentStatus != HealthStatusHealthy {
			allHealthy = false
		}
	}

	if allHealthy {
		status.Status = HealthStatusHealthy
		status.Details = "All key components are healthy"
	} else {
		status.Status = HealthStatusUnhealthy
		status.Details = "Some components are not healthy"
	}

	return status
}

func checkDeploymentStatus(cluster Cluster, deploymentName string) string {
	var outB, errB bytes.Buffer
	
	// Try different namespaces where the deployment might exist
	namespaces := []string{"kubeslice-controller", "kubeslice-system", "cert-manager"}
	
	for _, namespace := range namespaces {
		err := util.RunCommandCustomIO("kubectl", &outB, &errB, true,
			"--context="+cluster.ContextName,
			"--kubeconfig="+cluster.KubeConfigPath,
			"get", "deployment", deploymentName, "-n", namespace,
			"-o", "jsonpath={.status.readyReplicas}/{.status.replicas}")
		
		if err == nil && outB.String() != "" {
			replicas := outB.String()
			if strings.Contains(replicas, "/") {
				parts := strings.Split(replicas, "/")
				if len(parts) == 2 && parts[0] == parts[1] && parts[0] != "0" {
					return HealthStatusHealthy
				}
			}
			return HealthStatusUnhealthy
		}
		outB.Reset()
		errB.Reset()
	}
	
	return HealthStatusUnknown
}

func determineOverallStatus(nodeStatus, podStatus, componentStatus string) string {
	if nodeStatus == HealthStatusHealthy && podStatus == HealthStatusHealthy && componentStatus == HealthStatusHealthy {
		return HealthStatusHealthy
	} else if nodeStatus == HealthStatusUnhealthy || podStatus == HealthStatusUnhealthy || componentStatus == HealthStatusUnhealthy {
		return HealthStatusUnhealthy
	}
	return HealthStatusUnknown
}

func displayClusterHealth(health ClusterHealth, outputFormat string) {
	if outputFormat == "json" {
		util.PrintJSON(health)
		return
	}
	if outputFormat == "yaml" {
		util.PrintYAML(health)
		return
	}

	// Default table format
	util.Printf("\n=== Cluster Health Report ===")
	util.Printf("Cluster: %s (%s)", health.ClusterName, health.ClusterType)
	util.Printf("Overall Status: %s", getStatusWithIcon(health.OverallStatus))
	util.Printf("Timestamp: %s", health.Timestamp)
	util.Printf("")

	util.Printf("Node Status: %s", getStatusWithIcon(health.NodeStatus.Status))
	util.Printf("  %s", health.NodeStatus.Details)
	util.Printf("")

	util.Printf("Pod Status: %s", getStatusWithIcon(health.PodStatus.Status))
	util.Printf("  %s", health.PodStatus.Details)
	for ns, status := range health.PodStatus.NamespaceStatus {
		util.Printf("  - %s: %s", ns, getStatusWithIcon(status))
	}
	util.Printf("")

	util.Printf("Component Status: %s", getStatusWithIcon(health.ComponentStatus.Status))
	util.Printf("  %s", health.ComponentStatus.Details)
	for component, status := range health.ComponentStatus.Components {
		util.Printf("  - %s: %s", component, getStatusWithIcon(status))
	}
}

func displayAllClustersHealth(allHealth []ClusterHealth, outputFormat string) {
	if outputFormat == "json" {
		util.PrintJSON(allHealth)
		return
	}
	if outputFormat == "yaml" {
		util.PrintYAML(allHealth)
		return
	}

	// Default table format
	util.Printf("\n=== All Clusters Health Report ===")
	util.Printf("%-20s %-12s %-12s %-12s %-12s %-12s", "CLUSTER", "TYPE", "OVERALL", "NODES", "PODS", "COMPONENTS")
	util.Printf("%-20s %-12s %-12s %-12s %-12s %-12s", 
		strings.Repeat("-", 20), 
		strings.Repeat("-", 12), 
		strings.Repeat("-", 12), 
		strings.Repeat("-", 12), 
		strings.Repeat("-", 12), 
		strings.Repeat("-", 12))

	for _, health := range allHealth {
		util.Printf("%-20s %-12s %-12s %-12s %-12s %-12s",
			health.ClusterName,
			health.ClusterType,
			getStatusForTable(health.OverallStatus),
			getStatusForTable(health.NodeStatus.Status),
			getStatusForTable(health.PodStatus.Status),
			getStatusForTable(health.ComponentStatus.Status))
	}
	util.Printf("")
}

func getStatusWithIcon(status string) string {
	switch status {
	case HealthStatusHealthy:
		return fmt.Sprintf("%s %s", util.Check, status)
	case HealthStatusUnhealthy:
		return fmt.Sprintf("%s %s", util.Cross, status)
	default:
		return fmt.Sprintf("%s %s", util.Wait, status)
	}
}

func getStatusForTable(status string) string {
	switch status {
	case HealthStatusHealthy:
		return "✓ Healthy"
	case HealthStatusUnhealthy:
		return "✗ Unhealthy"
	default:
		return "? Unknown"
	}
}
