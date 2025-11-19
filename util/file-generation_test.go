package util

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDirectoryPath(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		statErr       error
		mkdirErr      error
		expectFatal   bool
		fatalContains string
	}{
		{
			name:    "directory already exists",
			path:    "/existing/path",
			statErr: nil,
		},
		{
			name:    "directory does not exist - creates successfully",
			path:    "/new/path",
			statErr: os.ErrNotExist,
		},
		{
			name:          "mkdir fails - permission denied",
			path:          "/restricted/path",
			statErr:       os.ErrNotExist,
			mkdirErr:      os.ErrPermission,
			expectFatal:   true,
			fatalContains: "Failed to create kubeslice directory",
		},
		{
			name:          "mkdir fails - disk full",
			path:          "/new/path",
			statErr:       os.ErrNotExist,
			mkdirErr:      errors.New("no space left on device"),
			expectFatal:   true,
			fatalContains: "Failed to create kubeslice directory",
		},
		{
			name:    "stat returns non-ErrNotExist error - no mkdir attempted",
			path:    "/some/path",
			statErr: errors.New("disk error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := NewTestEnvironment()

			fakeFS := FileSystem.(*FakeFileSystem)
			fakeOutput := Output.(*FakeOutput)

			mkdirCalled := false
			fakeFS.StatFunc = func(name string) (os.FileInfo, error) {
				assert.Equal(t, tt.path, name)
				return nil, tt.statErr
			}

			fakeFS.MkdirAllFunc = func(path string, perm os.FileMode) error {
				mkdirCalled = true
				assert.Equal(t, tt.path, path)
				assert.Equal(t, os.ModePerm, perm)
				return tt.mkdirErr
			}

			CreateDirectoryPath(tt.path)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls, "Expected Fatalf to be called")
				assert.Contains(t, fakeOutput.FatalCalls[0], tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls, "Did not expect Fatalf to be called")
			}

			if tt.statErr == os.ErrNotExist {
				assert.True(t, mkdirCalled, "Expected MkdirAll to be called")
			} else if tt.statErr != nil {
				assert.False(t, mkdirCalled, "Did not expect MkdirAll to be called when Stat returns non-ErrNotExist error")
			}

			cleanup()
		})
	}
}

func TestDumpFile(t *testing.T) {
	tests := []struct {
		name          string
		template      string
		filename      string
		writeErr      error
		expectFatal   bool
		fatalContains string
	}{
		{
			name:     "writes simple text successfully",
			template: "Hello, World!",
			filename: "/test/file.txt",
		},
		{
			name: "writes multi-line YAML successfully",
			template: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config`,
			filename: "/test/config.yaml",
		},
		{
			name:     "writes empty content successfully",
			template: "",
			filename: "/test/empty.txt",
		},
		{
			name: "writes JSON content successfully",
			template: `{
  "name": "test",
  "version": "1.0.0"
}`,
			filename: "/test/config.json",
		},
		{
			name:     "writes content with special characters",
			template: "Special: @#$%^&*()_+-=[]{}|;':\",./<>?",
			filename: "/test/special.txt",
		},
		{
			name:          "write fails - permission denied",
			template:      "test content",
			filename:      "/restricted/file.txt",
			writeErr:      os.ErrPermission,
			expectFatal:   true,
			fatalContains: "Failed to write /restricted/file.txt",
		},
		{
			name:          "write fails - disk full",
			template:      "test content",
			filename:      "/test/file.txt",
			writeErr:      errors.New("no space left on device"),
			expectFatal:   true,
			fatalContains: "Failed to write",
		},
		{
			name:          "write fails - invalid path",
			template:      "test content",
			filename:      "/invalid\x00path/file.txt",
			writeErr:      errors.New("invalid argument"),
			expectFatal:   true,
			fatalContains: "Failed to write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := NewTestEnvironment()

			fakeFS := FileSystem.(*FakeFileSystem)
			fakeOutput := Output.(*FakeOutput)

			fakeFS.WriteFileFunc = func(filename string, data []byte, perm os.FileMode) error {
				assert.Equal(t, tt.filename, filename)
				assert.Equal(t, []byte(tt.template), data)
				assert.Equal(t, os.FileMode(0644), perm)

				if tt.writeErr != nil {
					return tt.writeErr
				}

				fakeFS.WrittenFiles[filename] = data
				return nil
			}

			DumpFile(tt.template, tt.filename)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls, "Expected Fatalf to be called")
				assert.Contains(t, fakeOutput.FatalCalls[0], tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls, "Did not expect Fatalf to be called")

				content, exists := fakeFS.WrittenFiles[tt.filename]
				require.True(t, exists, "File should have been written")
				assert.Equal(t, tt.template, string(content))
			}

			cleanup()
		})
	}
}
