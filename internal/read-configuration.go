package internal

import (
	"fmt"
	"io/ioutil"

	"github.com/go-yaml/yaml"
	"github.com/kubeslice/slicectl/util"
)

const (
	ProfileFullDemo    = "full-demo"
	ProfileMinimalDemo = "minimal-demo"
)

var ApplicationConfiguration *ConfigurationSpecs

var defaultConfiguration = &ConfigurationSpecs{
	Configuration: Configuration{
		ClusterConfiguration: ClusterConfiguration{
			Profile: "full-demo",
			ControllerCluster: Cluster{
				Name: "ks-ctrl",
			},
			WorkerClusters: []Cluster{
				{
					Name: "ks-w-1",
				},
				{
					Name: "ks-w-2",
				},
			},
		},
		KubeSliceConfiguration: KubeSliceConfiguration{
			ProjectName: "demo",
		},
		HelmChartConfiguration: HelmChartConfiguration{
			RepoAlias: "kubeslice-demo",
			RepoUrl:   "https://kubeslice.github.io/charts/",
			CertManagerChart: HelmChart{
				ChartName: "cert-manager",
			},
			ControllerChart: HelmChart{
				ChartName: "kubeslice-controller",
			},
			WorkerChart: HelmChart{
				ChartName: "kubeslice-worker",
			},
		},
	},
}

func readConfiguration(fileName string) *ConfigurationSpecs {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		util.Fatalf("%s Failed to read configuration file %v", util.Cross, err)

	}
	specs := &ConfigurationSpecs{}
	err = yaml.Unmarshal(file, specs)
	if err != nil {
		util.Fatalf("%s Failed to parse configuration file %v", util.Cross, err)
	}
	return specs
}

func validateConfiguration(specs *ConfigurationSpecs) []string {
	var errors = make([]string, 0)
	if specs == nil {
		errors = append(errors, fmt.Sprintf("%s Invalid Configuration", util.Cross))
	}
	cc := &specs.Configuration.ClusterConfiguration
	ksc := &specs.Configuration.KubeSliceConfiguration
	hc := &specs.Configuration.HelmChartConfiguration
	if cc.Profile != "" {
		switch cc.Profile {
		case ProfileFullDemo:
		case ProfileMinimalDemo:
		default:
			errors = append(errors, fmt.Sprintf("%s Unknown profile: %s. Possible values %s", util.Cross, cc.Profile, []string{ProfileFullDemo,
				ProfileMinimalDemo}))
		}
		if cc.KubeConfigPath != "" || cc.ControllerCluster.KubeConfigPath != "" {
			errors = append(errors, fmt.Sprintf("%s Cannot specify configuration.cluster_configuration.kube_config_path or configuration.cluster_configuration.controller.kube_config_path when running a kind cluster demo", util.Cross))
		}
		cc.ControllerCluster.KubeConfigPath = kubeconfigPath
		if cc.ControllerCluster.ContextName != "" {
			errors = append(errors, fmt.Sprintf("%s Cannot specify configuration.cluster_configuration.controller.context_name when running a kind cluster demo", util.Cross))
		}
		cc.ControllerCluster.ContextName = "kind-" + cc.ControllerCluster.Name
		if len(cc.WorkerClusters) < 2 {
			errors = append(errors, fmt.Sprintf("%s At least 2 configuration.cluster_configuration.workers are required for kind cluster Demo", util.Cross))
		}
		for i, cluster := range cc.WorkerClusters {
			if cluster.KubeConfigPath != "" {
				errors = append(errors, fmt.Sprintf("%s Cannot specify configuration.cluster_configuration.kube_config_path or configuration.cluster_configuration.workers[%d].kube_config_path when running a kind cluster demo", util.Cross, i))
			}
			cc.WorkerClusters[i].KubeConfigPath = kubeconfigPath
			if cluster.ContextName != "" {
				errors = append(errors, fmt.Sprintf("%s Cannot specify configuration.cluster_configuration.workers[%d].context_name for worker when running a kind cluster demo", util.Cross, i))
			}
			cc.WorkerClusters[i].ContextName = "kind-" + cluster.Name
		}
	} else {
		if cc.KubeConfigPath == "" && cc.ControllerCluster.KubeConfigPath == "" {
			errors = append(errors, fmt.Sprintf("%s configuration.cluster_configuration.kube_config_path or configuration.cluster_configuration.controller.kube_config_path must be specified when setting up topology", util.Cross))
		}
		if cc.ControllerCluster.KubeConfigPath == "" && cc.KubeConfigPath != "" {
			cc.ControllerCluster.KubeConfigPath = cc.KubeConfigPath
		}
		if cc.ControllerCluster.ContextName != "" {
			errors = append(errors, fmt.Sprintf("%s configuration.cluster_configuration.controller.context_name must be specified when setting up topology", util.Cross))
		}
		for i, cluster := range cc.WorkerClusters {
			if cc.KubeConfigPath == "" && cluster.KubeConfigPath != "" {
				errors = append(errors, fmt.Sprintf("%s configuration.cluster_configuration.kube_config_path or configuration.cluster_configuration.workers[%d].kube_config_path must be specified when setting up topology", util.Cross, i))
			}
			if cluster.KubeConfigPath == "" && cc.KubeConfigPath != "" {
				cc.WorkerClusters[i].KubeConfigPath = cc.KubeConfigPath
			}
			if cluster.ContextName != "" {
				errors = append(errors, fmt.Sprintf("%s configuration.cluster_configuration.workers[%d].context_name must be specified when setting up topology", util.Cross, i))
			}
		}
	}
	if cc.ControllerCluster.Name == "" {
		errors = append(errors, fmt.Sprintf("%s configuration.cluster_configuration.controller.name must be specified", util.Cross))
	}
	for i, cluster := range cc.WorkerClusters {
		if cluster.Name == "" {
			errors = append(errors, fmt.Sprintf("%s configuration.cluster_configuration.workers[%d].name must be specified", util.Cross, i))
		}
	}
	if ksc.ProjectName == "" {
		errors = append(errors, fmt.Sprintf("%s configuration.kubeslice_configuration.project_name must be specified", util.Cross))
	}
	if hc.RepoAlias == "" {
		errors = append(errors, fmt.Sprintf("%s configuration.helm_chart_configuration.repo_alias must be specified", util.Cross))
	}
	if hc.RepoUrl == "" {
		errors = append(errors, fmt.Sprintf("%s configuration.helm_chart_configuration.repo_url must be specified", util.Cross))
	}
	if hc.CertManagerChart.ChartName == "" {
		errors = append(errors, fmt.Sprintf("%s configuration.helm_chart_configuration.cert_manager_chart must be specified", util.Cross))
	}
	if hc.ControllerChart.ChartName == "" {
		errors = append(errors, fmt.Sprintf("%s configuration.helm_chart_configuration.controller_chart must be specified", util.Cross))
	}
	if hc.WorkerChart.ChartName == "" {
		errors = append(errors, fmt.Sprintf("%s configuration.helm_chart_configuration.worker_chart must be specified", util.Cross))
	}
	return errors
}

func ReadAndValidateConfiguration(fileName string) *ConfigurationSpecs {
	var specs *ConfigurationSpecs
	if fileName != "" {
		specs = readConfiguration(fileName)
	} else {
		specs = defaultConfiguration
	}
	errors := validateConfiguration(specs)
	if len(errors) > 0 {
		for _, s := range errors {
			util.Printf(s)
		}
		util.Fatalf("%s Process failed due to invalid configuration", util.Cross)
	}
	ApplicationConfiguration = specs
	return specs
}
