package util

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunCommand(t *testing.T) {
	originalPaths := ExecutablePaths
	ExecutablePaths = map[string]string{
		"go":   "go",
		"fake": "/nonexistent/path/fake",
	}
	defer func() { ExecutablePaths = originalPaths }()

	tests := []struct {
		name        string
		cli         string
		args        []string
		expectError bool
	}{
		{
			name:        "successful command",
			cli:         "go",
			args:        []string{"version"},
			expectError: false,
		},
		{
			name:        "failing command",
			cli:         "fake",
			args:        []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunCommand(tt.cli, tt.args...)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRunCommandWithoutPrint(t *testing.T) {
	originalPaths := ExecutablePaths
	ExecutablePaths = map[string]string{
		"go":   "go",
		"fake": "/nonexistent/path/fake",
	}
	defer func() { ExecutablePaths = originalPaths }()

	tests := []struct {
		name        string
		cli         string
		args        []string
		expectError bool
	}{
		{
			name:        "successful command without print",
			cli:         "go",
			args:        []string{"version"},
			expectError: false,
		},
		{
			name:        "failing command without print",
			cli:         "fake",
			args:        []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunCommandWithoutPrint(tt.cli, tt.args...)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRunCommandOnStdIO(t *testing.T) {
	originalPaths := ExecutablePaths
	ExecutablePaths = map[string]string{
		"go": "go",
	}
	defer func() { ExecutablePaths = originalPaths }()

	originalStdout := os.Stdout
	originalStderr := os.Stderr
	defer func() {
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	err := RunCommandOnStdIO("go", "version")
	w.Close()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	output := buf.String()
	if !strings.Contains(output, "go version") {
		t.Errorf("expected output to contain 'go version', got: %s", output)
	}
}

func TestRunCommandCustomIO(t *testing.T) {
	originalPaths := ExecutablePaths
	ExecutablePaths = map[string]string{
		"go": "go",
	}
	defer func() { ExecutablePaths = originalPaths }()

	tests := []struct {
		name          string
		cli           string
		args          []string
		suppressPrint bool
		expectError   bool
		expectOutput  string
	}{
		{
			name:          "go version command with print",
			cli:           "go",
			args:          []string{"version"},
			suppressPrint: false,
			expectError:   false,
			expectOutput:  "go version",
		},
		{
			name:          "go version command suppress print",
			cli:           "go",
			args:          []string{"version"},
			suppressPrint: true,
			expectError:   false,
			expectOutput:  "go version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			err := RunCommandCustomIO(tt.cli, &stdout, &stderr, tt.suppressPrint, tt.args...)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectOutput != "" && !strings.Contains(stdout.String(), tt.expectOutput) {
				t.Errorf("expected stdout to contain '%s', got: %s", tt.expectOutput, stdout.String())
			}
		})
	}
}

func TestExecutableVerifyCommands(t *testing.T) {
	expectedCommands := map[string][]string{
		"kind":    {"version"},
		"kubectl": {"version", "--client=true"},
		"docker":  {"ps", "-a"},
		"helm":    {"version"},
	}

	for tool, expectedArgs := range expectedCommands {
		actualArgs, exists := ExecutableVerifyCommands[tool]
		if !exists {
			t.Errorf("ExecutableVerifyCommands missing entry for %s", tool)
			continue
		}

		if len(actualArgs) != len(expectedArgs) {
			t.Errorf("ExecutableVerifyCommands[%s] has %d args, expected %d", tool, len(actualArgs), len(expectedArgs))
			continue
		}

		for i, expectedArg := range expectedArgs {
			if actualArgs[i] != expectedArg {
				t.Errorf("ExecutableVerifyCommands[%s][%d] = %s, expected %s", tool, i, actualArgs[i], expectedArg)
			}
		}
	}
}

func TestExecutablePathsInitialization(t *testing.T) {
	originalPaths := ExecutablePaths
	defer func() { ExecutablePaths = originalPaths }()

	testPaths := map[string]string{
		"test-tool": "/usr/bin/test-tool",
		"another":   "/bin/another",
	}

	ExecutablePaths = testPaths

	for tool, expectedPath := range testPaths {
		actualPath, exists := ExecutablePaths[tool]
		if !exists {
			t.Errorf("ExecutablePaths missing entry for %s", tool)
			continue
		}
		if actualPath != expectedPath {
			t.Errorf("ExecutablePaths[%s] = %s, expected %s", tool, actualPath, expectedPath)
		}
	}
}

func TestRunCommandCustomIOWithNilWriters(t *testing.T) {
	originalPaths := ExecutablePaths
	ExecutablePaths = map[string]string{
		"go": "go",
	}
	defer func() { ExecutablePaths = originalPaths }()

	err := RunCommandCustomIO("go", nil, nil, true, "version")
	if err != nil {
		t.Errorf("unexpected error with nil writers: %v", err)
	}
}

func TestRunCommandWithNonExistentExecutable(t *testing.T) {
	originalPaths := ExecutablePaths
	ExecutablePaths = map[string]string{
		"nonexistent": "/path/to/nonexistent/binary",
	}
	defer func() { ExecutablePaths = originalPaths }()

	err := RunCommand("nonexistent", "arg1")
	if err == nil {
		t.Errorf("expected error for nonexistent executable, but got none")
	}
}

func TestRunCommandCustomIOErrorHandling(t *testing.T) {
	originalPaths := ExecutablePaths
	ExecutablePaths = map[string]string{
		"go": "go",
	}
	defer func() { ExecutablePaths = originalPaths }()

	var stdout, stderr bytes.Buffer
	err := RunCommandCustomIO("go", &stdout, &stderr, true, "invalid-command-that-should-fail")

	if err == nil {
		t.Errorf("expected error from invalid go command, but got none")
	}
}
