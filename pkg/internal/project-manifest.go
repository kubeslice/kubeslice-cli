package internal

import (
	"fmt"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
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
    readWrite: %s
`

func CreateKubeSliceProject(ApplicationConfiguration *ConfigurationSpecs, cliOptions *CliOptionsStruct) {
	util.Printf("\nCreating KubeSlice Project...")

	generateKubeSliceProjectManifest(ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName, ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectUsers)
	util.Printf("%s Generated project manifest %s", util.Tick, projectFileName)
	time.Sleep(200 * time.Millisecond)
	if cliOptions != nil {
		if cliOptions.FileName == "" {
			cliOptions.FileName = kubesliceDirectory + "/" + projectFileName
		}
		ApplyKubectlManifest(cliOptions.FileName, cliOptions.Namespace, cliOptions.Cluster)
	} else {
		ApplyKubectlManifest(kubesliceDirectory+"/"+projectFileName, KUBESLICE_CONTROLLER_NAMESPACE, &ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster)
	}
	util.Printf("%s Applied %s", util.Tick, projectFileName)
	time.Sleep(200 * time.Millisecond)
	util.Printf("Created KubeSlice Project.")
}

func GetKubeSliceProject(projectName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nFetching KubeSlice Project...")
	GetKubectlResources(ProjectObject, projectName, namespace, controllerCluster, "")
	time.Sleep(200 * time.Millisecond)
}
func generateKubeSliceProjectManifest(projectName string, users []string) {
	if len(users) == 0 {
		users = []string{"admin"}
	}
	userString := "\n"
	for _, user := range users {
		userString = fmt.Sprintf(`%s      - %s%s`, userString, user, "\n")
	}
	util.DumpFile(fmt.Sprintf(kubesliceProjectTemplate, projectName, userString), kubesliceDirectory+"/"+projectFileName)
}

func DeleteKubeSliceProject(projectName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDeleting KubeSlice Project...")
	DeleteKubectlResources(ProjectObject, projectName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}

func EditKubeSliceProject(projectName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nEditing KubeSlice Project...")
	EditKubectlResources(ProjectObject, projectName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}

func DescribeKubeSliceProject(projectName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDescribe KubeSlice Project...")
	DescribeKubectlResources(ProjectObject, projectName, namespace, controllerCluster)
	time.Sleep(200 * time.Millisecond)
}
