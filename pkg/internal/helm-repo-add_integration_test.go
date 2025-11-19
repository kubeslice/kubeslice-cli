//go:build integration
// +build integration

package internal

import (
	"testing"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelmRepoAddIntegration(t *testing.T) {
	if !isBinaryAvailable(t, "helm") {
		t.Skip("helm binary not found in PATH")
	}

	tests := []struct {
		name           string
		config         *ConfigurationSpecs
		skipCleanup    bool
		expectFatal    bool
		validateResult func(*testing.T, string)
	}{
		{
			name: "Successfully add public helm repository",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice-itest-success",
						RepoUrl:   "https://kubeslice.github.io/kubeslice/",
						UseLocal:  false,
					},
				},
			},
			expectFatal: false,
			validateResult: func(t *testing.T, repoAlias string) {
				output := getHelmRepoList(t)
				require.Contains(t, output, repoAlias)
				require.Contains(t, output, "https://kubeslice.github.io/kubeslice/")
			},
		},
		{
			name: "UseLocal flag skips helm operations",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice-itest-skip",
						RepoUrl:   "https://example.com/charts",
						UseLocal:  true,
					},
				},
			},
			skipCleanup: true,
			expectFatal: false,
			validateResult: func(t *testing.T, repoAlias string) {
				output := getHelmRepoList(t)
				require.NotContains(t, output, repoAlias)
			},
		},
		{
			name: "Fails to add non-existent repository",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					HelmChartConfiguration: HelmChartConfiguration{
						RepoAlias: "kubeslice-itest-fail",
						RepoUrl:   "https://non-existent-repo-12345.example.com/charts",
						UseLocal:  false,
					},
				},
			},
			skipCleanup: true,
			expectFatal: true,
			validateResult: func(t *testing.T, repoAlias string) {
				output := getHelmRepoList(t)
				require.NotContains(t, output, repoAlias)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoAlias := tt.config.Configuration.HelmChartConfiguration.RepoAlias
			originalExecutor := util.CommandExecutor
			cleanupEnv := util.NewTestEnvironment()
			util.CommandExecutor = originalExecutor
			fakeOutput := util.Output.(*util.FakeOutput)
			defer cleanupEnv()
			if !tt.skipCleanup {
				t.Cleanup(func() {
					cleanupHelmRepo(t, repoAlias)
				})
			}
			AddHelmCharts(tt.config)
			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fakeOutput.FatalCalls[0], "Process failed")
			} else {
				require.Empty(t, fakeOutput.FatalCalls)
			}
			if tt.validateResult != nil {
				tt.validateResult(t, repoAlias)
			}
		})
	}
}
