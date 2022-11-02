package internal

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const (
	kubesliceDirectory = "kubeslice"
	kindSubDirectory   = "kind"
)

const kubesliceControllerTemplate = `
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: %s
networking:
  disableDefaultCNI: true # disable kindnet
  podSubnet: 192.168.0.0/16 # set to Calico's default subnet
nodes:
  - role: control-plane
    image: kindest/node:v1.22.9
`

const kubesliceWorkerTemplate = `
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: %s
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

func GenerateKindConfiguration(ApplicationConfiguration *ConfigurationSpecs) {
	cc := ApplicationConfiguration.Configuration.ClusterConfiguration
	directory := kubesliceDirectory + "/" + kindSubDirectory
	util.Printf("\nGenerating Kind configuration files to %s directory...", directory)

	util.CreateDirectoryPath(directory)

	util.DumpFile(fmt.Sprintf(kubesliceControllerTemplate, cc.ControllerCluster.Name), directory+"/"+cc.ControllerCluster.Name+".yaml")
	util.Printf("%s Generated %s", util.Tick, directory+"/"+cc.ControllerCluster.Name+".yaml")
	time.Sleep(200 * time.Millisecond)

	for _, cluster := range cc.WorkerClusters {
		util.DumpFile(fmt.Sprintf(kubesliceWorkerTemplate, cluster.Name), directory+"/"+cluster.Name+".yaml")
		util.Printf("%s Generated %s", util.Tick, directory+"/"+cluster.Name+".yaml")
		time.Sleep(200 * time.Millisecond)
	}
}
