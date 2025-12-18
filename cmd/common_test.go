package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapFromSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string]string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: map[string]string{},
		},
		{
			name:     "single element",
			input:    []string{"test"},
			expected: map[string]string{"test": ""},
		},
		{
			name:     "multiple elements",
			input:    []string{"step1", "step2", "step3"},
			expected: map[string]string{"step1": "", "step2": "", "step3": ""},
		},
		{
			name:     "duplicate elements",
			input:    []string{"step1", "step1", "step2"},
			expected: map[string]string{"step1": "", "step2": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapFromSlice(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGlobalVariables(t *testing.T) {
	// Test that global variables are properly initialized
	assert.Equal(t, "", profile)
	assert.Equal(t, "", outputFormat)
	assert.Equal(t, "", Config)
	assert.Equal(t, []string{}, skipSteps)
}

func TestGlobalVariableTypes(t *testing.T) {
	// Test that global variables have correct types
	assert.IsType(t, "", profile)
	assert.IsType(t, []string{}, skipSteps)
	assert.IsType(t, "", outputFormat)
	assert.IsType(t, "", Config)
}

func TestMapFromSliceNil(t *testing.T) {
	// Test with nil input
	result := mapFromSlice(nil)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

func TestMapFromSliceWithSpecialCharacters(t *testing.T) {
	input := []string{"step-1", "step_2", "step.3", "step@4"}
	expected := map[string]string{
		"step-1": "",
		"step_2": "",
		"step.3": "",
		"step@4": "",
	}
	
	result := mapFromSlice(input)
	assert.Equal(t, expected, result)
}

func TestMapFromSliceWithEmptyStrings(t *testing.T) {
	input := []string{"", "step1", "", "step2"}
	expected := map[string]string{
		"":      "",
		"step1": "",
		"step2": "",
	}
	
	result := mapFromSlice(input)
	assert.Equal(t, expected, result)
}

func TestMapFromSliceReturnType(t *testing.T) {
	result := mapFromSlice([]string{"test"})
	assert.IsType(t, map[string]string{}, result)
}

func TestMapFromSliceLargeInput(t *testing.T) {
	// Test with a large number of elements
	input := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = "step" + string(rune(i))
	}
	
	result := mapFromSlice(input)
	assert.Equal(t, 1000, len(result))
	
	// Check that all keys exist and have empty string values
	for _, step := range input {
		value, exists := result[step]
		assert.True(t, exists)
		assert.Equal(t, "", value)
	}
}
