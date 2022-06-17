package internal

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/kubeslice/kubeslice-installer/util"
)

func VerifyExecutables() {
	util.Printf("Verifying Executables...")
	time.Sleep(200 * time.Millisecond)
	for key := range util.ExecutablePaths {
		time.Sleep(200 * time.Millisecond)
		verificationResult(verifyBinary(key), key)
	}

	time.Sleep(200 * time.Millisecond)
	util.Printf("All required executables were found\n")
}

func verifyBinary(name string) int {
	return _verifyBinary(name, strings.ToUpper(name)+"_PATH", util.ExecutableVerifyCommands[name])
}

func _verifyBinary(name, environmentVariable string, executable []string) int {
	cli := name
	if os.Getenv(environmentVariable) != "" {
		cli = strings.Trim(os.Getenv(environmentVariable), "\"")
	}
	path, err := exec.LookPath(cli)
	if err != nil || path == "" {
		return 1
	}
	if err = exec.Command(path, executable...).Run(); err != nil {
		return 2
	}
	util.ExecutablePaths[name] = path
	return 0
}

func executableDownloadMessage(executable string) string {
	switch executable {
	case "kind":
		return kindExecutableMessage[fmt.Sprintf("%s", runtime.GOOS)]
	case "kubectl":
		return kubectlExecutableMessage[fmt.Sprintf("%s", runtime.GOOS)]
	case "helm":
		return helmExecutableMessage[fmt.Sprintf("%s", runtime.GOOS)]
	case "docker":
		return dockerExecutableMessage[fmt.Sprintf("%s", runtime.GOOS)]
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
