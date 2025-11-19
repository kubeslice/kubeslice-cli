package util

import (
	"io"
	"os"
)

var ExecutablePaths map[string]string

var ExecutableVerifyCommands = map[string][]string{
	"kind":    {"version"},
	"kubectl": {"version", "--client=true"},
	"docker":  {"ps", "-a"},
	"helm":    {"version"},
}

// RunCommand executes a command using the global CommandExecutor.
func RunCommand(cli string, arg ...string) error {
	return CommandExecutor.Execute(cli, arg...)
}

// RunCommandWithoutPrint executes a command without printing output.
func RunCommandWithoutPrint(cli string, arg ...string) error {
	return CommandExecutor.Execute(cli, arg...)
}

// RunCommandOnStdIO executes a command with stdout/stderr.
func RunCommandOnStdIO(cli string, arg ...string) error {
	return CommandExecutor.ExecuteWithOutput(cli, os.Stdout, os.Stderr, arg...)
}

// RunCommandCustomIO executes a command with custom IO writers.
func RunCommandCustomIO(cli string, stdout, stderr io.Writer, suppressPrint bool, arg ...string) error {
	return CommandExecutor.ExecuteWithOutput(cli, stdout, stderr, arg...)
}
