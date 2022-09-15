package pkg

import (
	"github.com/kubeslice/slicectl/pkg/internal"
)

func RegisterWorker() {
	ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters = nil
	ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters = append(ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters, internal.Cluster{
		Name: CliOptions.ObjectName,
	})
	internal.RegisterWorkerClusters(ApplicationConfiguration, CliOptions)
}

func GetWorker() {
	internal.GetKubeSliceCluster(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster, CliOptions.OutputFormat)
}

func RemoveWorker() {
	internal.DeleteKubeSliceCluster(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}

func EditWorker() {
	internal.EditKubeSliceCluster(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}

func DescribeWorker() {
	internal.DescribeKubeSliceCluster(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}
