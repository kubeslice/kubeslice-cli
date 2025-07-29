package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

var ExecutablePaths map[string]string

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
		Printf("%s Failed to run command: %s %v\nOutput: %s\nError: %s\nSuggestion: Please check if the command and its arguments are correct, and ensure all dependencies are installed.", Cross, cli, arg, outB.String(), errB.String())
	}
	return err
}

func RunCommandWithoutPrint(cli string, arg ...string) error {
	var outB, errB bytes.Buffer
	err := RunCommandCustomIO(cli, &outB, &errB, true, arg...)
	// if err != nil {
	// 	Printf("%s Failed to run command\nOutput: %s\nError: %s %v", Cross, outB.String(), errB.String(), err)
	// }
	return err
}

func RunCommandOnStdIO(cli string, arg ...string) error {
	return RunCommandCustomIO(cli, os.Stdout, os.Stderr, false, arg...)
}

func RunCommandCustomIO(cli string, stdout, stderr io.Writer, suppressPrint bool, arg ...string) error {
	cmdPath, ok := ExecutablePaths[cli]
	if !ok {
		Printf("%s Executable '%s' not found in configured paths. Please check your installation.", Cross, cli)
		return fmt.Errorf("executable '%s' not found", cli)
	}
	cmd := exec.Command(cmdPath, arg...)
	if !suppressPrint {
		Printf("%s Running command: %s", Run, cmd.String())
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil && !suppressPrint {
		Printf("%s Command failed: %s\nSuggestion: Verify the command syntax and your environment setup.", Cross, cmd.String())
	}
	return err
}
