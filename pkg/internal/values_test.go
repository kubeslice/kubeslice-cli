package internal

import (
	"testing"

	"github.com/kubeslice/kubeslice-cli/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestGenerateValuesFile(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		helmChart     *HelmChart
		defaults      string
		expectError   bool
		errorContains string
		validateYAML  func(*testing.T, []byte)
	}{
		{
			name:     "simple values generation",
			filePath: "test-values.yaml",
			helmChart: &HelmChart{
				Values: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			defaults: `
default1: defaultvalue1
`,
			validateYAML: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := yaml.Unmarshal(data, &result)
				require.NoError(t, err)
				assert.Equal(t, "value1", result["key1"])
				assert.Equal(t, "value2", result["key2"])
				assert.Equal(t, "defaultvalue1", result["default1"])
			},
		},
		{
			name:     "nested values with dot notation",
			filePath: "test-values.yaml",
			helmChart: &HelmChart{
				Values: map[string]interface{}{
					"parent.child1": "value1",
					"parent.child2": "value2",
				},
			},
			defaults: ``,
			validateYAML: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := yaml.Unmarshal(data, &result)
				require.NoError(t, err)
				parent := result["parent"].(map[interface{}]interface{})
				assert.Equal(t, "value1", parent["child1"])
				assert.Equal(t, "value2", parent["child2"])
			},
		},
		{
			name:     "deep nested values",
			filePath: "test-values.yaml",
			helmChart: &HelmChart{
				Values: map[string]interface{}{
					"level1.level2.level3": "deepvalue",
				},
			},
			defaults: ``,
			validateYAML: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := yaml.Unmarshal(data, &result)
				require.NoError(t, err)
				level1 := result["level1"].(map[interface{}]interface{})
				level2 := level1["level2"].(map[interface{}]interface{})
				assert.Equal(t, "deepvalue", level2["level3"])
			},
		},
		{
			name:     "values override defaults",
			filePath: "test-values.yaml",
			helmChart: &HelmChart{
				Values: map[string]interface{}{
					"key1": "customvalue",
				},
			},
			defaults: `
key1: defaultvalue
key2: defaultvalue2
`,
			validateYAML: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := yaml.Unmarshal(data, &result)
				require.NoError(t, err)
				assert.Equal(t, "customvalue", result["key1"], "Custom value should override default")
				assert.Equal(t, "defaultvalue2", result["key2"], "Non-overridden default should remain")
			},
		},
		{
			name:     "empty values with defaults",
			filePath: "test-values.yaml",
			helmChart: &HelmChart{
				Values: map[string]interface{}{},
			},
			defaults: `
key1: value1
key2: value2
`,
			validateYAML: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := yaml.Unmarshal(data, &result)
				require.NoError(t, err)
				assert.Equal(t, "value1", result["key1"])
				assert.Equal(t, "value2", result["key2"])
			},
		},
		{
			name:     "complex nested merge",
			filePath: "test-values.yaml",
			helmChart: &HelmChart{
				Values: map[string]interface{}{
					"app.database.host": "custom-host",
					"app.cache.enabled": true,
				},
			},
			defaults: `
app:
  database:
    host: default-host
    port: 5432
  cache:
    enabled: false
    ttl: 3600
`,
			validateYAML: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := yaml.Unmarshal(data, &result)
				require.NoError(t, err)
				app := result["app"].(map[interface{}]interface{})
				database := app["database"].(map[interface{}]interface{})
				cache := app["cache"].(map[interface{}]interface{})
				assert.Equal(t, "custom-host", database["host"])
				assert.Equal(t, 5432, database["port"])
				assert.Equal(t, true, cache["enabled"])
				assert.Equal(t, 3600, cache["ttl"])
			},
		},
		{
			name:     "invalid YAML in defaults",
			filePath: "test-values.yaml",
			helmChart: &HelmChart{
				Values: map[string]interface{}{
					"key1": "value1",
				},
			},
			defaults: `
invalid yaml:
  this is: [not: valid
`,
			expectError:   true,
			errorContains: "error parsing defaults",
		},
		{
			name:     "nil values map",
			filePath: "test-values.yaml",
			helmChart: &HelmChart{
				Values: nil,
			},
			defaults: "key: value",
			validateYAML: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := yaml.Unmarshal(data, &result)
				require.NoError(t, err)
				assert.Equal(t, "value", result["key"])
			},
		},
		{
			name:     "empty defaults",
			filePath: "test-values.yaml",
			helmChart: &HelmChart{
				Values: map[string]interface{}{
					"key": "value",
				},
			},
			defaults: "",
			validateYAML: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := yaml.Unmarshal(data, &result)
				require.NoError(t, err)
				assert.Equal(t, "value", result["key"])
			},
		},
		{
			name:     "special characters in keys",
			filePath: "test-values.yaml",
			helmChart: &HelmChart{
				Values: map[string]interface{}{
					"key-with-dash":       "value1",
					"key_with_underscore": "value2",
				},
			},
			defaults: "",
			validateYAML: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := yaml.Unmarshal(data, &result)
				require.NoError(t, err)
				assert.Equal(t, "value1", result["key-with-dash"])
				assert.Equal(t, "value2", result["key_with_underscore"])
			},
		},
		{
			name:     "numeric and boolean values",
			filePath: "test-values.yaml",
			helmChart: &HelmChart{
				Values: map[string]interface{}{
					"port":    8080,
					"timeout": 30.5,
					"enabled": true,
				},
			},
			defaults: "",
			validateYAML: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := yaml.Unmarshal(data, &result)
				require.NoError(t, err)
				assert.Equal(t, 8080, result["port"])
				assert.Equal(t, 30.5, result["timeout"])
				assert.Equal(t, true, result["enabled"])
			},
		},
		{
			name:     "array values",
			filePath: "test-values.yaml",
			helmChart: &HelmChart{
				Values: map[string]interface{}{
					"items": []string{"item1", "item2", "item3"},
				},
			},
			defaults: "",
			validateYAML: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := yaml.Unmarshal(data, &result)
				require.NoError(t, err)
				items := result["items"].([]interface{})
				assert.Len(t, items, 3)
				assert.Equal(t, "item1", items[0])
				assert.Equal(t, "item2", items[1])
				assert.Equal(t, "item3", items[2])
			},
		},
		{
			name:     "nested map merging - custom and default values",
			filePath: "test-values.yaml",
			helmChart: &HelmChart{
				Values: map[string]interface{}{
					"parent.child1": "custom1",
				},
			},
			defaults: `
parent:
  child1: default1
  child2: default2
`,
			validateYAML: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				err := yaml.Unmarshal(data, &result)
				require.NoError(t, err)
				parent := result["parent"].(map[interface{}]interface{})
				assert.Equal(t, "custom1", parent["child1"], "Custom value should override default")
				assert.Equal(t, "default2", parent["child2"], "Default should be preserved")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := util.NewTestEnvironment()
			defer cleanup()

			fakeFS := util.FileSystem.(*util.FakeFileSystem)

			err := generateValuesFile(tt.filePath, tt.helmChart, tt.defaults)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)

				data, exists := fakeFS.WrittenFiles[tt.filePath]
				require.True(t, exists, "Expected file to be written")

				if tt.validateYAML != nil {
					tt.validateYAML(t, data)
				}
			}
		})
	}
}
