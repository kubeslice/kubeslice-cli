package pkg

import (
	"github.com/kubeslice/kubeslice-cli/pkg/internal"
)

func CreateProject() {
	ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName = CliOptions.ObjectName
	internal.CreateKubeSliceProject(ApplicationConfiguration, CliOptions)
}

func GetProject() {
	internal.GetKubeSliceProject(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}

func DeleteProject() {
	internal.DeleteKubeSliceProject(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}

func EditProject() {
	internal.EditKubeSliceProject(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}

func DescribeProject() {
	internal.DescribeKubeSliceProject(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}
