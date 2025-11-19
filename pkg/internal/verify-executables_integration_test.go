//go:build integration
// +build integration

package internal

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyExecutablesIntegration(t *testing.T) {
	requiredBinaries := []string{"kubectl", "helm"}
	for _, bin := range requiredBinaries {
		if !isBinaryAvailable(t, bin) {
			t.Skipf("Required binary '%s' not found in PATH, skipping suite", bin)
		}
	}

	tests := []struct {
		name            string
		config          *ConfigurationSpecs
		requiredForTest []string
		mockLookPath    func(string) (string, error)
		expectFatal     bool
		fatalContains   string
		validatePaths   func(*testing.T)
	}{
		{
			name: "Kind cluster requires all binaries",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ClusterType: "kind",
					},
				},
			},
			requiredForTest: []string{"kind", "docker"},
			expectFatal:     false,
			validatePaths: func(t *testing.T) {
				assert.Contains(t, util.ExecutablePaths, "kind")
				assert.Contains(t, util.ExecutablePaths, "kubectl")
				assert.Contains(t, util.ExecutablePaths, "docker")
				assert.Contains(t, util.ExecutablePaths, "helm")
			},
		},
		{
			name: "Non-kind cluster requires only kubectl and helm",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ClusterType: "eks",
					},
				},
			},
			requiredForTest: []string{},
			expectFatal:     false,
			validatePaths: func(t *testing.T) {
				assert.NotContains(t, util.ExecutablePaths, "kind")
				assert.NotContains(t, util.ExecutablePaths, "docker")
				assert.Contains(t, util.ExecutablePaths, "kubectl")
				assert.Contains(t, util.ExecutablePaths, "helm")
			},
		},
		{
			name: "Enterprise profile requires all binaries",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						Profile: ProfileEntDemo,
					},
				},
			},
			requiredForTest: []string{"kind", "docker"},
			expectFatal:     false,
			validatePaths: func(t *testing.T) {
				assert.Contains(t, util.ExecutablePaths, "kind")
				assert.Contains(t, util.ExecutablePaths, "kubectl")
				assert.Contains(t, util.ExecutablePaths, "docker")
				assert.Contains(t, util.ExecutablePaths, "helm")
			},
		},
		{
			name: "Missing kind binary triggers fatal error",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ClusterType: "kind",
					},
				},
			},
			requiredForTest: []string{"docker"},
			mockLookPath: func(file string) (string, error) {
				if file == "kind" {
					return "", errors.New("not found")
				}
				return exec.LookPath(file)
			},
			expectFatal:   true,
			fatalContains: "To Install Kind CLI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, bin := range tt.requiredForTest {
				if !isBinaryAvailable(t, bin) {
					t.Skipf("Skipping test case: Required binary '%s' not found", bin)
				}
			}

			originalExecutor := util.CommandExecutor
			originalLookPath := lookPathFunc

			cleanupEnv := util.NewTestEnvironment()

			util.CommandExecutor = originalExecutor

			if tt.mockLookPath != nil {
				lookPathFunc = tt.mockLookPath
			} else {
				lookPathFunc = exec.LookPath
			}

			defer func() {
				cleanupEnv()
				lookPathFunc = originalLookPath
			}()

			fakeOutput := util.Output.(*util.FakeOutput)
			util.ExecutablePaths = make(map[string]string)

			VerifyExecutables(tt.config)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fakeOutput.FatalCalls[0], tt.fatalContains)
			} else {
				require.Empty(t, fakeOutput.FatalCalls)
				if tt.validatePaths != nil {
					tt.validatePaths(t)
				}
			}
		})
	}
}
