package internal

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateServiceExportConfig(t *testing.T) {
	cluster := &Cluster{
		Name:           "controller",
		ContextName:    "controller-ctx",
		KubeConfigPath: "/path/config",
	}

	tests := []struct {
		name          string
		namespace     string
		filename      string
		mockApplyFile func(string, string, *Cluster)
		expectFatal   bool
		fatalContains string
	}{
		{
			name:      "create service export config",
			namespace: "test-namespace",
			filename:  "service-export.yaml",
			mockApplyFile: func(fileName, namespace string, c *Cluster) {
				assert.Equal(t, "service-export.yaml", fileName)
				assert.Equal(t, "test-namespace", namespace)
				assert.Equal(t, cluster, c)
			},
		},
		{
			name:      "create with different namespace",
			namespace: "custom-namespace",
			filename:  "custom-export.yaml",
			mockApplyFile: func(fileName, namespace string, c *Cluster) {
				assert.Equal(t, "custom-export.yaml", fileName)
				assert.Equal(t, "custom-namespace", namespace)
			},
		},
		{
			name:      "apply file fails",
			namespace: "test-namespace",
			filename:  "service-export.yaml",
			mockApplyFile: func(fileName, namespace string, c *Cluster) {
				util.Fatalf("Process failed %v", errors.New("apply failed"))
			},
			expectFatal:   true,
			fatalContains: "apply failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeOutput := util.Output.(*util.FakeOutput)

			applyCalled := false
			originalApply := applyFileFuncServiceExport
			applyFileFuncServiceExport = func(fileName, namespace string, cluster *Cluster) {
				applyCalled = true
				if tt.mockApplyFile != nil {
					tt.mockApplyFile(fileName, namespace, cluster)
				}
			}
			defer func() { applyFileFuncServiceExport = originalApply }()

			CreateServiceExportConfig(tt.namespace, cluster, tt.filename)

			assert.True(t, applyCalled, "ApplyFile should be called")

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
			}
		})
	}
}

func TestGetServiceExportConfig(t *testing.T) {
	cluster := &Cluster{Name: "controller", ContextName: "ctrl-ctx"}

	tests := []struct {
		name          string
		mockGet       func(string, string, string, *Cluster, string)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful get",
			mockGet: func(resourceType, resourceName, namespace string, c *Cluster, outputFormat string) {
				assert.Equal(t, ServiceExportConfigObject, resourceType)
				assert.Equal(t, "my-export", resourceName)
				assert.Equal(t, "test-namespace", namespace)
				assert.Equal(t, cluster, c)
				assert.Equal(t, "", outputFormat)
			},
		},
		{
			name: "get fails",
			mockGet: func(resourceType, resourceName, namespace string, c *Cluster, outputFormat string) {
				util.Fatalf("Process failed %v", errors.New("get failed"))
			},
			expectFatal:   true,
			fatalContains: "get failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeClock := util.SystemClock.(*util.FakeClock)
			fakeOutput := util.Output.(*util.FakeOutput)

			originalGet := getKubectlResourcesFuncServiceExport
			getKubectlResourcesFuncServiceExport = tt.mockGet
			defer func() { getKubectlResourcesFuncServiceExport = originalGet }()

			GetServiceExportConfig("my-export", "test-namespace", cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				require.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			}
		})
	}
}

func TestDeleteServiceExportConfig(t *testing.T) {
	cluster := &Cluster{Name: "controller"}

	tests := []struct {
		name          string
		mockDelete    func(string, string, string, *Cluster)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful delete",
			mockDelete: func(resourceType, resourceName, namespace string, c *Cluster) {
				assert.Equal(t, ServiceExportConfigObject, resourceType)
				assert.Equal(t, "my-export", resourceName)
				assert.Equal(t, "test-namespace", namespace)
			},
		},
		{
			name: "delete fails",
			mockDelete: func(resourceType, resourceName, namespace string, c *Cluster) {
				util.Fatalf("Process failed %v", errors.New("delete failed"))
			},
			expectFatal:   true,
			fatalContains: "delete failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeClock := util.SystemClock.(*util.FakeClock)
			fakeOutput := util.Output.(*util.FakeOutput)

			originalDelete := deleteKubectlResourcesFuncServiceExport
			deleteKubectlResourcesFuncServiceExport = tt.mockDelete
			defer func() { deleteKubectlResourcesFuncServiceExport = originalDelete }()

			DeleteServiceExportConfig("my-export", "test-namespace", cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				require.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			}
		})
	}
}

func TestEditServiceExportConfig(t *testing.T) {
	cluster := &Cluster{Name: "controller"}

	tests := []struct {
		name          string
		mockEdit      func(string, string, string, *Cluster)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful edit",
			mockEdit: func(resourceType, resourceName, namespace string, c *Cluster) {
				assert.Equal(t, ServiceExportConfigObject, resourceType)
				assert.Equal(t, "my-export", resourceName)
				assert.Equal(t, "test-namespace", namespace)
			},
		},
		{
			name: "edit fails",
			mockEdit: func(resourceType, resourceName, namespace string, c *Cluster) {
				util.Fatalf("Process failed %v", errors.New("edit failed"))
			},
			expectFatal:   true,
			fatalContains: "edit failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeClock := util.SystemClock.(*util.FakeClock)
			fakeOutput := util.Output.(*util.FakeOutput)

			originalEdit := editKubectlResourcesFuncServiceExport
			editKubectlResourcesFuncServiceExport = tt.mockEdit
			defer func() { editKubectlResourcesFuncServiceExport = originalEdit }()

			EditServiceExportConfig("my-export", "test-namespace", cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				require.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			}
		})
	}
}

func TestDescribeServiceExportConfig(t *testing.T) {
	cluster := &Cluster{Name: "controller"}

	tests := []struct {
		name          string
		mockDescribe  func(string, string, string, *Cluster)
		expectFatal   bool
		fatalContains string
	}{
		{
			name: "successful describe",
			mockDescribe: func(resourceType, resourceName, namespace string, c *Cluster) {
				assert.Equal(t, ServiceExportConfigObject, resourceType)
				assert.Equal(t, "my-export", resourceName)
				assert.Equal(t, "test-namespace", namespace)
			},
		},
		{
			name: "describe fails",
			mockDescribe: func(resourceType, resourceName, namespace string, c *Cluster) {
				util.Fatalf("Process failed %v", errors.New("describe failed"))
			},
			expectFatal:   true,
			fatalContains: "describe failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeClock := util.SystemClock.(*util.FakeClock)
			fakeOutput := util.Output.(*util.FakeOutput)

			originalDescribe := describeKubectlResourcesFuncServiceExport
			describeKubectlResourcesFuncServiceExport = tt.mockDescribe
			defer func() { describeKubectlResourcesFuncServiceExport = originalDescribe }()

			DescribeServiceExportConfig("my-export", "test-namespace", cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				require.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			}
		})
	}
}
