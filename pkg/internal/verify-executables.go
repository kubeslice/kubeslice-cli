package internal

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

var (
	lookPathFunc    = exec.LookPath
	getEnvFunc      = os.Getenv
	runtimeGOOSFunc = func() string { return runtime.GOOS }
)

func VerifyExecutables(ApplicationConfiguration *ConfigurationSpecs) {
	util.Printf("Verifying Executables...")
	util.Sleep(200 * time.Millisecond)

	if ApplicationConfiguration.Configuration.ClusterConfiguration.Profile != "" || ApplicationConfiguration.Configuration.ClusterConfiguration.ClusterType == "kind" {
		util.ExecutablePaths = map[string]string{
			"kind":    "kind",
			"kubectl": "kubectl",
			"docker":  "docker",
			"helm":    "helm",
		}
	} else {
		util.ExecutablePaths = map[string]string{
			"kubectl": "kubectl",
			"helm":    "helm",
		}
	}
	for key := range util.ExecutablePaths {
		util.Sleep(200 * time.Millisecond)
		verificationResult(verifyBinary(key), key)
	}

	util.Sleep(200 * time.Millisecond)
	util.Printf("All required executables were found\n")
}

func verifyBinary(name string) int {
	return _verifyBinary(name, strings.ToUpper(name)+"_PATH", util.ExecutableVerifyCommands[name])
}

func _verifyBinary(name, environmentVariable string, executable []string) int {
	cli := name
	if envPath := getEnvFunc(environmentVariable); envPath != "" {
		cli = strings.Trim(envPath, "\"")
	}

	path, err := lookPathFunc(cli)
	if err != nil || path == "" {
		return 1
	}
	args := append([]string{}, executable...)
	if err = util.CommandExecutor.Execute(name, args...); err != nil {
		return 2
	}
	util.ExecutablePaths[name] = path
	return 0
}

func executableDownloadMessage(executable string) string {
	goos := runtimeGOOSFunc()
	switch executable {
	case "kind":
		return kindExecutableMessage[goos]
	case "kubectl":
		return kubectlExecutableMessage[goos]
	case "helm":
		return helmExecutableMessage[goos]
	case "docker":
		return dockerExecutableMessage[goos]
	}
	return ""
}

func verificationResult(num int, cli string) {
	switch num {
	case 0:
		util.Printf("%s %s found", util.Tick, cli)
	case 1:
		util.Printf("%s %s not found on path", util.Cross, cli)
		util.Fatalf(executableDownloadMessage(cli))
	case 2:
		util.Fatalf("%s %s is not executable", util.Cross, cli)
	}
}
