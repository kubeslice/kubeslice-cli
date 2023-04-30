package internal

const (
	KUBESLICE_CONTROLLER_NAMESPACE = "kubeslice-controller"
	ProjectObject                  = "projects.controller.kubeslice.io"
	ClusterObject                  = "clusters.controller.kubeslice.io"
	SliceConfigObject              = "sliceconfigs.controller.kubeslice.io"
	ServiceExportConfigObject      = "serviceexportconfigs.controller.kubeslice.io"

	Kind_Component                = "kind"
	Calico_Component              = "calico"
	Controller_Component          = "controller"
	Worker_registration_Component = "worker-registration"
	UI_install_Component          = "ui"
	Worker_Component              = "worker"
	Demo_Component                = "demo"
	CertManager_Component         = "cert-manager"
	Prometheus_Component          = "prometheus"
	SecretObject                  = "secrets"
	OutputFormatYaml              = "yaml"
	OutputFormatJson              = "json"
)
