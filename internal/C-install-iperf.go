package internal

import (
	"log"
	"time"

	"github.com/kubeslice/slicectl/util"
)

const (
	iPerfClientFileName              = "iperf-client-worker-1.yaml"
	iPerfServerFileName              = "iperf-server-worker-2.yaml"
	iPerfServerServiceExportFileName = "iperf-server-service-export-worker-2.yaml"
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

func InstallIPerf() {
	util.Printf("\nInstalling iPerf Application...")

	clientFileName := iPerfClientFileName
	serverFileName := iPerfServerFileName

	util.ApplyKubectlManifest(kubesliceDirectory+"/"+clientFileName, "iperf", worker1Name)
	util.Printf("%s Applied %s to %s", util.Tick, clientFileName, worker1Name)
	time.Sleep(200 * time.Millisecond)

	util.Printf("%s Waiting for iPerf Client pod to be running...", util.Wait)
	util.PodVerification("Waiting for iPerf Client pod to be running", worker1Name, "iperf")
	util.Printf("%s Successfully installed iPerf Client on %s...", util.Tick, worker1Name)

	util.ApplyKubectlManifest(kubesliceDirectory+"/"+serverFileName, "iperf", worker2Name)
	util.Printf("%s Applied %s to %s", util.Tick, serverFileName, worker2Name)
	time.Sleep(200 * time.Millisecond)

	util.Printf("%s Waiting for iPerf Server pod to be running...", util.Wait)
	util.PodVerification("Waiting for iPerf Server pod to be running", worker2Name, "iperf")
	util.Printf("%s Successfully installed iPerf Server on %s...", util.Tick, worker2Name)

	util.Printf("Installed IPerf Applications")
}

func GenerateIPerfManifests() {
	// --- Client Manifests
	util.DumpFile(iPerfClientTemplate, kubesliceDirectory+"/"+iPerfClientFileName)
	util.Printf("%s Generated iPerf Client manifest %s for cluster %s", util.Tick, iPerfClientFileName, worker1Name)
	time.Sleep(200 * time.Millisecond)

	// --- Server Manifests
	util.DumpFile(iPerfServerTemplate, kubesliceDirectory+"/"+iPerfServerFileName)
	util.Printf("%s Generated iPerf Server manifest %s for cluster %s", util.Tick, iPerfServerFileName, worker2Name)
	time.Sleep(200 * time.Millisecond)
}

func GenerateIPerfServiceExportManifest() {
	util.DumpFile(iPerfServiceExportTemplate, kubesliceDirectory+"/"+iPerfServerServiceExportFileName)
	util.Printf("%s Generated iPerf Server Service Export manifest %s for cluster %s", util.Tick, iPerfServerServiceExportFileName, worker2Name)
	time.Sleep(200 * time.Millisecond)
}

func ApplyIPerfServiceExportManifest() {
	util.ApplyKubectlManifest(kubesliceDirectory+"/"+iPerfServerServiceExportFileName, "iperf", worker2Name)
}

func RolloutRestartIPerf() {
	err := util.RunCommand("kubectl", "rollout", "restart", "deployment/iperf-server", "-n", "iperf", "--context=kind-"+worker2Name)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
	err = util.RunCommand("kubectl", "rollout", "restart", "deployment/iperf-sleep", "-n", "iperf", "--context=kind-"+worker1Name)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}
