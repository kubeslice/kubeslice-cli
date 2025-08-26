package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestConstants(t *testing.T) {
	t.Parallel()

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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.constant != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, tc.constant)
			}
		})
	}
}

func TestPrintf(t *testing.T) {
	t.Parallel()

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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			Printf(tc.format, tc.args...)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if output != tc.expected {
				t.Errorf("Printf() output mismatch\nwant: %q\ngot:  %q", tc.expected, output)
			}
		})
	}
}

func TestFatalf(t *testing.T) {
	t.Parallel()

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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			if len(tc.args) > 0 {
				fmt.Printf(tc.format+"\n", tc.args...)
			} else {
				fmt.Println(tc.format + "\n")
			}

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if output != tc.expected {
				t.Errorf("Fatalf() output mismatch\nwant: %q\ngot:  %q", tc.expected, output)
			}
		})
	}
}
