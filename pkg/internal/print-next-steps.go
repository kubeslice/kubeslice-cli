package internal

import (
	"fmt"
	"os/exec"

	"github.com/kubeslice/slicectl/util"
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

func PrintNextSteps(verificationOnly bool, ApplicationConfiguration *ConfigurationSpecs) {
	if verificationOnly {
		printVerificationSteps(ApplicationConfiguration)
	} else {
		printNamespaceIsolationSteps(ApplicationConfiguration)
	}
}

func printVerificationSteps(ApplicationConfiguration *ConfigurationSpecs) {
	clusters := ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters
	iperfCommand := exec.Command(util.ExecutablePaths["kubectl"], "--context="+clusters[1].ContextName, "--kubeconfig="+clusters[1].KubeConfigPath, "exec", "-it", "deploy/iperf-sleep", "-c", "iperf", "-n", "iperf", "--", "iperf", "-c", "iperf-server.iperf.svc.slice.local", "-p", "5201", "-i", "1", "-b", "10Mb;")
	template := fmt.Sprintf(printVerificationStepsTemplate,
		util.Run, iperfCommand.String(),
	)
	util.Printf(template)
}

func printNamespaceIsolationSteps(ApplicationConfiguration *ConfigurationSpecs) {
	cc := ApplicationConfiguration.Configuration.ClusterConfiguration.ControllerCluster
	wc := ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters
	iperfCommand := exec.Command(util.ExecutablePaths["kubectl"], "--context="+wc[1].ContextName, "--kubeconfig="+wc[1].KubeConfigPath, "exec", "-it", "deploy/iperf-sleep", "-c", "iperf", "-n", "iperf", "--", "iperf", "-c", "iperf-server.iperf.svc.slice.local", "-p", "5201", "-i", "1", "-b", "10Mb;")
	sliceApplyCommand := exec.Command(util.ExecutablePaths["kubectl"], "--context="+cc.ContextName, "--kubeconfig="+cc.KubeConfigPath, "apply", "-f", kubesliceDirectory+"/"+sliceTemplateFileName)
	sliceVerifyCommandWorker1 := exec.Command(util.ExecutablePaths["kubectl"], "--context="+wc[0].ContextName, "--kubeconfig="+wc[0].KubeConfigPath, "get", "slice", "-n", "kubeslice-system")
	sliceVerifyCommandWorker2 := exec.Command(util.ExecutablePaths["kubectl"], "--context="+wc[1].ContextName, "--kubeconfig="+wc[1].KubeConfigPath, "get", "slice", "-n", "kubeslice-system")
	applyIPerfWorker1 := exec.Command(util.ExecutablePaths["kubectl"], "rollout ", "restart", "deployment/iperf-server", "-n", "iperf", "--context="+wc[0].ContextName, "--kubeconfig="+wc[0].KubeConfigPath)
	applyIPerfWorker2 := exec.Command(util.ExecutablePaths["kubectl"], "rollout ", "restart", "deployment/iperf-sleep", "-n", "iperf", "--context="+wc[1].ContextName, "--kubeconfig="+wc[1].KubeConfigPath)
	applyIPerfServiceExportWorker2 := exec.Command(util.ExecutablePaths["kubectl"], "--context="+wc[0].ContextName, "--kubeconfig="+wc[0].KubeConfigPath, "apply ", "-f", kubesliceDirectory+"/"+iPerfServerServiceExportFileName, "-n", "iperf")
	template := fmt.Sprintf(printNextStepsTemplateForSliceInstallation,
		util.Run, iperfCommand.String(),
		util.Run, sliceApplyCommand.String(),
		util.Run, sliceVerifyCommandWorker1.String(),
		util.Run, sliceVerifyCommandWorker2.String(),
		util.Run, applyIPerfWorker1.String(),
		util.Run, applyIPerfWorker2.String(),
		util.Run, applyIPerfServiceExportWorker2.String(),
		util.Run, iperfCommand.String(),
	)
	util.Printf(template)
}
