//go:build integration
// +build integration

package internal

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/require"
)

// isBinaryAvailable is the helper to check if a binary exists for skipping tests.
func isBinaryAvailable(t *testing.T, name string) bool {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Logf("Skipping: %s binary not found in PATH", name)
		return false
	}
	if util.ExecutablePaths == nil {
		util.ExecutablePaths = make(map[string]string)
	}
	util.ExecutablePaths[name] = path
	return true
}

// getKindClusterList shells out to the 'kind' binary to get the current list of clusters.
func getKindClusterList(t *testing.T) []string {
	t.Helper()
	if !isBinaryAvailable(t, "kind") {
		return []string{}
	}
	cmd := exec.Command(util.ExecutablePaths["kind"], "get", "clusters")
	var outB bytes.Buffer
	cmd.Stdout = &outB
	cmd.Stderr = &outB
	err := cmd.Run()
	if err != nil {
		t.Logf("Warning: kind get clusters returned error: %v", err)
	}

	clusters := []string{}
	for _, line := range strings.Split(outB.String(), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			clusters = append(clusters, line)
		}
	}
	return clusters
}

// createTestKindClusters is a setup helper to create clusters for deletion tests.
func createTestKindClusters(t *testing.T, clusterNames []string) {
	t.Helper()
	if !isBinaryAvailable(t, "kind") {
		t.Skip("kind binary not found, skipping cluster creation")
		return
	}
	for _, name := range clusterNames {
		cmd := exec.Command(util.ExecutablePaths["kind"], "create", "cluster", "--name", name)
		if err := cmd.Run(); err != nil {
			cleanupKindClusters(t, clusterNames) // Attempt cleanup
			t.Fatalf("Failed to create test cluster %s: %v", name, err)
		}
		t.Logf("Created test cluster: %s", name)
	}
}

// cleanupKindClusters is a teardown helper to delete clusters.
func cleanupKindClusters(t *testing.T, clusterNames []string) {
	t.Helper()
	if !isBinaryAvailable(t, "kind") {
		return
	}
	existingClusters := getKindClusterList(t)
	clustersToDelete := []string{}

	for _, name := range clusterNames {
		for _, existing := range existingClusters {
			if existing == name {
				clustersToDelete = append(clustersToDelete, name)
				break
			}
		}
	}

	if len(clustersToDelete) == 0 {
		return
	}

	args := append([]string{"delete", "clusters"}, clustersToDelete...)
	cmd := exec.Command(util.ExecutablePaths["kind"], args...)
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: Failed to cleanup: %v", err)
	} else {
		t.Logf("Successfully cleaned up: %v", clustersToDelete)
	}
}

// getHelmRepoList retrieves the current list of helm repositories.
func getHelmRepoList(t *testing.T) string {
	t.Helper()
	if !isBinaryAvailable(t, "helm") {
		return ""
	}
	cmd := exec.Command(util.ExecutablePaths["helm"], "repo", "list")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if strings.Contains(stderr.String(), "no repositories") {
			return ""
		}
		t.Logf("helm repo list error: %v, stderr: %s", err, stderr.String())
		return ""
	}
	return stdout.String()
}

// cleanupHelmRepo removes a helm repository if it exists.
func cleanupHelmRepo(t *testing.T, repoAlias string) {
	t.Helper()
	if !isBinaryAvailable(t, "helm") {
		return
	}
	output := getHelmRepoList(t)
	if !strings.Contains(output, repoAlias) {
		return
	}

	cmd := exec.Command(util.ExecutablePaths["helm"], "repo", "remove", repoAlias)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Logf("Failed to cleanup helm repo %s: %v", repoAlias, err)
	}
}

// setupKubeconfigForCluster configures kubectl to talk to the Kind cluster
func setupKubeconfigForCluster(t *testing.T, clusterName string) {
	t.Helper()

	util.CreateDirectoryPath(kubesliceDirectory)
	CreateKubeConfig()
	SetKubeConfigPath()

	if !isBinaryAvailable(t, "kind") {
		t.FailNow()
	}

	cmd := exec.Command(util.ExecutablePaths["kind"], "get", "kubeconfig", "--name", clusterName)
	kubeconfigBytes, err := cmd.Output()
	require.NoError(t, err, "Failed to get kubeconfig from kind")

	err = os.WriteFile(KubeconfigPath, kubeconfigBytes, 0644)
	require.NoError(t, err, "Failed to write real kubeconfig to test file")
}
