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

func TestCreateKubeSliceProject(t *testing.T) {
	tests := []struct {
		name              string
		config            *ConfigurationSpecs
		cliOptions        *CliOptionsStruct
		mockApplyManifest func(string, string, *Cluster)
		applyShouldFail   bool
		expectedFileName  string
		expectedNamespace string
		expectedUsers     []string
	}{
		{
			name: "create project with default options",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName:  "test-project",
						ProjectUsers: []string{"user1", "user2"},
					},
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{
							Name:        "controller",
							ContextName: "controller-ctx",
						},
					},
				},
			},
			cliOptions: nil,
			mockApplyManifest: func(fileName, namespace string, cluster *Cluster) {
				assert.Contains(t, fileName, projectFileName)
				assert.Equal(t, KUBESLICE_CONTROLLER_NAMESPACE, namespace)
				assert.Equal(t, "controller", cluster.Name)
			},
			expectedFileName:  kubesliceDirectory + "/" + projectFileName,
			expectedNamespace: KUBESLICE_CONTROLLER_NAMESPACE,
			expectedUsers:     []string{"user1", "user2"},
		},
		{
			name: "create project with custom CLI options",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName:  "custom-project",
						ProjectUsers: []string{"admin"},
					},
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
					},
				},
			},
			cliOptions: &CliOptionsStruct{
				FileName:  "custom-project.yaml",
				Namespace: "custom-namespace",
				Cluster: &Cluster{
					Name:        "custom-cluster",
					ContextName: "custom-ctx",
				},
			},
			mockApplyManifest: func(fileName, namespace string, cluster *Cluster) {
				assert.Equal(t, "custom-project.yaml", fileName)
				assert.Equal(t, "custom-namespace", namespace)
				assert.Equal(t, "custom-cluster", cluster.Name)
			},
			expectedFileName:  "custom-project.yaml",
			expectedNamespace: "custom-namespace",
			expectedUsers:     []string{"admin"},
		},
		{
			name: "create project with empty filename in CLI options",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName:  "test-project",
						ProjectUsers: []string{},
					},
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
					},
				},
			},
			cliOptions: &CliOptionsStruct{
				FileName:  "",
				Namespace: "test-namespace",
				Cluster: &Cluster{
					Name:        "test-cluster",
					ContextName: "test-ctx",
				},
			},
			mockApplyManifest: func(fileName, namespace string, cluster *Cluster) {
				assert.Contains(t, fileName, projectFileName)
				assert.Equal(t, "test-namespace", namespace)
			},
			expectedFileName:  kubesliceDirectory + "/" + projectFileName,
			expectedNamespace: "test-namespace",
			expectedUsers:     []string{"admin"},
		},
		{
			name: "apply manifest fails",
			config: &ConfigurationSpecs{
				Configuration: Configuration{
					KubeSliceConfiguration: KubeSliceConfiguration{
						ProjectName: "fail-project",
					},
					ClusterConfiguration: ClusterConfiguration{
						ControllerCluster: Cluster{Name: "controller"},
					},
				},
			},
			cliOptions:      nil,
			applyShouldFail: true,
			mockApplyManifest: func(fileName, namespace string, cluster *Cluster) {
				util.Fatalf("Process failed %v", errors.New("apply failed"))
			},
			expectedUsers: []string{"admin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeFS := util.FileSystem.(*util.FakeFileSystem)
			fakeOutput := util.Output.(*util.FakeOutput)
			fakeClock := util.SystemClock.(*util.FakeClock)

			applyCalled := false
			originalApply := applyKubectlManifestFuncProject
			applyKubectlManifestFuncProject = func(fileName, namespace string, cluster *Cluster) {
				applyCalled = true
				if tt.mockApplyManifest != nil {
					tt.mockApplyManifest(fileName, namespace, cluster)
				}
			}
			defer func() { applyKubectlManifestFuncProject = originalApply }()

			CreateKubeSliceProject(tt.config, tt.cliOptions)

			// Verify manifest was generated
			manifestPath := kubesliceDirectory + "/" + projectFileName
			content, exists := fakeFS.WrittenFiles[manifestPath]
			require.True(t, exists, "Project manifest should be generated")

			contentStr := string(content)
			assert.Contains(t, contentStr, "kind: Project")
			assert.Contains(t, contentStr, "apiVersion: controller.kubeslice.io/v1alpha1")
			assert.Contains(t, contentStr, "name: "+tt.config.Configuration.KubeSliceConfiguration.ProjectName)
			for _, user := range tt.expectedUsers {
				assert.Contains(t, contentStr, user)
			}

			assert.True(t, applyCalled, "Apply manifest should be called")

			if tt.applyShouldFail {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), "apply failed")
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				require.Len(t, fakeClock.SleepCalls, 2)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
				assert.Equal(t, 3*time.Second, fakeClock.SleepCalls[1])
			}
		})
	}
}

func TestGenerateKubeSliceProjectManifest(t *testing.T) {
	tests := []struct {
		name          string
		projectName   string
		users         []string
		expectedUsers []string
	}{
		{
			name:          "generate with multiple users",
			projectName:   "test-project",
			users:         []string{"user1", "user2", "user3"},
			expectedUsers: []string{"user1", "user2", "user3"},
		},
		{
			name:          "generate with single user",
			projectName:   "single-user-project",
			users:         []string{"admin"},
			expectedUsers: []string{"admin"},
		},
		{
			name:          "generate with no users (default to admin)",
			projectName:   "default-project",
			users:         []string{},
			expectedUsers: []string{"admin"},
		},
		{
			name:          "generate with nil users (default to admin)",
			projectName:   "nil-users-project",
			users:         nil,
			expectedUsers: []string{"admin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeFS := util.FileSystem.(*util.FakeFileSystem)

			generateKubeSliceProjectManifest(tt.projectName, tt.users)

			manifestPath := kubesliceDirectory + "/" + projectFileName
			content, exists := fakeFS.WrittenFiles[manifestPath]
			require.True(t, exists, "Manifest file should be created")

			contentStr := string(content)
			assert.Contains(t, contentStr, "kind: Project")
			assert.Contains(t, contentStr, "apiVersion: controller.kubeslice.io/v1alpha1")
			assert.Contains(t, contentStr, "name: "+tt.projectName)

			for _, user := range tt.expectedUsers {
				assert.Contains(t, contentStr, user, "Content should contain user: %s", user)
			}
		})
	}
}

func TestGetKubeSliceProject(t *testing.T) {
	cluster := &Cluster{Name: "controller", ContextName: "controller-ctx"}

	tests := []struct {
		name          string
		projectName   string
		namespace     string
		mockGet       func(string, string, string, *Cluster, string)
		expectFatal   bool
		fatalContains string
	}{
		{
			name:        "successful get",
			projectName: "my-project",
			namespace:   "test-namespace",
			mockGet: func(resourceType, resourceName, namespace string, c *Cluster, outputFormat string) {
				assert.Equal(t, ProjectObject, resourceType)
				assert.Equal(t, "my-project", resourceName)
				assert.Equal(t, "test-namespace", namespace)
				assert.Equal(t, cluster, c)
				assert.Equal(t, "", outputFormat)
			},
		},
		{
			name:        "get fails",
			projectName: "my-project",
			namespace:   "test-namespace",
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

			originalGet := getKubectlResourcesFuncProject
			getKubectlResourcesFuncProject = tt.mockGet
			defer func() { getKubectlResourcesFuncProject = originalGet }()

			GetKubeSliceProject(tt.projectName, tt.namespace, cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			}
		})
	}
}

func TestDeleteKubeSliceProject(t *testing.T) {
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
				assert.Equal(t, ProjectObject, resourceType)
				assert.Equal(t, "my-project", resourceName)
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

			originalDelete := deleteKubectlResourcesFuncProject
			deleteKubectlResourcesFuncProject = tt.mockDelete
			defer func() { deleteKubectlResourcesFuncProject = originalDelete }()

			DeleteKubeSliceProject("my-project", "test-namespace", cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			}
		})
	}
}

func TestEditKubeSliceProject(t *testing.T) {
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
				assert.Equal(t, ProjectObject, resourceType)
				assert.Equal(t, "my-project", resourceName)
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

			originalEdit := editKubectlResourcesFuncProject
			editKubectlResourcesFuncProject = tt.mockEdit
			defer func() { editKubectlResourcesFuncProject = originalEdit }()

			EditKubeSliceProject("my-project", "test-namespace", cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			}
		})
	}
}

func TestDescribeKubeSliceProject(t *testing.T) {
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
				assert.Equal(t, ProjectObject, resourceType)
				assert.Equal(t, "my-project", resourceName)
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

			originalDescribe := describeKubectlResourcesFuncProject
			describeKubectlResourcesFuncProject = tt.mockDescribe
			defer func() { describeKubectlResourcesFuncProject = originalDescribe }()

			DescribeKubeSliceProject("my-project", "test-namespace", cluster)

			if tt.expectFatal {
				require.NotEmpty(t, fakeOutput.FatalCalls)
				assert.Contains(t, fmt.Sprint(fakeOutput.FatalCalls[0]), tt.fatalContains)
			} else {
				assert.Empty(t, fakeOutput.FatalCalls)
				assert.Len(t, fakeClock.SleepCalls, 1)
				assert.Equal(t, 200*time.Millisecond, fakeClock.SleepCalls[0])
			}
		})
	}
}
