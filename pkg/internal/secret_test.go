package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Helper to create a fake kubectl binary for testing
func createFakeKubectl(t *testing.T, output string, fail bool) string {
	dir := t.TempDir()
	kubectlPath := filepath.Join(dir, "kubectl")
	script := "#!/bin/sh\n"
	if fail {
		script += "exit 1\n"
	} else {
		script += "echo \"" + output + "\"\n"
	}
	if err := os.WriteFile(kubectlPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write fake kubectl: %v", err)
	}
	return kubectlPath
}

func TestGetSecretName_Found(t *testing.T) {
	kubectl := createFakeKubectl(t, "worker-foo-secret some-data\nother-secret data", false)

	origKubectlPath := os.Getenv("KUBECTL_PATH")
	os.Setenv("KUBECTL_PATH", kubectl)
	defer os.Setenv("KUBECTL_PATH", origKubectlPath)

	origPath := os.Getenv("PATH")
	tempDir := filepath.Dir(kubectl)
	os.Setenv("PATH", tempDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	t.Logf("Testing command pipeline manually...")
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s get secret -n default | grep worker-foo | awk '{print $1}'", kubectl))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Manual command failed: %v, output: %s", err, string(output))
	} else {
		t.Logf("Manual command succeeded: %s", string(output))
	}

	t.Logf("Running GetSecretName...")
	name := GetSecretName("foo", "default", nil)
	t.Logf("GetSecretName returned: %q", name)

	if name != "worker-foo-secret" {
		t.Errorf("expected 'worker-foo-secret', got %q", name)
	}
}

func TestGetSecretName_NotFound(t *testing.T) {
	kubectl := createFakeKubectl(t, "other-secret data", false)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", filepath.Dir(kubectl)+":"+origPath)
	defer os.Setenv("PATH", origPath)

	orig := "/home/excellarate/.local/bin/kubectl"
	if _, err := os.Stat(orig); err == nil {
		os.Remove(orig)
	}
	os.Symlink(kubectl, orig)
	defer os.Remove(orig)

	name := GetSecretName("foo", "default", nil)
	if name != "" {
		t.Errorf("expected '', got %q", name)
	}
}

func TestGetSecretName_KubectlFails(t *testing.T) {
	kubectl := createFakeKubectl(t, "", true)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", filepath.Dir(kubectl)+":"+origPath)
	defer os.Setenv("PATH", origPath)

	orig := "/home/excellarate/.local/bin/kubectl"
	if _, err := os.Stat(orig); err == nil {
		os.Remove(orig)
	}
	os.Symlink(kubectl, orig)
	defer os.Remove(orig)

	name := GetSecretName("foo", "default", nil)
	if name != "" {
		t.Errorf("expected '', got %q", name)
	}
}

func TestGetSecrets_CallsKubectl(t *testing.T) {
	// This test just checks that the function runs without panic
	kubectl := createFakeKubectl(t, "worker-foo-secret some-data", false)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", filepath.Dir(kubectl)+":"+origPath)
	defer os.Setenv("PATH", origPath)

	orig := "/home/excellarate/.local/bin/kubectl"
	if _, err := os.Stat(orig); err == nil {
		os.Remove(orig)
	}
	os.Symlink(kubectl, orig)
	defer os.Remove(orig)
	origFunc := GetKubectlResources
	GetKubectlResources = func(a, b, c string, d *Cluster, e string) {}
	defer func() { GetKubectlResources = origFunc }()

	GetSecrets("foo", "default", nil, "yaml")
}

func TestGetSecrets_Sleep(t *testing.T) {
	start := time.Now()
	kubectl := createFakeKubectl(t, "worker-foo-secret some-data", false)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", filepath.Dir(kubectl)+":"+origPath)
	defer os.Setenv("PATH", origPath)

	orig := "/home/excellarate/.local/bin/kubectl"
	if _, err := os.Stat(orig); err == nil {
		os.Remove(orig)
	}
	os.Symlink(kubectl, orig)
	defer os.Remove(orig)

	origFunc := GetKubectlResources
	GetKubectlResources = func(a, b, c string, d *Cluster, e string) {}
	defer func() { GetKubectlResources = origFunc }()

	GetSecrets("foo", "default", nil, "yaml")
	if time.Since(start) < 200*time.Millisecond {
		t.Errorf("expected at least 200ms sleep")
	}
}
