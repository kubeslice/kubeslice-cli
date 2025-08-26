package internal

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name     string
		dest     map[interface{}]interface{}
		src      map[interface{}]interface{}
		expected map[interface{}]interface{}
	}{
		{
			name: "simple merge - source overwrites destination",
			dest: map[interface{}]interface{}{
				"a": 1,
				"b": 2,
			},
			src: map[interface{}]interface{}{
				"b": 3,
				"c": 4,
			},
			expected: map[interface{}]interface{}{
				"a": 1,
				"b": 3,
				"c": 4,
			},
		},
		{
			name: "nested map merge",
			dest: map[interface{}]interface{}{
				"config": map[interface{}]interface{}{
					"port": 8080,
					"host": "localhost",
				},
			},
			src: map[interface{}]interface{}{
				"config": map[interface{}]interface{}{
					"port":    9090,
					"timeout": 30,
				},
			},
			expected: map[interface{}]interface{}{
				"config": map[interface{}]interface{}{
					"port":    9090,
					"host":    "localhost",
					"timeout": 30,
				},
			},
		},
		{
			name: "deep nested merge",
			dest: map[interface{}]interface{}{
				"kubeslice": map[interface{}]interface{}{
					"controller": map[interface{}]interface{}{
						"loglevel": "info",
						"endpoint": "localhost:8080",
					},
				},
			},
			src: map[interface{}]interface{}{
				"kubeslice": map[interface{}]interface{}{
					"controller": map[interface{}]interface{}{
						"loglevel":           "debug",
						"rbacResourcePrefix": "kubeslice-rbac",
					},
				},
			},
			expected: map[interface{}]interface{}{
				"kubeslice": map[interface{}]interface{}{
					"controller": map[interface{}]interface{}{
						"loglevel":           "debug",
						"endpoint":           "localhost:8080",
						"rbacResourcePrefix": "kubeslice-rbac",
					},
				},
			},
		},
		{
			name: "empty source map",
			dest: map[interface{}]interface{}{
				"a": 1,
				"b": 2,
			},
			src: map[interface{}]interface{}{},
			expected: map[interface{}]interface{}{
				"a": 1,
				"b": 2,
			},
		},
		{
			name: "empty destination map",
			dest: map[interface{}]interface{}{},
			src: map[interface{}]interface{}{
				"a": 1,
				"b": 2,
			},
			expected: map[interface{}]interface{}{
				"a": 1,
				"b": 2,
			},
		},
		{
			name: "nil source map",
			dest: map[interface{}]interface{}{
				"a": 1,
			},
			src: nil,
			expected: map[interface{}]interface{}{
				"a": 1,
			},
		},
		{
			name: "nil destination map",
			dest: nil,
			src: map[interface{}]interface{}{
				"a": 1,
			},
			expected: map[interface{}]interface{}{
				"a": 1,
			},
		},
		{
			name: "mixed value types",
			dest: map[interface{}]interface{}{
				"string": "hello",
				"int":    42,
				"bool":   true,
				"slice":  []interface{}{1, 2, 3},
			},
			src: map[interface{}]interface{}{
				"string": "world",
				"float":  3.14,
				"map": map[interface{}]interface{}{
					"key": "value",
				},
			},
			expected: map[interface{}]interface{}{
				"string": "world",
				"int":    42,
				"bool":   true,
				"slice":  []interface{}{1, 2, 3},
				"float":  3.14,
				"map": map[interface{}]interface{}{
					"key": "value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of dest to avoid modifying the original
			var destCopy map[interface{}]interface{}
			if tt.dest != nil {
				destCopy = make(map[interface{}]interface{})
				for k, v := range tt.dest {
					destCopy[k] = v
				}
			}

			result := MergeMaps(destCopy, tt.src)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("mergeMaps() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateValuesFile(t *testing.T) {
	tests := []struct {
		name         string
		hc           *HelmChart
		defaults     string
		expectedYAML string
		expectError  bool
	}{
		{
			name: "basic values generation",
			hc: &HelmChart{
				ChartName: "test-chart",
				Values: map[string]interface{}{
					"replicaCount":     3,
					"image.repository": "nginx",
					"image.tag":        "latest",
				},
			},
			defaults: `
replicaCount: 1
image:
  tag: v1.0.0
  pullPolicy: IfNotPresent
`,
			expectedYAML: `
replicaCount: 3
image:
  repository: nginx
  tag: latest
  pullPolicy: IfNotPresent
`,
			expectError: false,
		},
		{
			name: "complex nested values",
			hc: &HelmChart{
				ChartName: "kubeslice-controller",
				Values: map[string]interface{}{
					"kubeslice.controller.loglevel":           "debug",
					"kubeslice.controller.endpoint":           "localhost:8080",
					"kubeslice.controller.rbacResourcePrefix": "kubeslice-rbac",
					"kubeslice.controller.projectnsPrefix":    "kubeslice",
				},
			},
			defaults: `
kubeslice:
  controller:
    loglevel: info
    endpoint: ""
    rbacResourcePrefix: ""
    projectnsPrefix: ""
`,
			expectedYAML: `
kubeslice:
  controller:
    loglevel: debug
    endpoint: localhost:8080
    rbacResourcePrefix: kubeslice-rbac
    projectnsPrefix: kubeslice
`,
			expectError: false,
		},
		{
			name: "empty values map",
			hc: &HelmChart{
				ChartName: "empty-chart",
				Values:    map[string]interface{}{},
			},
			defaults: `
replicaCount: 1
image:
  repository: nginx
`,
			expectedYAML: `
replicaCount: 1
image:
  repository: nginx
`,
			expectError: false,
		},
		{
			name: "empty defaults",
			hc: &HelmChart{
				ChartName: "test-chart",
				Values: map[string]interface{}{
					"replicaCount":     3,
					"image.repository": "nginx",
				},
			},
			defaults: "",
			expectedYAML: `
replicaCount: 3
image:
  repository: nginx
`,
			expectError: false,
		},
		{
			name:        "nil helm chart",
			hc:          nil,
			defaults:    "replicaCount: 1",
			expectError: true,
		},
		{
			name: "invalid defaults YAML",
			hc: &HelmChart{
				ChartName: "test-chart",
				Values: map[string]interface{}{
					"replicaCount": 3,
				},
			},
			defaults: `
replicaCount: 1
  invalid: yaml: structure
`,
			expectError: true,
		},
		{
			name: "deep nested dot notation",
			hc: &HelmChart{
				ChartName: "complex-chart",
				Values: map[string]interface{}{
					"kubeslice.controller.logging.level":   "debug",
					"kubeslice.controller.logging.format":  "json",
					"kubeslice.controller.network.timeout": 30,
					"kubeslice.worker.config.endpoint":     "worker:8080",
				},
			},
			defaults: `
kubeslice:
  controller:
    logging:
      level: info
      format: text
    network:
      timeout: 60
`,
			expectedYAML: `
kubeslice:
  controller:
    logging:
      level: debug
      format: json
    network:
      timeout: 30
  worker:
    config:
      endpoint: worker:8080
`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "test-values-*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())

			err = GenerateValuesFile(tmpFile.Name(), tt.hc, tt.defaults)

			if tt.expectError {
				if err == nil {
					t.Errorf("generateValuesFile() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("generateValuesFile() failed: %v", err)
			}

			// Read the generated file
			content, err := os.ReadFile(tmpFile.Name())
			if err != nil {
				t.Fatal(err)
			}

			// Parse expected and actual YAML
			var result map[interface{}]interface{}
			var expected map[interface{}]interface{}

			err = yaml.Unmarshal(content, &result)
			if err != nil {
				t.Fatalf("Failed to parse generated YAML: %v", err)
			}

			err = yaml.Unmarshal([]byte(strings.TrimSpace(tt.expectedYAML)), &expected)
			if err != nil {
				t.Fatalf("Failed to parse expected YAML: %v", err)
			}

			// Compare the parsed structures
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("generateValuesFile() generated YAML does not match expected")
				t.Errorf("Generated: %v", result)
				t.Errorf("Expected: %v", expected)
			}
		})
	}
}

func TestGenerateValuesFileWithRealHelmChart(t *testing.T) {
	// Test with a realistic HelmChart configuration
	hc := &HelmChart{
		ChartName: "kubeslice-controller",
		Version:   "0.6.0",
		Values: map[string]interface{}{
			"kubeslice.controller.loglevel":           "debug",
			"kubeslice.controller.endpoint":           "controller.kubeslice.local:8080",
			"kubeslice.controller.rbacResourcePrefix": "kubeslice-rbac",
			"kubeslice.controller.projectnsPrefix":    "kubeslice",
		},
	}

	defaults := `
kubeslice:
  controller:
    loglevel: info
    endpoint: ""
    rbacResourcePrefix: ""
    projectnsPrefix: ""
`

	tmpFile, err := os.CreateTemp("", "real-values-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	err = GenerateValuesFile(tmpFile.Name(), hc, defaults)
	if err != nil {
		t.Fatalf("generateValuesFile() failed: %v", err)
	}

	// Verify the file was created and has content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if len(content) == 0 {
		t.Error("Generated values file is empty")
	}

	// Parse and verify the structure
	var result map[interface{}]interface{}
	err = yaml.Unmarshal(content, &result)
	if err != nil {
		t.Fatalf("Failed to parse generated YAML: %v", err)
	}

	// Check that the expected values are present
	kubeslice, ok := result["kubeslice"].(map[interface{}]interface{})
	if !ok {
		t.Fatal("Expected 'kubeslice' key in generated YAML")
	}

	controller, ok := kubeslice["controller"].(map[interface{}]interface{})
	if !ok {
		t.Fatal("Expected 'controller' key in kubeslice section")
	}

	expectedValues := map[string]interface{}{
		"loglevel":           "debug",
		"endpoint":           "controller.kubeslice.local:8080",
		"rbacResourcePrefix": "kubeslice-rbac",
		"projectnsPrefix":    "kubeslice",
	}

	for key, expectedValue := range expectedValues {
		if value, exists := controller[key]; !exists {
			t.Errorf("Expected key '%s' not found in controller section", key)
		} else if value != expectedValue {
			t.Errorf("Expected '%s' to be '%v', got '%v'", key, expectedValue, value)
		}
	}
}

func TestGenerateValuesFileErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		hc          *HelmChart
		defaults    string
		expectError bool
	}{
		{
			name:     "invalid file path",
			filePath: "/invalid/path/that/does/not/exist/values.yaml",
			hc: &HelmChart{
				ChartName: "test-chart",
				Values: map[string]interface{}{
					"replicaCount": 3,
				},
			},
			defaults:    "replicaCount: 1",
			expectError: true,
		},
		{
			name:        "nil helm chart",
			filePath:    "test.yaml",
			hc:          nil,
			defaults:    "replicaCount: 1",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GenerateValuesFile(tt.filePath, tt.hc, tt.defaults)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
