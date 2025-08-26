package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDirectoryPath(t *testing.T) {
	t.Parallel()

	testDir := filepath.Join(os.TempDir(), "kubeslice-test-"+t.Name())
	t.Cleanup(func() {
		os.RemoveAll(testDir)
	})

	tests := []struct {
		name string
		path string
	}{
		{
			name: "Create simple directory",
			path: "simple-dir",
		},
		{
			name: "Create nested directories",
			path: filepath.Join("nested", "deep", "deeper", "deepest"),
		},
		{
			name: "Create directory with special characters",
			path: "special-chars-dir_123",
		},
		{
			name: "Handle directory that already exists",
			path: "existing-dir",
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			targetPath := filepath.Join(testDir, tc.path)

			if tc.name == "Handle directory that already exists" {
				if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
					t.Fatalf("Failed to setup test: %v", err)
				}
			}

			CreateDirectoryPath(targetPath)

			info, err := os.Stat(targetPath)
			if err != nil {
				t.Errorf("CreateDirectoryPath() failed to create directory: %v", err)
				return
			}
			if !info.IsDir() {
				t.Errorf("CreateDirectoryPath() created a file instead of directory")
			}
		})
	}
}

func TestDumpFile(t *testing.T) {
	t.Parallel()

	testDir := filepath.Join(os.TempDir(), "kubeslice-test-"+t.Name())
	if err := os.MkdirAll(testDir, os.ModePerm); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(testDir)
	})

	tests := []struct {
		name     string
		content  string
		filename string
	}{
		{
			name:     "Write simple text file",
			content:  "Hello, World!",
			filename: "test-simple.txt",
		},
		{
			name: "Write multi-line content",
			content: `Line 1
Line 2
Line 3`,
			filename: "test-multiline.txt",
		},
		{
			name: "Write YAML content",
			content: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key1: value1
  key2: value2`,
			filename: "test-config.yaml",
		},
		{
			name:     "Write empty content",
			content:  "",
			filename: "test-empty.txt",
		},
		{
			name: "Write JSON content",
			content: `{
  "name": "test",
  "version": "1.0.0",
  "description": "test file"
}`,
			filename: "test-config.json",
		},
		{
			name:     "Overwrite existing file",
			content:  "New content",
			filename: "test-overwrite.txt",
		},
		{
			name:     "Write to nested directory",
			content:  "Nested content",
			filename: filepath.Join("nested", "dir", "test-nested.txt"),
		},
		{
			name:     "Write file with special characters in content",
			content:  "Special chars: @#$%^&*()_+-=[]{}|;':\",./<>?",
			filename: "test-special-chars.txt",
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			targetFile := filepath.Join(testDir, tc.filename)

			if tc.name == "Overwrite existing file" {
				if err := os.MkdirAll(filepath.Dir(targetFile), os.ModePerm); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(targetFile, []byte("Old content"), 0644); err != nil {
					t.Fatalf("Failed to setup test: %v", err)
				}
			}

			if tc.name == "Write to nested directory" {
				if err := os.MkdirAll(filepath.Dir(targetFile), os.ModePerm); err != nil {
					t.Fatalf("Failed to create nested directory: %v", err)
				}
			}

			DumpFile(tc.content, targetFile)

			actualContent, err := os.ReadFile(targetFile)
			if err != nil {
				t.Errorf("DumpFile() failed to create file: %v", err)
				return
			}
			if string(actualContent) != tc.content {
				t.Errorf("DumpFile() content mismatch\nwant: %q\ngot:  %q", tc.content, string(actualContent))
			}

			info, err := os.Stat(targetFile)
			if err == nil && info.IsDir() {
				t.Errorf("DumpFile() created a directory instead of a file")
			}
		})
	}
}

func TestDumpFile_CreatesFileWithCorrectPermissions(t *testing.T) {
	t.Parallel()

	testDir := filepath.Join(os.TempDir(), "kubeslice-test-"+t.Name())
	if err := os.MkdirAll(testDir, os.ModePerm); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(testDir)
	})

	targetFile := filepath.Join(testDir, "test-permissions.txt")

	DumpFile("test content", targetFile)

	info, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	mode := info.Mode()
	if mode&0600 == 0 {
		t.Errorf("File should be readable and writable by owner, got mode: %v", mode)
	}
}
