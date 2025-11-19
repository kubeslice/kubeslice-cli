package internal

import (
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const (
	iPerfClientFileName              = "iperf-client.yaml"
	iPerfServerFileName              = "iperf-server.yaml"
	iPerfServerServiceExportFileName = "iperf-server-service-export.yaml"
	iPerfNamespace                   = "iperf"
)

const iPerfServiceExportTemplate = `
---
apiVersion: networking.kubeslice.io/v1beta1
kind: ServiceExport
metadata:
  name: iperf-server
  namespace: iperf
spec:
  slice: demo
  selector:
    matchLabels:
      app: iperf-server
  ingressEnabled: false
  ports:
  - name: tcp
    containerPort: 5201
    protocol: TCP
`

const iPerfServerTemplate = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: iperf
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: iperf-server
  namespace: iperf
  labels:
    app: iperf-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: iperf-server
  template:
    metadata:
      labels:
        app: iperf-server
    spec:
      containers:
      - name: iperf
        image: mlabbe/iperf
        imagePullPolicy: Always
        args:
          - '-s'
          - '-p'
          - '5201'
        ports:
        - containerPort: 5201
          name: server
      - name: sidecar
        image: nicolaka/netshoot
        imagePullPolicy: IfNotPresent
        command: ["/bin/sleep", "3650d"]
        securityContext:
          capabilities:
            add: ["NET_ADMIN"]
          allowPrivilegeEscalation: true
          privileged: true
`

const iPerfClientTemplate = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: iperf
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: iperf-sleep
  namespace: iperf
  labels:
    app: iperf-sleep
spec:
  replicas: 1
  selector:
    matchLabels:
      app: iperf-sleep
  template:
    metadata:
      labels:
        app: iperf-sleep
    spec:
      containers:
      - name: iperf
        image: mlabbe/iperf
        imagePullPolicy: Always
        command: ["/bin/sleep", "3650d"]
      - name: sidecar
        image: nicolaka/netshoot
        imagePullPolicy: IfNotPresent
        command: ["/bin/sleep", "3650d"]
        securityContext:
          capabilities:
            add: ["NET_ADMIN"]
          allowPrivilegeEscalation: true
          privileged: true
`

// Function variables for testing
var (
	applyKubectlManifestFunc = ApplyKubectlManifest
	podVerificationFuncIPerf = PodVerification
)

func InstallIPerf(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("\nInstalling iPerf Application...")

	clientFileName := iPerfClientFileName
	serverFileName := iPerfServerFileName
	cc := ApplicationConfiguration.Configuration.ClusterConfiguration
	wc := cc.WorkerClusters

	applyKubectlManifestFunc(kubesliceDirectory+"/"+serverFileName, iPerfNamespace, &wc[0])
	util.Printf("%s Applied %s to %s", util.Tick, serverFileName, wc[0].Name)
	util.Sleep(200 * time.Millisecond)

	util.Printf("%s Waiting for iPerf Server pod to be running...", util.Wait)
	podVerificationFuncIPerf("Waiting for iPerf Server pod to be running", wc[0], iPerfNamespace)
	util.Printf("%s Successfully installed iPerf Server on %s...", util.Tick, wc[0].Name)

	for i := 1; i < len(wc); i++ {
		applyKubectlManifestFunc(kubesliceDirectory+"/"+clientFileName, iPerfNamespace, &wc[i])
		util.Printf("%s Applied %s to %s", util.Tick, clientFileName, wc[i].Name)
		util.Sleep(200 * time.Millisecond)

		util.Printf("%s Waiting for iPerf Client pod to be running...", util.Wait)
		podVerificationFuncIPerf("Waiting for iPerf Client pod to be running", wc[i], iPerfNamespace)
		util.Printf("%s Successfully installed iPerf Client on %s...", util.Tick, wc[i].Name)
	}

	util.Printf("Installed IPerf Applications")
}

func GenerateIPerfManifests() {
	// --- Client Manifests
	util.DumpFile(iPerfClientTemplate, kubesliceDirectory+"/"+iPerfClientFileName)
	util.Printf("%s Generated iPerf Client manifest %s", util.Tick, iPerfClientFileName)
	util.Sleep(200 * time.Millisecond)

	// --- Server Manifests
	util.DumpFile(iPerfServerTemplate, kubesliceDirectory+"/"+iPerfServerFileName)
	util.Printf("%s Generated iPerf Server manifest %s", util.Tick, iPerfServerFileName)
	util.Sleep(200 * time.Millisecond)
}

func GenerateIPerfServiceExportManifest(ApplicationConfiguration *ConfigurationSpecs) {
	util.DumpFile(iPerfServiceExportTemplate, kubesliceDirectory+"/"+iPerfServerServiceExportFileName)
	util.Printf("%s Generated iPerf Server Service Export manifest %s for cluster %s", util.Tick, iPerfServerServiceExportFileName, ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters[0].Name)
	util.Sleep(200 * time.Millisecond)
}

func ApplyIPerfServiceExportManifest(ApplicationConfiguration *ConfigurationSpecs) {
	applyKubectlManifestFunc(kubesliceDirectory+"/"+iPerfServerServiceExportFileName, iPerfNamespace, &ApplicationConfiguration.Configuration.ClusterConfiguration.WorkerClusters[0])
}

func RolloutRestartIPerf(ApplicationConfiguration *ConfigurationSpecs) {
	clusters := getAllClusters(&ApplicationConfiguration.Configuration.ClusterConfiguration)[1:]

	err := util.CommandExecutor.Execute("kubectl", "rollout", "restart", "deployment/iperf-server", "-n", iPerfNamespace, "--context="+clusters[0].ContextName, "--kubeconfig="+clusters[0].KubeConfigPath)
	if err != nil {
		util.Fatalf("Process failed %v", err)
	}

	for i := 1; i < len(clusters); i++ {
		err = util.CommandExecutor.Execute("kubectl", "rollout", "restart", "deployment/iperf-sleep", "-n", iPerfNamespace, "--context="+clusters[i].ContextName, "--kubeconfig="+clusters[i].KubeConfigPath)
		if err != nil {
			util.Fatalf("Process failed %v", err)
		}
	}

}
