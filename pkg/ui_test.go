package pkg

import (
	"testing"

	"github.com/kubeslice/kubeslice-cli/pkg/internal"
)

func TestGetUIEndpoint_WithMock(t *testing.T) {
	// Initialize CliOptions using the helper
	cliParams := CliParams{
		ObjectType:   "project",
		ObjectName:   "mock-cluster",
		Namespace:    "test-namespace",
		FileName:     "test-file.yaml",
		Config:       "",
		OutputFormat: "json",
	}
	SetCliOptions(cliParams)

	// Initialize ApplicationConfiguration and nested fields
	ApplicationConfiguration = &internal.ConfigurationSpecs{}
	ApplicationConfiguration.Configuration = internal.Configuration{}
	ApplicationConfiguration.Configuration.ClusterConfiguration = internal.ClusterConfiguration{}
	ApplicationConfiguration.Configuration.ClusterConfiguration.Profile = "mock-profile"

	called := false
	mockFunc := func(c *internal.Cluster, profile string) string {
		called = true
		if c.Name != "mock-cluster" || profile != "mock-profile" {
			t.Errorf("Unexpected values: cluster=%v, profile=%v", c.Name, profile)
		}
		return "https://mock-endpoint"
	}

	// Inject mock
	getUIEndpointFunc = mockFunc
	defer func() { getUIEndpointFunc = internal.GetUIEndpoint }() // Restore after test

	// Inject mock config
	CliOptions.Cluster = &internal.Cluster{Name: "mock-cluster"}
	ApplicationConfiguration.Configuration.ClusterConfiguration.Profile = "mock-profile"

	GetUIEndpoint()

	if !called {
		t.Error("Expected mock GetUIEndpoint to be called")
	}
}
