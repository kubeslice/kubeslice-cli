package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// defaultExecutor implements Executor using actual exec.Command.
type defaultExecutor struct{}

func (e *defaultExecutor) Execute(cli string, args ...string) error {
	var outB, errB bytes.Buffer
	err := e.ExecuteWithOutput(cli, &outB, &errB, args...)
	if err != nil {
		Printf("%s Failed to run command\nOutput: %s\nError: %s %v", Cross, outB.String(), errB.String(), err)
	}
	return err
}

func (e *defaultExecutor) ExecuteWithOutput(cli string, stdout, stderr io.Writer, args ...string) error {
	path, exists := ExecutablePaths[cli]
	if !exists {
		return fmt.Errorf("executable %s not found in path", cli)
	}

	cmd := exec.Command(path, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	Printf("%s Running command: %s", Run, cmd.String())
	return cmd.Run()
}

// defaultFileSystem implements FileOperations using actual os operations.
type defaultFileSystem struct{}

func (fs *defaultFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (fs *defaultFileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func (fs *defaultFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (fs *defaultFileSystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (fs *defaultFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fs *defaultFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// defaultOutput implements OutputWriter using fmt and os operations.
type defaultOutput struct{}

func (o *defaultOutput) Infof(format string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf(format+"\n", args...)
	} else {
		fmt.Println(format)
	}
}

func (o *defaultOutput) Fatalf(format string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf(format+"\n", args...)
	} else {
		fmt.Println(format + "\n")
	}
	os.Exit(1)
}

// defaultClock implements Clock using actual time operations.
type defaultClock struct{}

func (c *defaultClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (c *defaultClock) Now() time.Time {
	return time.Now()
}
