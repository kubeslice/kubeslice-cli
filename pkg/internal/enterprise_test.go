package internal

import (
	"errors"
	"io"
	"testing"

	"github.com/kubeslice/kubeslice-cli/util"
)

// Mock getNodeIP to return dummy IP
func mockGetNodeIP(_ *Cluster) (string, error) {
	return "192.168.1.100", nil
}

func TestGetUIEndpoint_NodePort(t *testing.T) {
	runCalled := false

	// Mock kubectl output for NodePort
	nodePortJSON := `{
		"type": "NodePort",
		"ports": [{
			"name": "http",
			"nodePort": 30080
		}]
	}`

	runCommandCustomIO = func(name string, stdout, stderr io.Writer, _ bool, args ...string) error {
		runCalled = true
		stdout.Write([]byte("'" + nodePortJSON + "'")) // wrap in quotes to simulate jsonpath output
		return nil
	}
	getNodeIPFunc = mockGetNodeIP
	defer func() {
		runCommandCustomIO = util.RunCommandCustomIO
		getNodeIPFunc = getNodeIP
	}()

	cluster := &Cluster{
		ContextName:    "mock-context",
		KubeConfigPath: "/fake/config",
	}
	endpoint := GetUIEndpoint(cluster, "some-profile")

	expected := "https://192.168.1.100:30080"
	if endpoint != expected {
		t.Errorf("Expected endpoint %q, got %q", expected, endpoint)
	}
	if !runCalled {
		t.Error("Expected RunCommandCustomIO to be called")
	}
}

// Mocks a LoadBalancer service and checks the endpoint.
func TestGetUIEndpoint_LoadBalancer(t *testing.T) {
	runCalled := false

	// Mock output for LoadBalancer
	loadBalancerJSON := `{
		"type": "LoadBalancer",
		"externalIPs": ["1.2.3.4"],
		"ports": [{
			"name": "http",
			"port": 443
		}]
	}`

	runCommandCustomIO = func(name string, stdout, stderr io.Writer, _ bool, args ...string) error {
		runCalled = true
		stdout.Write([]byte("'" + loadBalancerJSON + "'"))
		return nil
	}
	defer func() {
		runCommandCustomIO = util.RunCommandCustomIO
	}()

	cluster := &Cluster{
		ContextName:    "mock-context",
		KubeConfigPath: "/fake/config",
	}
	endpoint := GetUIEndpoint(cluster, "some-profile")

	expected := "https://1.2.3.4:443"
	if endpoint != expected {
		t.Errorf("Expected endpoint %q, got %q", expected, endpoint)
	}
	if !runCalled {
		t.Error("Expected RunCommandCustomIO to be called")
	}
}

// Mocks invalid JSON output and checks that the function returns an empty string
func TestGetUIEndpoint_InvalidJSON(t *testing.T) {
	runCommandCustomIO = func(name string, stdout, stderr io.Writer, _ bool, args ...string) error {
		stdout.Write([]byte("'not-a-json'"))
		return nil
	}
	defer func() {
		runCommandCustomIO = util.RunCommandCustomIO
	}()

	cluster := &Cluster{
		ContextName:    "mock-context",
		KubeConfigPath: "/fake/config",
	}
	endpoint := GetUIEndpoint(cluster, "profile")
	if endpoint != "" {
		t.Errorf("Expected empty endpoint on invalid JSON, got %q", endpoint)
	}
}

// Mocks a command failure and checks that the function returns an empty string
func TestGetUIEndpoint_CommandFailure(t *testing.T) {
	runCommandCustomIO = func(name string, stdout, stderr io.Writer, _ bool, args ...string) error {
		return errors.New("kubectl failed")
	}
	defer func() {
		runCommandCustomIO = util.RunCommandCustomIO
	}()

	cluster := &Cluster{
		ContextName:    "mock-context",
		KubeConfigPath: "/fake/config",
	}
	endpoint := GetUIEndpoint(cluster, "profile")
	if endpoint != "" {
		t.Errorf("Expected empty endpoint on command failure, got %q", endpoint)
	}
}
