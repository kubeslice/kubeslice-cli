package pkg

import "github.com/kubeslice/kubeslice-cli/pkg/internal"

func GetUIEndpoint() {
	internal.GetUIEndpoint(CliOptions.Cluster)
}
