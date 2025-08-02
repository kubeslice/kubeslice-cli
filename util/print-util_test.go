package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

var stdoutMutex sync.Mutex

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"Cross constant", Cross, string(rune(0x274c))},
		{"Tick constant", Tick, string(rune(0x2714))},
		{"Wait constant", Wait, string(rune(0x267B))},
		{"Run constant", Run, string(rune(0x1F3C3))},
		{"Warn constant", Warn, string(rune(0x26A0))},
		{"Lock constant", Lock, string(rune(0x1F512))},
		{"Globe constant", Globe, string(rune(0x1F310))},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.constant != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, tc.constant)
			}
		})
	}
}

func captureOutput(t *testing.T, fn func()) string {
	stdoutMutex.Lock()
	defer stdoutMutex.Unlock()

	tmpFile, err := os.CreateTemp("", "test_output_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	oldStdout := os.Stdout
	os.Stdout = tmpFile

	fn()

	os.Stdout = oldStdout
	tmpFile.Close()

	output, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	return string(output)
}

func TestPrintf(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "Printf with no arguments",
			format:   "Hello World",
			args:     []interface{}{},
			expected: "Hello World\n",
		},
		{
			name:     "Printf with single argument",
			format:   "Hello %s",
			args:     []interface{}{"World"},
			expected: "Hello World\n",
		},
		{
			name:     "Printf with multiple arguments",
			format:   "Error %d: %s",
			args:     []interface{}{404, "not found"},
			expected: "Error 404: not found\n",
		},
		{
			name:     "Printf with empty string",
			format:   "",
			args:     []interface{}{},
			expected: "\n",
		},
		{
			name:     "Printf with unicode constants",
			format:   "Status: %s Success: %s",
			args:     []interface{}{Cross, Tick},
			expected: fmt.Sprintf("Status: %s Success: %s\n", Cross, Tick),
		},
		{
			name:     "Printf with nil args",
			format:   "Hello %s",
			args:     nil,
			expected: "Hello %s\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := captureOutput(t, func() {
				Printf(tc.format, tc.args...)
			})

			if output != tc.expected {
				t.Errorf("Printf() output mismatch\nwant: %q\ngot:  %q", tc.expected, output)
			}
		})
	}
}

func TestFatalf(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "Fatalf with no arguments",
			format:   "Error occurred",
			args:     []interface{}{},
			expected: "Error occurred\n\n",
		},
		{
			name:     "Fatalf with single argument",
			format:   "Error: %s",
			args:     []interface{}{"file not found"},
			expected: "Error: file not found\n",
		},
		{
			name:     "Fatalf with multiple arguments",
			format:   "Error %d: %s",
			args:     []interface{}{404, "not found"},
			expected: "Error 404: not found\n",
		},
		{
			name:     "Fatalf with empty string",
			format:   "",
			args:     []interface{}{},
			expected: "\n\n",
		},
		{
			name:     "Fatalf with nil args",
			format:   "Fatal error %s",
			args:     nil,
			expected: "Fatal error %s\n\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			if os.Getenv("BE_CRASHER") == "1" {
				Fatalf(tc.format, tc.args...)
				return
			}

			cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
			cmd.Env = append(os.Environ(), "BE_CRASHER=1")

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			if e, ok := err.(*exec.ExitError); ok && !e.Success() {
				output := stdout.String()
				if output != tc.expected {
					t.Errorf("Fatalf() output mismatch\nwant: %q\ngot:  %q", tc.expected, output)
				}
			} else {
				t.Errorf("Fatalf() should exit with non-zero code, got: %v", err)
			}
		})
	}
}
func TestPrintfBranches(t *testing.T) {

	t.Run("Printf with args", func(t *testing.T) {
		output := captureOutput(t, func() {
			Printf("Test %s", "message")
		})

		expected := "Test message\n"
		if output != expected {
			t.Errorf("Expected %q, got %q", expected, output)
		}
	})

	t.Run("Printf without args", func(t *testing.T) {
		output := captureOutput(t, func() {
			Printf("Test message")
		})

		expected := "Test message\n"
		if output != expected {
			t.Errorf("Expected %q, got %q", expected, output)
		}
	})
}

// edge Case tests(additional) .
func TestEdgeCases(t *testing.T) {
	edgeCases := []struct {
		name   string
		format string
		args   []interface{}
	}{
		{"Empty args slice", "test", []interface{}{}},
		{"Single nil arg", "test %v", []interface{}{nil}},
		{"Mixed types", "int: %d, string: %s, bool: %t", []interface{}{42, "hello", true}},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			output := captureOutput(t, func() {
				Printf(tc.format, tc.args...)
			})

			if len(output) == 0 {
				t.Error("Expected some output, got empty string")
			}
			if !strings.HasSuffix(output, "\n") {
				t.Error("Expected output to end with newline")
			}
		})
	}
}
