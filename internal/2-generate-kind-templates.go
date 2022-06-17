package internal

import (
	"log"
	"os"
	"time"

	"github.com/kubeslice/kubeslice-installer/util"
)

const (
	kubesliceDirectory = "kubeslice"
	kindSubDirectory   = "kind"

	controllerFilename = "controller.yaml"
	worker1Filename    = "worker-1.yaml"
	worker2Filename    = "worker-2.yaml"

	controllerName = "ks-ctrl"
	worker1Name    = "ks-w-1"
	worker2Name    = "ks-w-2"
)

const kubesliceControllerTemplate = `
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: ` + controllerName + `
networking:
  disableDefaultCNI: true # disable kindnet
  podSubnet: 192.168.0.0/16 # set to Calico's default subnet
nodes:
  - role: control-plane
    image: kindest/node:v1.22.9
`

const kubesliceWorker1Template = `
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: ` + worker1Name + `
networking:
  disableDefaultCNI: true # disable kindnet
  podSubnet: 192.168.0.0/16 # set to Calico's default subnet
nodes:
  - role: control-plane
    image: kindest/node:v1.22.9
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "kubeslice.io/node-type=gateway"
`

const kubesliceWorker2Template = `
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: ` + worker2Name + `
networking:
  disableDefaultCNI: true # disable kindnet
  podSubnet: 192.168.0.0/16 # set to Calico's default subnet
nodes:
  - role: control-plane
    image: kindest/node:v1.22.9
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "kubeslice.io/node-type=gateway"
`

func DeleteKubeSliceDirectory() {
	err := os.RemoveAll(kubesliceDirectory)
	if err != nil {
		log.Fatalf("\nFailed to delete directory %s\n", kubesliceDirectory)
	}
}

func GenerateKubeSliceDirectory() {
	util.CreateDirectoryPath(kubesliceDirectory)
}

func GenerateKindConfiguration() {
	directory := kubesliceDirectory + "/" + kindSubDirectory
	util.Printf("\nGenerating Kind configuration files to %s directory...", directory)

	util.CreateDirectoryPath(directory)

	util.DumpFile(kubesliceControllerTemplate, directory+"/"+controllerFilename)
	util.Printf("%s Generated %s", util.Tick, controllerFilename)
	time.Sleep(200 * time.Millisecond)

	util.DumpFile(kubesliceWorker1Template, directory+"/"+worker1Filename)
	util.Printf("%s Generated %s", util.Tick, worker1Filename)
	time.Sleep(200 * time.Millisecond)

	util.DumpFile(kubesliceWorker2Template, directory+"/"+worker2Filename)
	util.Printf("%s Generated %s", util.Tick, worker2Filename)
	time.Sleep(200 * time.Millisecond)
	util.Printf("Kind configuration files generated.")
}

