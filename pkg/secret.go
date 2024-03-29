package pkg

import (
	"github.com/kubeslice/kubeslice-cli/pkg/internal"
)

func GetSecrets(worker string) {
	internal.GetSecrets(worker, CliOptions.Namespace, CliOptions.Cluster, CliOptions.OutputFormat)
}
