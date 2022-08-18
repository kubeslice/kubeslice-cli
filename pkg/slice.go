package pkg

import (
	"github.com/kubeslice/slicectl/pkg/internal"
	"github.com/kubeslice/slicectl/util"
)

func CreateSliceConfig() {
	util.Printf("testing create.")
}

func GetSliceConfig() {
	internal.GetSliceConfig(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}

func DeleteSliceConfig() {
	internal.DeleteSliceConfig(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}
