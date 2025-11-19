package internal

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateImagePullSecretsValue tests the pure function for generating image pull secrets.
func TestGenerateImagePullSecretsValue(t *testing.T) {

	tests := []struct {
		name     string
		input    ImagePullSecrets
		expected string
	}{
		{
			name: "All fields provided",
			input: ImagePullSecrets{
				Registry: "my-registry.com",
				Username: "user",
				Password: "password123",
				Email:    "user@example.com",
			},
			expected: fmt.Sprintf(imagePullSecretsTemplate, "my-registry.com", "user", "password123", "email: user@example.com"),
		},
		{
			name: "Default registry used when registry is empty",
			input: ImagePullSecrets{
				Username: "user",
				Password: "password123",
				Email:    "user@example.com",
			},
			expected: fmt.Sprintf(imagePullSecretsTemplate, "https://index.docker.io/v1/", "user", "password123", "email: user@example.com"),
		},
		{
			name: "Email is optional",
			input: ImagePullSecrets{
				Registry: "my-registry.com",
				Username: "user",
				Password: "password123",
			},
			expected: fmt.Sprintf(imagePullSecretsTemplate, "my-registry.com", "user", "password123", ""),
		},
		{
			name: "Returns empty string if username is missing",
			input: ImagePullSecrets{
				Password: "password123",
			},
			expected: "",
		},
		{
			name: "Returns empty string if password is missing",
			input: ImagePullSecrets{
				Username: "user",
			},
			expected: "",
		},
		{
			name:     "Returns empty string for empty input struct",
			input:    ImagePullSecrets{},
			expected: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			got := generateImagePullSecretsValue(tc.input)

			assert.Equal(t,
				strings.TrimSpace(tc.expected),
				strings.TrimSpace(got),
				"generateImagePullSecretsValue() output mismatch")
		})
	}
}

// TestAddHelmCharts tests the main helm chart addition workflow.
func TestAddHelmCharts(t *testing.T) {

	tests := []struct {
		name          string
		config        *ConfigurationSpecs
		mockExecutor  func(*util.FakeExecutor)
		expectFatal   bool
		fatalContains string
		expectedCalls int
	}{
		{
			name: "UseLocal skips helm commands",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					HelmChartConfiguration: HelmChartConfiguration{
						UseLocal: true,
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					t.Error("Should not execute any commands when UseLocal is true")
					return nil
				}
			},
			expectedCalls: 0,
		},
		{
			name: "Successful helm repo add and update",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						RepoUrl:   "https://kubeslice.github.io/kubeslice/",
						UseLocal:  false,
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			expectedCalls: 2,
		},
		{
			name: "Helm repo add with credentials",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias:    "kubeslice-ent",
						RepoUrl:      "https://private.example.com/charts",
						HelmUsername: "user",
						HelmPassword: "pass",
						UseLocal:     false,
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					assert.Equal(t, "helm", cli, "Expected helm command")
					if len(args) > 1 && args[0] == "repo" && args[1] == "add" {
						argsStr := strings.Join(args, " ")
						assert.Contains(t, argsStr, "--username user", "Missing username")
						assert.Contains(t, argsStr, "--password pass", "Missing password")
						assert.Contains(t, argsStr, "--pass-credentials", "Missing pass-credentials flag")
					}
					return nil
				}
			},
			expectedCalls: 2,
		},
		{
			name: "Helm repo add fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						RepoUrl:   "https://kubeslice.github.io/kubeslice/",
						UseLocal:  false,
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					if len(args) > 1 && args[0] == "repo" && args[1] == "add" {
						return errors.New("repository not found")
					}
					return nil
				}
			},
			expectFatal:   true,
			fatalContains: "Process failed",
			expectedCalls: 1,
		},
		{
			name: "Helm repo update fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						RepoUrl:   "https://kubeslice.github.io/kubeslice/",
						UseLocal:  false,
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					if len(args) > 1 && args[0] == "repo" && args[1] == "update" {
						return errors.New("update failed")
					}
					return nil
				}
			},
			expectFatal:   true,
			fatalContains: "Process failed",
			expectedCalls: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeOutput := util.Output.(*util.FakeOutput)
			fakeClock := util.SystemClock.(*util.FakeClock)

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			AddHelmCharts(tt.config)

			if tt.expectFatal {
				assert.NotEmpty(t, fakeOutput.FatalCalls, "Expected fatal call but got none")
				if tt.fatalContains != "" {
					assert.Contains(t, fakeOutput.FatalCalls[0], tt.fatalContains,
						"Fatal message should contain expected text")
				}
			} else {
				assert.Empty(t, fakeOutput.FatalCalls, "Unexpected fatal calls")
			}

			if !tt.expectFatal {
				assert.Len(t, fakeExec.Calls, tt.expectedCalls,
					"Command execution count mismatch")
			}

			if !tt.expectFatal && !tt.config.Configuration.HelmChartConfiguration.UseLocal {
				assert.Len(t, fakeClock.SleepCalls, 2, "Expected 2 sleep calls")
				for i, d := range fakeClock.SleepCalls {
					assert.Equal(t, 200*time.Millisecond, d,
						"Sleep call %d should be 200ms", i)
				}
			}
		})
	}
}

// TestAddHelmChart tests the internal helm repo add function.
func TestAddHelmChart(t *testing.T) {

	tests := []struct {
		name          string
		config        *ConfigurationSpecs
		mockExecutor  func(*util.FakeExecutor)
		expectFatal   bool
		validateCalls func(*testing.T, []util.ExecutorCall)
	}{
		{
			name: "Basic repo add without credentials",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						RepoUrl:   "https://kubeslice.github.io/kubeslice/",
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			validateCalls: func(t *testing.T, calls []util.ExecutorCall) {
				require.Len(t, calls, 1, "Expected exactly 1 call")
				call := calls[0]
				assert.Equal(t, "helm", call.CLI, "Expected helm command")
				expectedArgs := []string{
					"repo", "add", "kubeslice",
					"https://kubeslice.github.io/kubeslice/",
					"--force-update",
				}
				assert.Equal(t, expectedArgs, call.Args, "Args mismatch")
			},
		},
		{
			name: "Repo add with credentials",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias:    "kubeslice",
						RepoUrl:      "https://kubeslice.github.io/kubeslice/",
						HelmUsername: "testuser",
						HelmPassword: "testpass",
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			validateCalls: func(t *testing.T, calls []util.ExecutorCall) {
				require.Len(t, calls, 1, "Expected exactly 1 call")
				args := calls[0].Args
				argsStr := strings.Join(args, " ")
				assert.Contains(t, argsStr, "--pass-credentials",
					"Missing --pass-credentials flag")
				assert.True(t, containsSequence(args, []string{"--username", "testuser"}),
					"Missing --username testuser")
				assert.True(t, containsSequence(args, []string{"--password", "testpass"}),
					"Missing --password testpass")
			},
		},
		{
			name: "Command fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice",
						RepoUrl:   "https://invalid.url",
					},
				},
			},
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("connection refused")
				}
			},
			expectFatal: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeOutput := util.Output.(*util.FakeOutput)

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			addHelmChart(tt.config)

			if tt.expectFatal {
				assert.NotEmpty(t, fakeOutput.FatalCalls, "Expected fatal call")
			} else {
				assert.Empty(t, fakeOutput.FatalCalls, "Unexpected fatal calls")
				if tt.validateCalls != nil {
					tt.validateCalls(t, fakeExec.Calls)
				}
			}
		})
	}
}

// TestUpdateHelmChart tests the helm repo update function.
func TestUpdateHelmChart(t *testing.T) {

	tests := []struct {
		name         string
		mockExecutor func(*util.FakeExecutor)
		expectFatal  bool
	}{
		{
			name: "Update succeeds",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return nil
				}
			},
			expectFatal: false,
		},
		{
			name: "Update fails",
			mockExecutor: func(fe *util.FakeExecutor) {
				fe.ExecuteFunc = func(cli string, args ...string) error {
					return errors.New("update failed")
				}
			},
			expectFatal: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeExec := util.CommandExecutor.(*util.FakeExecutor)
			fakeOutput := util.Output.(*util.FakeOutput)

			if tt.mockExecutor != nil {
				tt.mockExecutor(fakeExec)
			}

			updateHelmChart()

			if tt.expectFatal {
				assert.NotEmpty(t, fakeOutput.FatalCalls, "Expected fatal call")
			} else {
				assert.Empty(t, fakeOutput.FatalCalls, "Unexpected fatal calls")
				if assert.NotEmpty(t, fakeExec.Calls, "Expected at least one command execution") {
					call := fakeExec.Calls[0]
					assert.Equal(t, "helm", call.CLI, "Expected helm command")
					expectedArgs := []string{"repo", "update"}
					assert.Equal(t, expectedArgs, call.Args, "Args mismatch")
				}
			}
		})
	}
}

// This helper checks if a slice contains a specific sequence of elements in order
func containsSequence(slice, sequence []string) bool {
	if len(sequence) == 0 {
		return true
	}
	if len(slice) < len(sequence) {
		return false
	}
	for i := 0; i <= len(slice)-len(sequence); i++ {
		match := true
		for j := range sequence {
			if slice[i+j] != sequence[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
