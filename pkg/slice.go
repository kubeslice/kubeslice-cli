package pkg

import (
	"github.com/kubeslice/slicectl/pkg/internal"
)

func CreateSliceConfig(filename string, worker []string) {
	if len(filename) != 0 {
		internal.CreateSliceConfig(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster, filename)
	} else if len(worker) != 0 {
		internal.GenerateSliceConfiguration(ApplicationConfiguration, worker, CliOptions.ObjectName, CliOptions.Namespace)
		internal.ApplyFile("kubeslice/slice-"+CliOptions.ObjectName+".yaml", CliOptions.Namespace, CliOptions.Cluster)
	}
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
