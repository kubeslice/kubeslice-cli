package pkg

import "github.com/kubeslice/kubeslice-cli/pkg/internal"

var getUIEndpointFunc = internal.GetUIEndpoint

func GetUIEndpoint() {
	getUIEndpointFunc(CliOptions.Cluster, ApplicationConfiguration.Configuration.ClusterConfiguration.Profile)
}
