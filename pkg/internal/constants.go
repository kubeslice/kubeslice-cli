package internal

const (
	ProjectObject     = "projects.controller.kubeslice.io"
	ClusterObject     = "clusters.controller.kubeslice.io"
	SliceConfigObject = "sliceconfigs.controller.kubeslice.io"

	// skipStep suffix is for the set of installation steps to skip.
	Kind_skipStep                = "kind"
	Calico_skipStep              = "calico"
	Controller_skipStep          = "controller"
	Worker_registration_skipStep = "worker-registration"
	Worker_skipStep              = "worker"
)
