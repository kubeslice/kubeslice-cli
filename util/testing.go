package util

import (
	"fmt"
	"io"
	"os"
	"time"
)

type FakeExecutor struct {
	ExecuteFunc           func(cli string, args ...string) error
	ExecuteWithOutputFunc func(cli string, stdout, stderr io.Writer, args ...string) error

	Calls []ExecutorCall
}

type ExecutorCall struct {
	CLI  string
	Args []string
}

func (f *FakeExecutor) Execute(cli string, args ...string) error {
	f.Calls = append(f.Calls, ExecutorCall{CLI: cli, Args: args})
	if f.ExecuteFunc != nil {
		return f.ExecuteFunc(cli, args...)
	}
	return nil
}

func (f *FakeExecutor) ExecuteWithOutput(cli string, stdout, stderr io.Writer, args ...string) error {
	f.Calls = append(f.Calls, ExecutorCall{CLI: cli, Args: args})
	if f.ExecuteWithOutputFunc != nil {
		return f.ExecuteWithOutputFunc(cli, stdout, stderr, args...)
	}
	return nil
}

type FakeFileSystem struct {
	StatFunc      func(name string) (os.FileInfo, error)
	CreateFunc    func(name string) (*os.File, error)
	WriteFileFunc func(filename string, data []byte, perm os.FileMode) error
	ReadFileFunc  func(filename string) ([]byte, error)
	MkdirAllFunc  func(path string, perm os.FileMode) error
	RemoveAllFunc func(path string) error

	WrittenFiles map[string][]byte
}

func NewFakeFileSystem() *FakeFileSystem {
	return &FakeFileSystem{
		WrittenFiles: make(map[string][]byte),
	}
}

func (f *FakeFileSystem) Stat(name string) (os.FileInfo, error) {
	if f.StatFunc != nil {
		return f.StatFunc(name)
	}
	return nil, os.ErrNotExist
}

func (f *FakeFileSystem) Create(name string) (*os.File, error) {
	if f.CreateFunc != nil {
		return f.CreateFunc(name)
	}
	return nil, nil
}

func (f *FakeFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	f.WrittenFiles[filename] = data
	if f.WriteFileFunc != nil {
		return f.WriteFileFunc(filename, data, perm)
	}
	return nil
}

func (f *FakeFileSystem) ReadFile(filename string) ([]byte, error) {
	if f.ReadFileFunc != nil {
		return f.ReadFileFunc(filename)
	}
	if data, ok := f.WrittenFiles[filename]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (f *FakeFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if f.MkdirAllFunc != nil {
		return f.MkdirAllFunc(path, perm)
	}
	return nil
}

func (f *FakeFileSystem) RemoveAll(path string) error {
	if f.RemoveAllFunc != nil {
		return f.RemoveAllFunc(path)
	}
	return nil
}

type FakeOutput struct {
	InfofFunc  func(format string, args ...interface{})
	FatalfFunc func(format string, args ...interface{})

	FatalCalls []string
	InfoCalls  []string
}

func (f *FakeOutput) Infof(format string, args ...interface{}) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	f.InfoCalls = append(f.InfoCalls, msg)

	if f.InfofFunc != nil {
		f.InfofFunc(format, args...)
	}
}

func (f *FakeOutput) Fatalf(format string, args ...interface{}) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	f.FatalCalls = append(f.FatalCalls, msg)

	if f.FatalfFunc != nil {
		f.FatalfFunc(format, args...)
	}
	// Don't call os.Exit in tests
}

type FakeClock struct {
	SleepFunc func(d time.Duration)
	NowFunc   func() time.Time

	SleepCalls []time.Duration

	// Current fake time
	FakeTime time.Time
}

func NewFakeClock(t time.Time) *FakeClock {
	return &FakeClock{
		FakeTime: t,
	}
}

func (f *FakeClock) Sleep(d time.Duration) {
	f.SleepCalls = append(f.SleepCalls, d)
	f.FakeTime = f.FakeTime.Add(d)
	if f.SleepFunc != nil {
		f.SleepFunc(d)
	}
}

func (f *FakeClock) Now() time.Time {
	if f.NowFunc != nil {
		return f.NowFunc()
	}
	return f.FakeTime
}

// NewTestEnvironment sets up fake implementations for testing and returns a cleanup function.
func NewTestEnvironment() (cleanup func()) {
	origExecutor := CommandExecutor
	origFS := FileSystem
	origOutput := Output
	origClock := SystemClock

	CommandExecutor = &FakeExecutor{}
	FileSystem = NewFakeFileSystem()
	Output = &FakeOutput{}
	SystemClock = NewFakeClock(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	return func() {
		CommandExecutor = origExecutor
		FileSystem = origFS
		Output = origOutput
		SystemClock = origClock
	}
}
