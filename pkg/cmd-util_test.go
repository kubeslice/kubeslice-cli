package pkg

import (
	"testing"

	"github.com/kubeslice/kubeslice-cli/pkg/internal"
)

func TestSetCliOptions(t *testing.T) {
	cliParams := CliParams{
		ObjectType:   "project",
		ObjectName:   "test-project",
		Namespace:    "test-namespace",
		FileName:     "test-file.yaml",
		Config:       "",
		OutputFormat: "json",
	}

	SetCliOptions(cliParams)

	if CliOptions.ObjectType != "project" {
		t.Errorf("Expected ObjectType to be 'project', got %s", CliOptions.ObjectType)
	}
	if CliOptions.ObjectName != "test-project" {
		t.Errorf("Expected ObjectName to be 'test-project', got %s", CliOptions.ObjectName)
	}
	if CliOptions.Namespace != "test-namespace" {
		t.Errorf("Expected Namespace to be 'test-namespace', got %s", CliOptions.Namespace)
	}
}

func TestReadAndValidateConfiguration_WithDefaults(t *testing.T) {
	if defaultConfiguration == nil {
		t.Fatal("Expected defaultConfiguration to not be nil")
	}
	if defaultConfiguration.Configuration.ClusterConfiguration.Profile != "full-demo" {
		t.Errorf("Expected profile to be 'full-demo', got %s", defaultConfiguration.Configuration.ClusterConfiguration.Profile)
	}
	if defaultConfiguration.Configuration.KubeSliceConfiguration.ProjectName != "demo" {
		t.Errorf("Expected project name to be 'demo', got %s", defaultConfiguration.Configuration.KubeSliceConfiguration.ProjectName)
	}
}

func TestReadAndValidateConfiguration_WithEntProfile(t *testing.T) {
	if defaultEntConfiguration == nil {
		t.Fatal("Expected defaultEntConfiguration to not be nil")
	}
	if defaultEntConfiguration.RepoAlias != "kubeslice-ent-demo" {
		t.Errorf("Expected repo alias to be 'kubeslice-ent-demo', got %s", defaultEntConfiguration.RepoAlias)
	}
	if defaultEntConfiguration.UIChart.ChartName != "kubeslice-ui" {
		t.Errorf("Expected UI chart name to be 'kubeslice-ui', got %s", defaultEntConfiguration.UIChart.ChartName)
	}
}

func TestValidateConfiguration_ValidConfig(t *testing.T) {
	specs := &internal.ConfigurationSpecs{
		Configuration: internal.Configuration{
			ClusterConfiguration: internal.ClusterConfiguration{
				ControllerCluster: internal.Cluster{
					Name:           "controller",
					KubeConfigPath: "/path/to/kubeconfig",
					ContextName:    "controller-context",
				},
				WorkerClusters: []internal.Cluster{
					{
						Name:           "worker1",
						KubeConfigPath: "/path/to/kubeconfig",
						ContextName:    "worker1-context",
					},
					{
						Name:           "worker2",
						KubeConfigPath: "/path/to/kubeconfig",
						ContextName:    "worker2-context",
					},
				},
			},
			KubeSliceConfiguration: internal.KubeSliceConfiguration{
				ProjectName: "test-project",
			},
			HelmChartConfiguration: internal.HelmChartConfiguration{
				RepoAlias: "test-repo",
				RepoUrl:   "https://test.com",
				CertManagerChart: internal.HelmChart{
					ChartName: "cert-manager",
				},
				ControllerChart: internal.HelmChart{
					ChartName: "controller",
				},
				WorkerChart: internal.HelmChart{
					ChartName: "worker",
				},
			},
		},
	}

	errors := validateConfiguration(specs)
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors, got %d errors: %v", len(errors), errors)
	}
}

func TestValidateConfiguration_InvalidProfile(t *testing.T) {
	specs := &internal.ConfigurationSpecs{
		Configuration: internal.Configuration{
			ClusterConfiguration: internal.ClusterConfiguration{
				Profile: "invalid-profile",
			},
		},
	}

	errors := validateConfiguration(specs)
	if len(errors) == 0 {
		t.Error("Expected validation errors for invalid profile, got none")
	}
}

func TestValidateConfiguration_MissingControllerName(t *testing.T) {
	specs := &internal.ConfigurationSpecs{
		Configuration: internal.Configuration{
			ClusterConfiguration: internal.ClusterConfiguration{
				ControllerCluster: internal.Cluster{
					Name: "", // Missing name
				},
			},
			KubeSliceConfiguration: internal.KubeSliceConfiguration{
				ProjectName: "test",
			},
			HelmChartConfiguration: internal.HelmChartConfiguration{
				RepoAlias:        "test",
				RepoUrl:          "https://test.com",
				CertManagerChart: internal.HelmChart{ChartName: "cert"},
				ControllerChart:  internal.HelmChart{ChartName: "ctrl"},
				WorkerChart:      internal.HelmChart{ChartName: "worker"},
			},
		},
	}

	errors := validateConfiguration(specs)
	if len(errors) == 0 {
		t.Error("Expected validation error for missing controller name")
	}
}
