package internal

import (
	"fmt"
	"strings"

	"github.com/kubeslice/kubeslice-cli/util"
)

const windowsEnvSet = `
PowerShell(ps):
	$env:KUBECONFIG=` + KubeconfigPath + `

Command Prompt(cmd):
	set KUBECONFIG=` + KubeconfigPath + `
`

const linuxEnvSet = `export KUBECONFIG=` + KubeconfigPath

const printVerificationStepsTemplate = `
========================================================================
Now that the KubeSlice Cluster Setup (1 Controller + 2 Worker) is complete 
with a sample iPerf deployment, you can verify the cluster inter-connectivity 
that KubeSlice provides.

Verify the iPerf Connectivity.
Here, the iPerf client, which is installed on Worker 1, will attempt to 
reach out to iPerf service, which is installed on Worker 2.

Note: The DNS propagation may take a minute or two.

%s %s
`
const printEntVerificationStepsTemplate = `
========================================================================
KubeSlice Enterprise Setup (1 Controller + 2 Worker) is complete 
with a sample iPerf deployment.
You can now access Kubeslice Manager UI using the following URL:

%s %s

To access Kubeslice Manager use the following token in the login screen:

%s %s

Verify the iPerf Connectivity.
Here, the iPerf client, which is installed on Worker 1, will attempt to 
reach out to iPerf service, which is installed on Worker 2.

Note: The DNS propagation may take a minute or two.

%s %s
`

const printNextStepsTemplateForSliceInstallation = `

========================================================================
Now that the KubeSlice Cluster Setup (1 Controller + 2 Worker) is complete 
with a sample iPerf deployment, you can verify the cluster inter-connectivity 
that KubeSlice provides.

You can verify the connectivity before the creation of Slice using the following command:

%s %s

Since the slice hasn't been created yet, the connectivity is not present.

===
Now, you can create a Slice using the following command:

%s %s

===
The slice propagation will take a few seconds. You can run the following commands to verify that slice
has propagated to worker clusters

For Worker 1
%s %s

For Worker 2
%s %s

===
Once the Slice has propagated to worker clusters, you need to restart the iPerf deployment to onboard the applications on the slice

For Worker 1
%s %s

For Worker 2
%s %s

===
Before you can verify the connectivity, the iPerf server needs to be exported for visibility. Run the following command
to export the iPerf server

%s %s

===
Verify the iPerf Connectivity Again.
Note: The DNS propagation may take a minute or two.

%s %s
`

// Function variables for testing
var (
	getUIAdminTokenFunc = GetUIAdminToken
	getUIEndpointFunc   = GetUIEndpoint
)

// buildKubectlCommand constructs a kubectl command string for display
func buildKubectlCommand(cluster *Cluster, args ...string) string {
	kubectlPath := util.ExecutablePaths["kubectl"]
	if kubectlPath == "" {
		kubectlPath = "kubectl"
	}
	parts := []string{kubectlPath}

	if cluster != nil {
		parts = append(parts, "--context="+cluster.ContextName)
		parts = append(parts, "--kubeconfig="+cluster.KubeConfigPath)
	}

	parts = append(parts, args...)
	return strings.Join(parts, " ")
}

func PrintNextSteps(verificationOnly bool, ApplicationConfiguration *ConfigurationSpecs) {
	if verificationOnly {
		printVerificationSteps(ApplicationConfiguration)
	} else {
		printNamespaceIsolationSteps(ApplicationConfiguration)
	}
}

func printVerificationSteps(ApplicationConfiguration *ConfigurationSpecs) {
	var template string
	username := "admin"
	clusters := ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters
	if len(clusters) < 2 {
		util.Printf("Error: At least 2 worker clusters required\n")
		return
	}

	iperfCommand := buildKubectlCommand(
		&clusters[1],
		"exec", "-it", "deploy/iperf-sleep",
		"-c", "iperf",
		"-n", "iperf",
		"--",
		"iperf", "-c", "iperf-server.iperf.svc.slice.local",
		"-p", "5201",
		"-i", "1",
		"-b", "10Mb;",
	)

	if ApplicationConfiguration.Configuration.ClusterConfiguration.Profile == ProfileEntDemo {
		token := getUIAdminTokenFunc(
			&ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster,
			username,
			ApplicationConfiguration.Configuration.KubeSliceConfiguration.ProjectName)
		endpoint := getUIEndpointFunc(
			&ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster,
			ProfileEntDemo)

		template = fmt.Sprintf(printEntVerificationStepsTemplate,
			util.Globe, endpoint,
			util.Lock, token,
			util.Run, iperfCommand,
		)
	} else {
		template = fmt.Sprintf(printVerificationStepsTemplate,
			util.Run, iperfCommand,
		)
	}
	util.Printf(template)
}

func printNamespaceIsolationSteps(ApplicationConfiguration *ConfigurationSpecs) {
	cc := ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster
	wc := ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters
	if len(wc) < 2 {
		util.Printf("Error: At least 2 worker clusters required\n")
		return
	}

	iperfCommand := buildKubectlCommand(
		&wc[1],
		"exec", "-it", "deploy/iperf-sleep",
		"-c", "iperf",
		"-n", "iperf",
		"--",
		"iperf", "-c", "iperf-server.iperf.svc.slice.local",
		"-p", "5201",
		"-i", "1",
		"-b", "10Mb;",
	)

	sliceApplyCommand := buildKubectlCommand(
		&cc,
		"apply", "-f", kubesliceDirectory+"/"+sliceTemplateFileName,
	)

	sliceVerifyCommandWorker1 := buildKubectlCommand(
		&wc[0],
		"get", "slice", "-n", "kubeslice-system",
	)

	sliceVerifyCommandWorker2 := buildKubectlCommand(
		&wc[1],
		"get", "slice", "-n", "kubeslice-system",
	)

	applyIPerfWorker1 := buildKubectlCommand(
		&wc[0],
		"rollout", "restart", "deployment/iperf-server",
		"-n", "iperf",
	)

	applyIPerfWorker2 := buildKubectlCommand(
		&wc[1],
		"rollout", "restart", "deployment/iperf-sleep",
		"-n", "iperf",
	)

	applyIPerfServiceExportWorker2 := buildKubectlCommand(
		&wc[0],
		"apply", "-f", kubesliceDirectory+"/"+iPerfServerServiceExportFileName,
		"-n", "iperf",
	)

	template := fmt.Sprintf(printNextStepsTemplateForSliceInstallation,
		util.Run, iperfCommand,
		util.Run, sliceApplyCommand,
		util.Run, sliceVerifyCommandWorker1,
		util.Run, sliceVerifyCommandWorker2,
		util.Run, applyIPerfWorker1,
		util.Run, applyIPerfWorker2,
		util.Run, applyIPerfServiceExportWorker2,
		util.Run, iperfCommand,
	)

	util.Printf(template)
}
