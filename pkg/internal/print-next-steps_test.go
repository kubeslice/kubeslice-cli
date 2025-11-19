package internal

import (
	"testing"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintNextSteps(t *testing.T) {
	tests := []struct {
		name             string
		verificationOnly bool
		config           *ConfigurationSpecs
		expectedContains []string
		notExpected      []string
	}{
		{
			name:             "verification only - standard profile",
			verificationOnly: true,
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						Profile: "",
						WorkerClusters: []Cluster{
							{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path1"},
							{Name: "worker-2", ContextName: "w2-ctx", KubeConfigPath: "/path2"},
						},
					},
				},
			},
			expectedContains: []string{
				"KubeSlice Cluster Setup",
				"iPerf Connectivity",
				"--context=w2-ctx",
				"--kubeconfig=/path2",
				"iperf-server.iperf.svc.slice.local",
			},
			notExpected: []string{
				"create a Slice",
				"Kubeslice Manager UI",
			},
		},
		{
			name:             "verification only - enterprise profile",
			verificationOnly: true,
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						Profile: ProfileEntDemo,
						ControllerCluster: Cluster{
							Name:           "controller",
							ContextName:    "ctrl-ctx",
							KubeConfigPath: "/path/ctrl",
						},
						WorkerClusters: []Cluster{
							{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path1"},
							{Name: "worker-2", ContextName: "w2-ctx", KubeConfigPath: "/path2"},
						},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName: "test-project",
					},
				},
			},
			expectedContains: []string{
				"KubeSlice Enterprise Setup",
				"Kubeslice Manager UI",
				"following token",
			},
		},
		{
			name:             "namespace isolation steps",
			verificationOnly: false,
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:           "controller",
							ContextName:    "ctrl-ctx",
							KubeConfigPath: "/path/ctrl",
						},
						WorkerClusters: []Cluster{
							{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path1"},
							{Name: "worker-2", ContextName: "w2-ctx", KubeConfigPath: "/path2"},
						},
					},
				},
			},
			expectedContains: []string{
				"create a Slice",
				"slice propagation",
				"restart the iPerf deployment",
				"export the iPerf server",
				"--context=ctrl-ctx",
				"--context=w1-ctx",
				"--context=w2-ctx",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeOutput := util.Output.(*util.FakeOutput)
			if util.ExecutablePaths == nil {
				util.ExecutablePaths = make(map[string]string)
			}
			util.ExecutablePaths["kubectl"] = "kubectl"

			// Mock GetUIAdminToken and GetUIEndpoint for enterprise tests
			originalTokenFunc := getUIAdminTokenFunc
			getUIAdminTokenFunc = func(cluster *Cluster, user, project string) string {
				return "mock-token-12345"
			}
			defer func() { getUIAdminTokenFunc = originalTokenFunc }()

			originalEndpointFunc := getUIEndpointFunc
			getUIEndpointFunc = func(cluster *Cluster, profile string) string {
				return "http://mock-endpoint:8080"
			}
			defer func() { getUIEndpointFunc = originalEndpointFunc }()

			PrintNextSteps(tt.verificationOnly, tt.config)

			require.NotEmpty(t, fakeOutput.InfoCalls)

			allOutput := ""
			for _, call := range fakeOutput.InfoCalls {
				allOutput += call
			}

			for _, expected := range tt.expectedContains {
				assert.Contains(t, allOutput, expected, "Output should contain: %s", expected)
			}

			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, allOutput, notExpected, "Output should NOT contain: %s", notExpected)
			}
		})
	}
}

func TestPrintVerificationSteps(t *testing.T) {
	tests := []struct {
		name             string
		config           *ConfigurationSpecs
		mockToken        string
		mockEndpoint     string
		expectedContains []string
		notExpected      []string
	}{
		{
			name: "standard profile",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						Profile: "",
						WorkerClusters: []Cluster{
							{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path1"},
							{Name: "worker-2", ContextName: "w2-ctx", KubeConfigPath: "/path2"},
						},
					},
				},
			},
			expectedContains: []string{
				"KubeSlice Cluster Setup",
				"kubectl --context=w2-ctx --kubeconfig=/path2",
				"exec -it deploy/iperf-sleep",
				"iperf-server.iperf.svc.slice.local",
			},
			notExpected: []string{
				"Enterprise",
				"Manager UI",
				"token",
			},
		},
		{
			name: "enterprise profile",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					ClusterConfiguration: ClusterConfiguration{
						Profile: ProfileEntDemo,
						ControllerCluster: Cluster{
							Name:           "controller",
							ContextName:    "ctrl-ctx",
							KubeConfigPath: "/path/ctrl",
						},
						WorkerClusters: []Cluster{
							{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path1"},
							{Name: "worker-2", ContextName: "w2-ctx", KubeConfigPath: "/path2"},
						},
					},
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName: "test-project",
					},
				},
			},
			mockToken:    "enterprise-token-xyz",
			mockEndpoint: "https://enterprise.example.com",
			expectedContains: []string{
				"KubeSlice Enterprise Setup",
				"Kubeslice Manager UI",
				"https://enterprise.example.com",
				"enterprise-token-xyz",
				"kubectl --context=w2-ctx",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeOutput := util.Output.(*util.FakeOutput)
			if util.ExecutablePaths == nil {
				util.ExecutablePaths = make(map[string]string)
			}
			util.ExecutablePaths["kubectl"] = "kubectl"

			originalTokenFunc := getUIAdminTokenFunc
			getUIAdminTokenFunc = func(cluster *Cluster, user, project string) string {
				if tt.mockToken != "" {
					assert.Equal(t, "ctrl-ctx", cluster.ContextName)
					assert.Equal(t, "admin", user)
					assert.Equal(t, "test-project", project)
					return tt.mockToken
				}
				return ""
			}
			defer func() { getUIAdminTokenFunc = originalTokenFunc }()

			originalEndpointFunc := getUIEndpointFunc
			getUIEndpointFunc = func(cluster *Cluster, profile string) string {
				if tt.mockEndpoint != "" {
					assert.Equal(t, "ctrl-ctx", cluster.ContextName)
					assert.Equal(t, ProfileEntDemo, profile)
					return tt.mockEndpoint
				}
				return ""
			}
			defer func() { getUIEndpointFunc = originalEndpointFunc }()

			printVerificationSteps(tt.config)

			require.Len(t, fakeOutput.InfoCalls, 1)
			output := fakeOutput.InfoCalls[0]

			for _, expected := range tt.expectedContains {
				assert.Contains(t, output, expected)
			}

			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, output, notExpected)
			}
		})
	}
}

func TestPrintVerificationSteps_InsufficientClusters(t *testing.T) {
	cleanup := util.NewTestEnvironment()
	defer cleanup()
	fakeOutput := util.Output.(*util.FakeOutput)

	config := &ConfigurationSpecs{
		Configuration: Configuration{
			ClusterConfiguration: ClusterConfiguration{
				WorkerClusters: []Cluster{
					{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path1"},
					// Only 1 cluster - should not crash
				},
			},
		},
	}

	printVerificationSteps(config)
	require.NotEmpty(t, fakeOutput.InfoCalls)
	assert.Contains(t, fakeOutput.InfoCalls[0], "Error: At least 2 worker clusters required")
}

func TestPrintNamespaceIsolationSteps(t *testing.T) {
	cleanup := util.NewTestEnvironment()
	defer cleanup()

	fakeOutput := util.Output.(*util.FakeOutput)
	if util.ExecutablePaths == nil {
		util.ExecutablePaths = make(map[string]string)
	}
	util.ExecutablePaths["kubectl"] = "kubectl"

	config := &ConfigurationSpecs{
		Configuration: Configuration{
			ClusterConfiguration: ClusterConfiguration{
				ControllerCluster: Cluster{
					Name:           "controller",
					ContextName:    "ctrl-ctx",
					KubeConfigPath: "/path/ctrl",
				},
				WorkerClusters: []Cluster{
					{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path1"},
					{Name: "worker-2", ContextName: "w2-ctx", KubeConfigPath: "/path2"},
				},
			},
		},
	}

	printNamespaceIsolationSteps(config)

	require.Len(t, fakeOutput.InfoCalls, 1)
	output := fakeOutput.InfoCalls[0]

	expectedCommands := []string{
		"kubectl --context=w2-ctx --kubeconfig=/path2 exec -it deploy/iperf-sleep",
		"kubectl --context=ctrl-ctx --kubeconfig=/path/ctrl apply -f",
		"kubectl --context=w1-ctx --kubeconfig=/path1 get slice -n kubeslice-system",
		"kubectl --context=w2-ctx --kubeconfig=/path2 get slice -n kubeslice-system",
		"kubectl --context=w1-ctx --kubeconfig=/path1 rollout restart deployment/iperf-server -n iperf",
		"kubectl --context=w2-ctx --kubeconfig=/path2 rollout restart deployment/iperf-sleep -n iperf",
		"kubectl --context=w1-ctx --kubeconfig=/path1 apply -f",
	}

	for _, cmd := range expectedCommands {
		assert.Contains(t, output, cmd, "Output should contain command: %s", cmd)
	}

	expectedSections := []string{
		"create a Slice",
		"slice propagation",
		"restart the iPerf deployment",
		"export the iPerf server",
		"Verify the iPerf Connectivity Again",
	}

	for _, section := range expectedSections {
		assert.Contains(t, output, section)
	}
}

func TestPrintNamespaceIsolationSteps_InsufficientClusters(t *testing.T) {
	cleanup := util.NewTestEnvironment()
	defer cleanup()
	fakeOutput := util.Output.(*util.FakeOutput)

	config := &ConfigurationSpecs{
		Configuration: Configuration{
			ClusterConfiguration: ClusterConfiguration{
				WorkerClusters: []Cluster{
					{Name: "worker-1", ContextName: "w1-ctx", KubeConfigPath: "/path1"},
				},
			},
		},
	}

	printNamespaceIsolationSteps(config)

	require.NotEmpty(t, fakeOutput.InfoCalls)
	assert.Contains(t, fakeOutput.InfoCalls[0], "Error: At least 2 worker clusters required")
}

func TestBuildKubectlCommand(t *testing.T) {
	tests := []struct {
		name     string
		cluster  *Cluster
		args     []string
		expected string
	}{
		{
			name: "with cluster context",
			cluster: &Cluster{
				ContextName:    "test-ctx",
				KubeConfigPath: "/path/config",
			},
			args:     []string{"get", "pods", "-n", "default"},
			expected: "kubectl --context=test-ctx --kubeconfig=/path/config get pods -n default",
		},
		{
			name:     "without cluster context",
			cluster:  nil,
			args:     []string{"version"},
			expected: "kubectl version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if util.ExecutablePaths == nil {
				util.ExecutablePaths = make(map[string]string)
			}
			util.ExecutablePaths["kubectl"] = "kubectl"
			result := buildKubectlCommand(tt.cluster, tt.args...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
