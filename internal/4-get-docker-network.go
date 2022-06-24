package internal

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kubeslice/slicectl/util"
)

var dockerNetworkMap = map[string]string{
	controllerName: "",
	worker1Name:    "",
	worker2Name:    "",
}

func PopulateDockerNetworkMap() {
	util.Printf("\nFetching Network Address for Created Clusters...")

	runDockerInspectForNodeIP(controllerName)
	util.Printf("%s Fetched Network Address for %s : %s", util.Tick, controllerName, dockerNetworkMap[controllerName])
	time.Sleep(200 * time.Millisecond)

	runDockerInspectForNodeIP(worker1Name)
	util.Printf("%s Fetched Network Address for %s : %s", util.Tick, worker1Name, dockerNetworkMap[worker1Name])
	time.Sleep(200 * time.Millisecond)

	runDockerInspectForNodeIP(worker2Name)
	util.Printf("%s Fetched Network Address for %s : %s", util.Tick, worker2Name, dockerNetworkMap[worker2Name])
	time.Sleep(200 * time.Millisecond)

	util.Printf("Successfully fetched network addresses for created clusters.")
}

func runDockerInspectForNodeIP(clusterName string) {
	var outB, errB bytes.Buffer
	err := util.RunCommandCustomIO("docker", &outB, &errB, true, "inspect", "--format={{.NetworkSettings.Networks.kind.IPAddress}}", fmt.Sprintf("%s-control-plane", clusterName))
	if err != nil {
		util.Printf("%s Failed to run command\nOutput: %s\nError: %s %v", util.Cross, outB.String(), errB.String(), err)
		os.Exit(1)
	}
	dockerNetworkMap[clusterName] = strings.TrimSpace(outB.String())
}
