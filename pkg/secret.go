package pkg

import (
	"github.com/kubeslice/slicectl/pkg/internal"
)

func GetSecrets(worker string) {
	internal.GetSecrets(worker, CliOptions.Namespace, CliOptions.Cluster, CliOptions.OutputFormat)
}
