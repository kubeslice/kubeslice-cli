package internal

var kindExecutableMessage = map[string]string{
	"windows": `
To Install Kind CLI, Please visit https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries.
Make sure the downloaded file has an executable extension (.exe, .cmd, .bat, etc)
If the kind CLI is already installed, but not on your path, you can set the environment variable KIND_PATH to specify the path to kind executable CLI.
Example:
Command Prompt(cmd):
	set KIND_PATH=C:\tools\kubernetes\kind.exe

PowerShell(ps):
	$env:KIND_PATH=C:\tools\kubernetes\kind.exe
`,
	"linux": `
To Install Kind CLI, Please visit https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries
If the kind CLI is already installed, but not on your path, you can set the environment variable KIND_PATH to specify the path to kind executable CLI
Example:
	export KIND_PATH=/home/user/tools/kubernetes/kind
`,
	"darwin": `
To Install Kind CLI, Please visit https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries
If the kind CLI is already installed, but not on your path, you can set the environment variable KIND_PATH to specify the path to kind executable CLI
Example:
	export KIND_PATH=/home/user/tools/kubernetes/kind
`,
}

var kubectlExecutableMessage = map[string]string{
	"windows": `
To Install KubeCTL CLI, Please visit https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries
Make sure the downloaded file has an executable extension (.exe, .cmd, .bat, etc)
If the KubeCTL CLI is already installed, but not on your path, you can set the environment variable KUBECTL_PATH to specify the path to KubeCTL executable CLI
Example:
Command Prompt(cmd):
	set KUBECTL_PATH=C:\tools\kubernetes\kubectl.exe

PowerShell(ps):
	$env:KUBECTL_PATH=C:\tools\kubernetes\kubectl.exe
`,
	"linux": `
To Install KubeCTL CLI, Please visit https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries
If the KubeCTL CLI is already installed, but not on your path, you can set the environment variable KUBECTL_PATH to specify the path to KubeCTL executable CLI
Example:
	export KUBECTL_PATH=/home/user/tools/kubernetes/kubectl
`,
	"darwin": `
To Install KubeCTL CLI, Please visit https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries
If the KubeCTL CLI is already installed, but not on your path, you can set the environment variable KUBECTL_PATH to specify the path to KubeCTL executable CLI
Example:
	export KUBECTL_PATH=/home/user/tools/kubernetes/kubectl
`,
}

var helmExecutableMessage = map[string]string{
	"windows": `
To Install Helm CLI, Please visit https://github.com/helm/helm/releases
If the Helm CLI is already installed, but not on your path, you can set the environment variable HELM_PATH to specify the path to helm executable CLI
Example:
Command Prompt(cmd):
	set HELM_PATH=C:\tools\kubernetes\helm.exe

PowerShell(ps):
	$env:HELM_PATH=C:\tools\kubernetes\helm.exe
`,
	"linux": `
To Install Helm CLI, Please visit https://github.com/helm/helm/releases
If the Helm CLI is already installed, but not on your path, you can set the environment variable HELM_PATH to specify the path to helm executable CLI
Example:
	export HELM_PATH=/home/user/tools/kubernetes/helm
`,
	"darwin": `
To Install Helm CLI, Please visit https://github.com/helm/helm/releases
If the Helm CLI is already installed, but not on your path, you can set the environment variable HELM_PATH to specify the path to helm executable CLI
Example:
	export HELM_PATH=/home/user/tools/kubernetes/helm
`,
}

var dockerExecutableMessage = map[string]string{
	"windows": `
To Install Docker, Please visit https://docs.docker.com/engine/install/
`,
	"linux": `
To Install Docker, Please visit https://docs.docker.com/engine/install/
`,
	"darwin": `
To Install Docker, Please visit https://docs.docker.com/engine/install/
`,
}
