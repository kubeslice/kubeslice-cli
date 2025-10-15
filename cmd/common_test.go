package cmd

import (
	"reflect"
	"testing"
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
			input:    []string{"step1"},
			expected: map[string]string{"step1": ""},
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
		{
			name:     "case sensitivity",
			input:    []string{"Step1", "step1", "STEP1"},
			expected: map[string]string{"Step1": "", "step1": "", "STEP1": ""},
		},
		{
			name:     "special characters",
			input:    []string{"step-1", "step_2", "step@3"},
			expected: map[string]string{"step-1": "", "step_2": "", "step@3": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapFromSlice(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("mapFromSlice(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGlobalVariables(t *testing.T) {
	// Test global variables are accessible and modifiable
	profile = "test-profile"
	if profile != "test-profile" {
		t.Errorf("profile = %s, want test-profile", profile)
	}

	skipSteps = []string{"step1", "step2"}
	if len(skipSteps) != 2 || skipSteps[0] != "step1" || skipSteps[1] != "step2" {
		t.Errorf("skipSteps = %v, want [step1 step2]", skipSteps)
	}

	outputFormat = "json"
	if outputFormat != "json" {
		t.Errorf("outputFormat = %s, want json", outputFormat)
	}

	Config = "config.yaml"
	if Config != "config.yaml" {
		t.Errorf("Config = %s, want config.yaml", Config)
	}
}

func TestMapFromSlice_NilInput(t *testing.T) {
	result := mapFromSlice(nil)
	if result == nil {
		t.Error("mapFromSlice(nil) should not return nil map")
	}
	if len(result) != 0 {
		t.Errorf("mapFromSlice(nil) = %v, want empty map", result)
	}
}
