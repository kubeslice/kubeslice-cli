package util

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)


func createFakeCmd(t *testing.T, script string) string {
	dir := t.TempDir()
	name := "fakecmd"

	if runtime.GOOS == "windows" {
		name += ".bat"
		script = "@echo off\r\n" + script
	} else {
		script = "#!/bin/sh\n" + script
	}

	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(script), 0755)
	if err != nil {
		t.Fatal(err)
	}
	return path
}


func TestRunCommandCustomIOSuccess(t *testing.T) {
	fake := createFakeCmd(t, "echo hello")
	ExecutablePaths = map[string]string{"test": fake}

	var out bytes.Buffer
	var errOut bytes.Buffer

	err := RunCommandCustomIO("test", &out, &errOut, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRunCommandFailure(t *testing.T) {
	fake := createFakeCmd(t, "exit 1")
	ExecutablePaths = map[string]string{"test": fake}

	err := RunCommand("test")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestRunCommandWithoutPrint(t *testing.T) {
	fake := createFakeCmd(t, "echo silent")
	ExecutablePaths = map[string]string{"test": fake}

	err := RunCommandWithoutPrint("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMissingExecutablePath(t *testing.T) {
	ExecutablePaths = map[string]string{}

	err := RunCommand("missing")
	if err == nil {
		t.Fatalf("expected error for missing executable path")
	}
}
