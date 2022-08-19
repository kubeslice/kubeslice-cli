package pkg

import (
	"fmt"
	"io/ioutil"

	"github.com/go-yaml/yaml"
	"github.com/kubeslice/slicectl/pkg/internal"
	"github.com/kubeslice/slicectl/util"
)

const (
	ProfileFullDemo    = "full-demo"
	ProfileMinimalDemo = "minimal-demo"
)

type CliParams struct {
	ObjectType string   // "project", "cluster", "sliceConfig"
	ObjectName string   // "projectName", "clusterName", "sliceConfigName"
	Namespace  string   // namespace for the workloads
	FileName   string   // path to the resource description file
	Config     string   // cluster
	Worker     []string //workerList
}

var ApplicationConfiguration *internal.ConfigurationSpecs

var CliOptions *internal.CliOptionsStruct

func SetCliOptions(cliParams CliParams) {
	var controllerCluster *internal.Cluster
	configSpecs := ReadAndValidateConfiguration(cliParams.Config)
	if cliParams.Config != "" {
		controllerCluster = &configSpecs.Configuration.ClusterConfiguration.ControllerCluster
	}
	options := &internal.CliOptionsStruct{
		Namespace:  cliParams.Namespace,
		ObjectName: cliParams.ObjectName,
		ObjectType: cliParams.ObjectType,
		FileName:   cliParams.FileName,
		Cluster:    controllerCluster,
	}
	CliOptions = options
	util.ExecutablePaths = map[string]string{
		"kubectl": "kubectl",
	}
}

var defaultConfiguration = &internal.ConfigurationSpecs{
	Configuration: internal.Configuration{
		ClusterConfiguration: internal.ClusterConfiguration{
			Profile: "full-demo",
			ControllerCluster: internal.Cluster{
				Name: "ks-ctrl",
			},
			WorkerClusters: []internal.Cluster{
				{
					Name: "ks-w-1",
				},
				{
					Name: "ks-w-2",
				},
			},
		},
		KubeSliceConfiguration: internal.KubeSliceConfiguration{
			ProjectName: "demo",
		},
		HelmChartConfiguration: internal.HelmChartConfiguration{
			RepoAlias: "kubeslice-demo",
			RepoUrl:   "https://kubeslice.github.io/charts/",
			CertManagerChart: internal.HelmChart{
				ChartName: "cert-manager",
			},
			ControllerChart: internal.HelmChart{
				ChartName: "kubeslice-controller",
			},
			WorkerChart: internal.HelmChart{
				ChartName: "kubeslice-worker",
			},
		},
	},
}

func readConfiguration(fileName string) *internal.ConfigurationSpecs {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		util.Fatalf("%s Failed to read configuration file %v", util.Cross, err)

	}
	specs := &internal.ConfigurationSpecs{}
	err = yaml.Unmarshal(file, specs)
	if err != nil {
		util.Fatalf("%s Failed to parse configuration file %v", util.Cross, err)
	}
	return specs
}

func validateConfiguration(specs *internal.ConfigurationSpecs) []string {
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
		cc.ControllerCluster.KubeConfigPath = internal.KubeconfigPath
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
			cc.WorkerClusters[i].KubeConfigPath = internal.KubeconfigPath
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
		if cc.ControllerCluster.ContextName == "" {
			errors = append(errors, fmt.Sprintf("%s configuration.cluster_configuration.controller.context_name must be specified when setting up topology", util.Cross))
		}
		for i, cluster := range cc.WorkerClusters {
			if cc.KubeConfigPath == "" && cluster.KubeConfigPath == "" {
				errors = append(errors, fmt.Sprintf("%s configuration.cluster_configuration.kube_config_path or configuration.cluster_configuration.workers[%d].kube_config_path must be specified when setting up topology", util.Cross, i))
			}
			if cluster.KubeConfigPath == "" && cc.KubeConfigPath != "" {
				cc.WorkerClusters[i].KubeConfigPath = cc.KubeConfigPath
			}
			if cluster.ContextName == "" {
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

func ReadAndValidateConfiguration(fileName string) *internal.ConfigurationSpecs {
	var specs *internal.ConfigurationSpecs
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
