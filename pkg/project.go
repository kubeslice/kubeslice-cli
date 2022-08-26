package pkg

import (
	"github.com/kubeslice/slicectl/pkg/internal"
	"github.com/kubeslice/slicectl/util"
)

func CreateProject() {
	util.Printf("testing create.")
}

func GetProject() {
	internal.GetKubeSliceProject(CliOptions.ObjectName, CliOptions.Namespace, CliOptions.Cluster)
}
