package internal

type ConfigurationSpecs struct {
	Configuration Configuration `yaml:"configuration"`
}

type Configuration struct {
	Executables            ExecutablePath         `yaml:"executables"`
	ClusterConfiguration   ClusterConfiguration   `yaml:"cluster_configuration"`
	KubeSliceConfiguration KubeSliceConfiguration `yaml:"kubeslice_configuration"`
	HelmChartConfiguration HelmChartConfiguration `yaml:"helm_chart_configuration"`
}

type HelmChartConfiguration struct {
	RepoAlias        string `yaml:"repo_alias"`
	RepoUrl          string `yaml:"repo_url"`
	CertManagerChart string `yaml:"cert_manager_chart"`
	ControllerChart  string `yaml:"controller_chart"`
	WorkerChart      string `yaml:"worker_chart"`
	HelmUsername     string `yaml:"helm_username"`
	HelmPassword     string `yaml:"helm_password"`
	ImagePullSecret  string `yaml:"image_pull_secret"`
}

type KubeSliceConfiguration struct {
	ProjectName string `yaml:"project_name"`
}

type ClusterConfiguration struct {
	KindDemo          bool      `yaml:"kind_demo"`
	KubeConfigPath    string    `yaml:"kube_config_path"`
	ControllerCluster Cluster   `yaml:"controller"`
	WorkerClusters    []Cluster `yaml:"workers"`
}

type Cluster struct {
	Name           string `yaml:"name"`
	ContextName    string `yaml:"context_name"`
	KubeConfigPath string `yaml:"kube_config_path"`
}

type ExecutablePath struct {
	Kubectl string `yaml:"kubectl"`
	Kind    string `yaml:"kind"`
	Helm    string `yaml:"helm"`
	Docker  string `yaml:"docker"`
}
