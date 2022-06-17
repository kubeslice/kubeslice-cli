package internal

type Configuration struct {
	Executables            ExecutablePath         `json:"executables"`
	ClusterConfiguration   ClusterConfiguration   `json:"cluster_configuration"`
	KubeSliceConfiguration KubeSliceConfiguration `json:"kubeslice_configuration"`
	HelmChartConfiguration HelmChartConfiguration `json:"helm_chart_configuration"`
	ApplicationDeployment  []ApplicationDeployments `json:"application_deployment"`
}

type ApplicationDeployments struct {
	HelmDeployment    []ApplicationHelmChartConfiguration `json:"helm_deployment"`
	KubectlDeployment []ApplicationKubectlDefinition      `json:"kubectl_deployment"`
}

type ApplicationKubectlDefinition struct {
	FilePath       string   `json:"file_path"`
	Namespace      string   `json:"namespace"`
	WorkerClusters []string `json:"worker_clusters"`
}

type ApplicationHelmChartConfiguration struct {
	RepoAlias       string   `json:"repo_alias"`
	RepoUrl         string   `json:"repo_url"`
	ControllerChart string   `json:"controller_chart"`
	WorkerChart     string   `json:"worker_chart"`
	HelmUsername    string   `json:"helm_username"`
	HelmPassword    string   `json:"helm_password"`
	Flags           []string `json:"flags"`
	WorkerClusters  []string `json:"worker_clusters"`
}

type HelmChartConfiguration struct {
	RepoAlias       string `json:"repo_alias"`
	RepoUrl         string `json:"repo_url"`
	ControllerChart string `json:"controller_chart"`
	WorkerChart     string `json:"worker_chart"`
	HelmUsername    string `json:"helm_username"`
	HelmPassword    string `json:"helm_password"`
}

type KubeSliceConfiguration struct {
	ProjectName string `json:"project_name"`
}

type ClusterConfiguration struct {
	KindDemo          bool      `json:"kind_demo"`
	KubeConfigPath    string    `json:"kube_config_path"`
	ControllerCluster Cluster   `json:"controller"`
	WorkerClusters    []Cluster `json:"worker"`
}

type Cluster struct {
	Name        string `json:"name"`
	ContextName string `json:"context_name"`
}

type ExecutablePath struct {
	Kubectl string `json:"kubectl"`
	Kind    string `json:"kind"`
	Helm    string `json:"helm"`
	Docker  string `json:"docker"`
}
