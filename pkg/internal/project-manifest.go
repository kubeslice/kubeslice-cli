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

// Function variables for testing
var (
	applyKubectlManifestFuncProject     = ApplyKubectlManifest
	getKubectlResourcesFuncProject      = GetKubectlResources
	deleteKubectlResourcesFuncProject   = DeleteKubectlResources
	editKubectlResourcesFuncProject     = EditKubectlResources
	describeKubectlResourcesFuncProject = DescribeKubectlResources
)

func CreateKubeSliceProject(ApplicationConfiguration *ConfigurationSpecs, cliOptions *CliOptionsStruct) {
	util.Printf("\nCreating KubeSlice Project...")

	generateKubeSliceProjectManifest(ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName, ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectUsers)
	util.Printf("%s Generated project manifest %s", util.Tick, projectFileName)
	util.SystemClock.Sleep(200 * time.Millisecond)

	if cliOptions != nil {
		if cliOptions.FileName == "" {
			cliOptions.FileName = kubesliceDirectory + "/" + projectFileName
		}
		applyKubectlManifestFuncProject(cliOptions.FileName, cliOptions.Namespace, cliOptions.Cluster)
	} else {
		applyKubectlManifestFuncProject(kubesliceDirectory+"/"+projectFileName, KUBESLICE_CONTROLLER_NAMESPACE, &ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster)
	}
	util.Printf("%s Applied %s", util.Tick, projectFileName)
	util.SystemClock.Sleep(3 * time.Second)
	util.Printf("Created KubeSlice Project.")
}

func GetKubeSliceProject(projectName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nFetching KubeSlice Project...")
	getKubectlResourcesFuncProject(ProjectObject, projectName, namespace, controllerCluster, "")
	util.SystemClock.Sleep(200 * time.Millisecond)
}

func generateKubeSliceProjectManifest(projectName string, users []string) {
	if len(users) == 0 {
		users = []string{"admin"}
	}
	userString := "\n"
	for _, user := range users {
		userString = fmt.Sprintf("%s      - %s%s", userString, user, "\n")
	}
	util.DumpFile(fmt.Sprintf(kubesliceProjectTemplate, projectName, userString), kubesliceDirectory+"/"+projectFileName)
}

func DeleteKubeSliceProject(projectName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDeleting KubeSlice Project...")
	deleteKubectlResourcesFuncProject(ProjectObject, projectName, namespace, controllerCluster)
	util.SystemClock.Sleep(200 * time.Millisecond)
}

func EditKubeSliceProject(projectName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nEditing KubeSlice Project...")
	editKubectlResourcesFuncProject(ProjectObject, projectName, namespace, controllerCluster)
	util.SystemClock.Sleep(200 * time.Millisecond)
}

func DescribeKubeSliceProject(projectName string, namespace string, controllerCluster *Cluster) {
	util.Printf("\nDescribe KubeSlice Project...")
	describeKubectlResourcesFuncProject(ProjectObject, projectName, namespace, controllerCluster)
	util.SystemClock.Sleep(200 * time.Millisecond)
}
