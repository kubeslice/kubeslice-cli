package pkg

import (
	"github.com/kubeslice/slicectl/pkg/internal"
)

func CreateSliceConfig(filename string, worker []string) {
	internal.CreateSliceConfig(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster, filename, worker)
}

func GetSliceConfig() {
	internal.GetSliceConfig(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}

func DeleteSliceConfig() {
	internal.DeleteSliceConfig(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}

func EditSliceConfig() {
	internal.EditSliceConfig(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}

func DescribeSliceConfig() {
	internal.DescribeSliceConfig(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}
