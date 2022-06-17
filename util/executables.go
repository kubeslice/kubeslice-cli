package util

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

var ExecutablePaths = map[string]string{
	"kind":    "kind",
	"kubectl": "kubectl",
	"docker":  "docker",
	"helm":    "helm",
}
var ExecutableVerifyCommands = map[string][]string{
	"kind":    {"version"},
	"kubectl": {"version", "--client=true"},
	"docker":  {"ps", "-a"},
	"helm":    {"version"},
}

func RunCommand(cli string, arg ...string) error {
	var outB, errB bytes.Buffer
	err := RunCommandCustomIO(cli, &outB, &errB, false, arg...)
	if err != nil {
		Printf("%s Failed to run command\nOutput: %s\nError: %s %v", Cross, outB.String(), errB.String(), err)
	}
	return err
}

func RunCommandOnStdIO(cli string, arg ...string) error {
	return RunCommandCustomIO(cli, os.Stdout, os.Stderr, false, arg...)
}

func RunCommandCustomIO(cli string, stdout, stderr io.Writer, suppressPrint bool, arg ...string) error {
	cmd := exec.Command(ExecutablePaths[cli], arg...)
	if !suppressPrint {
		Printf("%s Running command: %s", Run, cmd.String())
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
