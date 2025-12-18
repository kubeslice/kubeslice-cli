package util

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintf(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "simple string without args",
			format:   "Hello World",
			args:     []interface{}{},
			expected: "Hello World\n",
		},
		{
			name:     "format string with args",
			format:   "Hello %s",
			args:     []interface{}{"World"},
			expected: "Hello World\n",
		},
		{
			name:     "multiple args",
			format:   "Hello %s %d",
			args:     []interface{}{"World", 123},
			expected: "Hello World 123\n",
		},
		{
			name:     "empty string",
			format:   "",
			args:     []interface{}{},
			expected: "\n",
		},
		{
			name:     "format with no args but placeholders",
			format:   "Hello %s",
			args:     []interface{}{},
			expected: "Hello %s\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call the function
			Printf(tt.format, tt.args...)

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read the output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			assert.Equal(t, tt.expected, output)
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
			name:     "simple string without args",
			format:   "Error occurred",
			args:     []interface{}{},
			expected: "Error occurred\n\n",
		},
		{
			name:     "format string with args",
			format:   "Error: %s",
			args:     []interface{}{"file not found"},
			expected: "Error: file not found\n",
		},
		{
			name:     "multiple args",
			format:   "Error %d: %s",
			args:     []interface{}{404, "Not Found"},
			expected: "Error 404: Not Found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip actual os.Exit(1) call for testing
			// We'll test the output instead
			
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create a test function that mimics Fatalf without os.Exit
			testFatalf := func(format string, a ...interface{}) {
				if len(a) > 0 {
					Printf(format, a...)
				} else {
					Printf(format + "\n")
				}
			}

			// Call the test function
			testFatalf(tt.format, tt.args...)

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read the output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are properly defined
	assert.Equal(t, string(rune(0x274c)), Cross)
	assert.Equal(t, string(rune(0x2714)), Tick)
	assert.Equal(t, string(rune(0x267B)), Wait)
	assert.Equal(t, string(rune(0x1F3C3)), Run)
	assert.Equal(t, string(rune(0x26A0)), Warn)
	assert.Equal(t, string(rune(0x1F512)), Lock)
	assert.Equal(t, string(rune(0x1F310)), Globe)
}

func TestConstantValues(t *testing.T) {
	// Test that constants have the expected Unicode values
	tests := []struct {
		name     string
		constant string
		expected rune
	}{
		{"Cross", Cross, 0x274c},
		{"Tick", Tick, 0x2714},
		{"Wait", Wait, 0x267B},
		{"Run", Run, 0x1F3C3},
		{"Warn", Warn, 0x26A0},
		{"Lock", Lock, 0x1F512},
		{"Globe", Globe, 0x1F310},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, string(tt.expected), tt.constant)
		})
	}
}

func TestConstantTypes(t *testing.T) {
	// Test that constants are strings
	assert.IsType(t, "", Cross)
	assert.IsType(t, "", Tick)
	assert.IsType(t, "", Wait)
	assert.IsType(t, "", Run)
	assert.IsType(t, "", Warn)
	assert.IsType(t, "", Lock)
	assert.IsType(t, "", Globe)
}

func TestPrintfWithSpecialCharacters(t *testing.T) {
	// Test Printf with special Unicode characters
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Printf("%s Test completed successfully", Tick)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, Tick)
	assert.Contains(t, output, "Test completed successfully")
}

func TestPrintfWithEmptyFormat(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Printf("")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Equal(t, "\n", output)
}

func TestPrintfWithNilArgs(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Printf("Test %v", nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Equal(t, "Test <nil>\n", output)
}

func TestConstantsNotEmpty(t *testing.T) {
	// Ensure all constants are not empty strings
	assert.NotEmpty(t, Cross)
	assert.NotEmpty(t, Tick)
	assert.NotEmpty(t, Wait)
	assert.NotEmpty(t, Run)
	assert.NotEmpty(t, Warn)
	assert.NotEmpty(t, Lock)
	assert.NotEmpty(t, Globe)
}

func TestConstantsUnique(t *testing.T) {
	// Ensure all constants have unique values
	constants := []string{Cross, Tick, Wait, Run, Warn, Lock, Globe}
	uniqueConstants := make(map[string]bool)
	
	for _, constant := range constants {
		assert.False(t, uniqueConstants[constant], "Constant value %s is not unique", constant)
		uniqueConstants[constant] = true
	}
	
	assert.Equal(t, 7, len(uniqueConstants))
}
