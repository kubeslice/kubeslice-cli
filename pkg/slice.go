package pkg

import (
	"github.com/kubeslice/slicectl/pkg/internal"
)

func CreateSliceConfig(worker []string) {
	if len(CliOptions.FileName) != 0 {
		internal.CreateSliceConfig(CliOptions.Namespace, CliOptions.Cluster, CliOptions.FileName)
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
