package internal

import (
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
  name: demo
  namespace: kubeslice-controller
spec:
  serviceAccount:
    readWrite:
      - john
`

func CreateKubeSliceProject() {
	util.Printf("\nCreating KubeSlice Project...")

	generateKubeSliceProjectManifest()
	util.Printf("%s Generated project manifest %s", util.Tick, projectFileName)
	time.Sleep(200 * time.Millisecond)

	util.ApplyKubectlManifest(kubesliceDirectory+"/"+projectFileName, "kubeslice-controller", controllerName)
	util.Printf("%s Applied %s", util.Tick, projectFileName)
	time.Sleep(200 * time.Millisecond)
	util.Printf("Created KubeSlice Project.")
}

func generateKubeSliceProjectManifest() {
	util.DumpFile(kubesliceProjectTemplate, kubesliceDirectory+"/"+projectFileName)
}
