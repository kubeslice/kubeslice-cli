package internal

import (
	"fmt"
	"time"

	"github.com/kubeslice/slicectl/util"
)

const (
	projectFileName = "project.yaml"
)

const kubesliceProjectTemplate = `
apiVersion: controller.kubeslice.io/v1alpha1
kind: Project
metadata:
  name: %s
  namespace: kubeslice-controller
spec:
  serviceAccount:
    readWrite:
      - john
`

func CreateKubeSliceProject(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nCreating KubeSlice Project...")

	generateKubeSliceProjectManifest(ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName)
	util.Printf("%s Generated project manifest %s", util.Tick, projectFileName)
	time.Sleep(200 * time.Millisecond)

	ApplyKubectlManifest(kubesliceDirectory+"/"+projectFileName, "kubeslice-controller", ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster)
	util.Printf("%s Applied %s", util.Tick, projectFileName)
	time.Sleep(200 * time.Millisecond)
	util.Printf("Created KubeSlice Project.")
}

func GetKubeSliceProject(projectName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nFetching KubeSlice Project...")
	GetKubectlResources(ProjectObject, projectName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}
func generateKubeSliceProjectManifest(projectName string) {
	util.DumpFile(fmt.Sprintf(kubesliceProjectTemplate, projectName), kubesliceDirectory+"/"+projectFileName)
}
