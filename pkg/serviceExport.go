package pkg

import (
	"github.com/kubeslice/kubeslice-cli/pkg/internal"
)

func CreateServiceExportConfig(filename string) {
	internal.CreateServiceExportConfig(CliOptions.Namespace, CliOptions.Cluster, filename)
}

func GetServiceExportConfig() {
	internal.GetServiceExportConfig(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}

func DeleteServiceExportConfig() {
	internal.DeleteServiceExportConfig(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}

func EditServiceExportConfig() {
	internal.EditServiceExportConfig(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}

func DescribeServiceExportConfig() {
	internal.DescribeServiceExportConfig(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}
