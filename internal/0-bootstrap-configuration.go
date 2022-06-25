package internal

type ConfigurationSpecs struct {
	Configuration Configuration `yaml:"configuration"`
}

type Configuration struct {
	ClusterConfiguration   ClusterConfiguration   `yaml:"cluster_configuration"`
	KubeSliceConfiguration KubeSliceConfiguration `yaml:"kubeslice_configuration"`
	HelmChartConfiguration HelmChartConfiguration `yaml:"helm_chart_configuration"`
}

type HelmChartConfiguration struct {
	RepoAlias        string    `yaml:"repo_alias"`
	RepoUrl          string    `yaml:"repo_url"`
	CertManagerChart HelmChart `yaml:"cert_manager_chart"`
	ControllerChart  HelmChart `yaml:"controller_chart"`
	WorkerChart      HelmChart `yaml:"worker_chart"`
	HelmUsername     string    `yaml:"helm_username"`
	HelmPassword     string    `yaml:"helm_password"`
	ImagePullSecret  string    `yaml:"image_pull_secret"`
}

type HelmChart struct {
	ChartName string `yaml:"chart_name"`
	Version   string `yaml:"version"`
}

type KubeSliceConfiguration struct {
	ProjectName string `yaml:"project_name"`
}

type ClusterConfiguration struct {
	Profile           string    `yaml:"profile"`
	KubeConfigPath    string    `yaml:"kube_config_path"`
	ControllerCluster Cluster   `yaml:"controller"`
	WorkerClusters    []Cluster `yaml:"workers"`
}

type Cluster struct {
	Name                string `yaml:"name"`
	ContextName         string `yaml:"context_name"`
	KubeConfigPath      string `yaml:"kube_config_path"`
	ControlPlaneAddress string `yaml:"control_plane_address"`
	NodeIP              string `yaml:"node_ip"`
}
