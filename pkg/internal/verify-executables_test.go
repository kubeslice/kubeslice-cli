package internal

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyExecutables(t *testing.T) {
	tests := []struct {
		name                string
		config              *ConfigurationSpecs
		mockLookPath        func(string) (string, error)
		mockExecutor        func(*util.FakeExecutor)
		mockGetEnv          func(string) string
		expectedExecutables map[string]string
		expectFatal         bool
		fatalContains       string
	}{
		{
			name: "all executables found for kind cluster",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ClusterType: "kind",
					},
				},
			},
			mockLookPath: func(file string) (string, error) {
				return "/usr/local/bin/" + file, nil
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			mockGetEnv: func(key string) string {
				return ""
			},
			expectedExecutables: map[string]string{
				"kind":    "/usr/local/bin/kind",
				"kubectl": "/usr/local/bin/kubectl",
				"docker":  "/usr/local/bin/docker",
				"helm":    "/usr/local/bin/helm",
			},
		},
		{
			name: "all executables found for non-kind cluster",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ClusterType: "eks",
					},
				},
			},
			mockLookPath: func(file string) (string, error) {
				return "/usr/local/bin/" + file, nil
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			mockGetEnv: func(key string) string {
				return ""
			},
			expectedExecutables: map[string]string{
				"kubectl": "/usr/local/bin/kubectl",
				"helm":    "/usr/local/bin/helm",
			},
		},
		{
			name: "enterprise profile requires all executables",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						Profile: ProfileEntDemo,
					},
				},
			},
			mockLookPath: func(file string) (string, error) {
				return "/usr/local/bin/" + file, nil
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			mockGetEnv: func(key string) string {
				return ""
			},
			expectedExecutables: map[string]string{
				"kind":    "/usr/local/bin/kind",
				"kubectl": "/usr/local/bin/kubectl",
				"docker":  "/usr/local/bin/docker",
				"helm":    "/usr/local/bin/helm",
			},
		},
		{
			name: "kubectl not found",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ClusterType: "eks",
					},
				},
			},
			mockLookPath: func(file string) (string, error) {
				if file == "kubectl" {
					return "", errors.New("not found")
				}
				return "/usr/local/bin/" + file, nil
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			mockGetEnv: func(key string) string {
				return ""
			},
			expectFatal:   true,
			fatalContains: "not found on path",
		},
		{
			name: "helm not executable",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ClusterType: "eks",
					},
				},
			},
			mockLookPath: func(file string) (string, error) {
				return "/usr/local/bin/" + file, nil
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					if cli == "helm" {
						return errors.New("permission denied")
					}
					return nil
				}
			},
			mockGetEnv: func(key string) string {
				return ""
			},
			expectFatal:   true,
			fatalContains: "not executable",
		},
		{
			name: "docker not found for kind cluster",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ClusterType: "kind",
					},
				},
			},
			mockLookPath: func(file string) (string, error) {
				if file == "docker" {
					return "", errors.New("not found")
				}
				return "/usr/local/bin/" + file, nil
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			mockGetEnv: func(key string) string {
				return ""
			},
			expectFatal:   true,
			fatalContains: "not found on path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeOutput := util.Output.(*util.FakeOutput)

			originalLookPath := lookPathFunc
			lookPathFunc = tt.mockLookPath
			defer func() { lookPathFunc = originalLookPath }()

			originalGetEnv := getEnvFunc
			getEnvFunc = tt.mockGetEnv
			defer func() { getEnvFunc = originalGetEnv }()

			originalRuntimeGOOS := runtimeGOOSFunc
			runtimeGOOSFunc = func() string { return "linux" }
			defer func() { runtimeGOOSFunc = originalRuntimeGOOS }()

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			if util.ExecutableVerifyCommands == nil {
				util.ExecutableVerifyCommands = map[string][]string{
					"kubectl": {"version", "--client=true"},
					"helm":    {"version"},
					"kind":    {"version"},
					"docker":  {"version"},
				}
			}

			VerifyExecutables(tt.config)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls, "Expected fatal call")
				if tt.fatalContains != "" {
					found := false
					for _, call := range fakeOutput.FatalCalls {
						callStr := fmt.Sprint(call)
						if strings.Contains(callStr, tt.fatalContains) {
							found = true
							break
						}
					}
					if !found {
						for _, call := range fakeOutput.InfoCalls {
							callStr := fmt.Sprint(call)
							if strings.Contains(callStr, tt.fatalContains) {
								found = true
								break
							}
						}
					}
					assert.True(t, found, "Expected message containing: %s", tt.fatalContains)
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls, "Unexpected fatal calls")
				for name, expectedPath := range tt.expectedExecutables {
					path, exists := util.ExecutablePaths[name]
					assert.True(t, exists, "Expected %s to be in ExecutablePaths", name)
					assert.Equal(t, expectedPath, path, "Path mismatch for %s", name)
				}
			}
		})
	}
}

func TestVerifyBinaryWithEnvironmentVariable(t *testing.T) {
	tests := []struct {
		name         string
		binaryName   string
		envValue     string
		mockLookPath func(string) (string, error)
		mockExecutor func(*util.FakeExecutor)
		expectedCode int
	}{
		{
			name:       "use custom path from environment",
			binaryName: "kubectl",
			envValue:   "/custom/path/kubectl",
			mockLookPath: func(file string) (string, error) {
				if file == "/custom/path/kubectl" {
					return file, nil
				}
				return "", errors.New("not found")
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			expectedCode: 0,
		},
		{
			name:       "environment variable with quotes",
			binaryName: "helm",
			envValue:   "\"/custom/path/helm\"",
			mockLookPath: func(file string) (string, error) {
				if file == "/custom/path/helm" {
					return file, nil
				}
				return "", errors.New("not found")
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			expectedCode: 0,
		},
		{
			name:       "fallback to default when env not set",
			binaryName: "kubectl",
			envValue:   "",
			mockLookPath: func(file string) (string, error) {
				if file == "kubectl" {
					return "/usr/bin/kubectl", nil
				}
				return "", errors.New("not found")
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			expectedCode: 0,
		},
		{
			name:       "lookPath returns empty string",
			binaryName: "kubectl",
			envValue:   "",
			mockLookPath: func(file string) (string, error) {
				return "", nil
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			expectedCode: 1,
		},
		{
			name:       "binary not executable returns code 2",
			binaryName: "helm",
			envValue:   "",
			mockLookPath: func(file string) (string, error) {
				return "/usr/bin/helm", nil
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("permission denied")
				}
			},
			expectedCode: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)

			originalLookPath := lookPathFunc
			lookPathFunc = tt.mockLookPath
			defer func() { lookPathFunc = originalLookPath }()

			originalGetEnv := getEnvFunc
			getEnvFunc = func(key string) string {
				if key == "KUBECTL_PATH" || key == "HELM_PATH" {
					return tt.envValue
				}
				return ""
			}
			defer func() { getEnvFunc = originalGetEnv }()

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			if util.ExecutableVerifyCommands == nil {
				util.ExecutableVerifyCommands = map[string][]string{
					"kubectl": {"version", "--client=true"},
					"helm":    {"version"},
				}
			}

			code := verifyBinary(tt.binaryName)
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}
