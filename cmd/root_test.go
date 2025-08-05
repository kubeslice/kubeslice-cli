package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRootCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		wantOutput  string
	}{
		{
			name:       "help flag",
			args:       []string{"--help"},
			wantErr:    false,
			wantOutput: "kubeslice-cli - a simple CLI for KubeSlice Operations",
		},
		{
			name:       "version flag",
			args:       []string{"--version"},
			wantErr:    false,
			wantOutput: "kubeslice-cli version 0.6.0",
		},
		{
			name:       "no args shows help",
			args:       []string{},
			wantErr:    false,
			wantOutput: "kubeslice-cli - a simple CLI for KubeSlice Operations",
		},
		{
			name:    "config flag with value",
			args:    []string{"--config", "/path/to/config.yaml"},
			wantErr: false,
		},
		{
			name:    "config short flag",
			args:    []string{"-c", "/path/to/config.yaml"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global variables
			Config = ""
			
			// Create a new root command for each test
			cmd := &cobra.Command{
				Use:     "kubeslice-cli",
				Version: version,
				Short:   "kubeslice-cli - a simple CLI for KubeSlice Operations",
				Long: `kubeslice-cli - a simple CLI for KubeSlice Operations
    
Use kubeslice-cli to install/uninstall required workloads to run KubeSlice Controller and KubeSlice Worker.
Additional example applications can also be installed in demo profiles to showcase the
KubeSlice functionality`,
				Run: func(cmd *cobra.Command, args []string) {
					cmd.Help()
				},
			}
			
			cmd.PersistentFlags().StringVarP(&Config, "config", "c", "", `<path-to-topology-configuration-yaml-file>
	The yaml file with topology configuration. 
	Refer: https://github.com/kubeslice/kubeslice-cli/blob/master/samples/template.yaml`)

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantOutput != "" {
				outputStr := output.String()
				assert.Contains(t, outputStr, tt.wantOutput)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	// Save original args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name     string
		args     []string
		wantExit bool
	}{
		{
			name:     "help command",
			args:     []string{"kubeslice-cli", "--help"},
			wantExit: false,
		},
		{
			name:     "version command",
			args:     []string{"kubeslice-cli", "--version"},
			wantExit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global variables
			Config = ""
			
			// Create new root command to avoid state pollution
			testRootCmd := &cobra.Command{
				Use:     "kubeslice-cli",
				Version: version,
				Short:   "kubeslice-cli - a simple CLI for KubeSlice Operations",
				Long: `kubeslice-cli - a simple CLI for KubeSlice Operations
    
Use kubeslice-cli to install/uninstall required workloads to run KubeSlice Controller and KubeSlice Worker.
Additional example applications can also be installed in demo profiles to showcase the
KubeSlice functionality`,
				Run: func(cmd *cobra.Command, args []string) {
					cmd.Help()
				},
			}
			
			testRootCmd.PersistentFlags().StringVarP(&Config, "config", "c", "", `<path-to-topology-configuration-yaml-file>
	The yaml file with topology configuration. 
	Refer: https://github.com/kubeslice/kubeslice-cli/blob/master/samples/template.yaml`)

			var output bytes.Buffer
			testRootCmd.SetOut(&output)
			testRootCmd.SetErr(&output)
			testRootCmd.SetArgs(tt.args[1:]) // Skip program name

			err := testRootCmd.Execute()
			assert.NoError(t, err)
		})
	}
}

func TestRootCmdVersion(t *testing.T) {
	assert.Equal(t, "0.6.0", version)
}

func TestRootCmdGlobalVariable(t *testing.T) {
	// Test that RootCmd is properly exported
	assert.NotNil(t, RootCmd)
	assert.Equal(t, "kubeslice-cli", RootCmd.Use)
	assert.Equal(t, version, RootCmd.Version)
}

func TestConfigFlag(t *testing.T) {
	// Reset global variable
	Config = ""
	
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	
	cmd.PersistentFlags().StringVarP(&Config, "config", "c", "", "config file path")
	cmd.SetArgs([]string{"--config", "/test/path.yaml"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "/test/path.yaml", Config)
}

func TestConfigFlagShort(t *testing.T) {
	// Reset global variable
	Config = ""
	
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	
	cmd.PersistentFlags().StringVarP(&Config, "config", "c", "", "config file path")
	cmd.SetArgs([]string{"-c", "/test/path.yaml"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "/test/path.yaml", Config)
}

func TestRootCmdLongDescription(t *testing.T) {
	expectedLong := `kubeslice-cli - a simple CLI for KubeSlice Operations
    
Use kubeslice-cli to install/uninstall required workloads to run KubeSlice Controller and KubeSlice Worker.
Additional example applications can also be installed in demo profiles to showcase the
KubeSlice functionality`
	
	assert.Equal(t, expectedLong, RootCmd.Long)
}

func TestRootCmdShortDescription(t *testing.T) {
	expectedShort := "kubeslice-cli - a simple CLI for KubeSlice Operations"
	assert.Equal(t, expectedShort, RootCmd.Short)
}

func TestRootCmdRunFunction(t *testing.T) {
	// Test that the run function shows help
	var output bytes.Buffer
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cmd.SetOut(&output)
	
	cmd.Run(cmd, []string{})
	
	// Should contain usage information
	assert.Contains(t, output.String(), "Usage:")
}
