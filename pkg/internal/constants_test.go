package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	// Test that all constants are properly defined and have expected values
	assert.Equal(t, "kubeslice-controller", KUBESLICE_CONTROLLER_NAMESPACE)
	assert.Equal(t, "projects.controller.kubeslice.io", ProjectObject)
	assert.Equal(t, "clusters.controller.kubeslice.io", ClusterObject)
	assert.Equal(t, "sliceconfigs.controller.kubeslice.io", SliceConfigObject)
	assert.Equal(t, "serviceexportconfigs.controller.kubeslice.io", ServiceExportConfigObject)
	assert.Equal(t, "kubeslice-license-file", LicenseFileName)
}

func TestComponentConstants(t *testing.T) {
	// Test component-related constants
	assert.Equal(t, "kind", Kind_Component)
	assert.Equal(t, "calico", Calico_Component)
	assert.Equal(t, "controller", Controller_Component)
	assert.Equal(t, "worker-registration", Worker_registration_Component)
	assert.Equal(t, "ui", UI_install_Component)
	assert.Equal(t, "worker", Worker_Component)
	assert.Equal(t, "demo", Demo_Component)
	assert.Equal(t, "cert-manager", CertManager_Component)
	assert.Equal(t, "prometheus", Prometheus_Component)
}

func TestObjectConstants(t *testing.T) {
	// Test Kubernetes object-related constants
	assert.Equal(t, "secrets", SecretObject)
}

func TestOutputFormatConstants(t *testing.T) {
	// Test output format constants
	assert.Equal(t, "yaml", OutputFormatYaml)
	assert.Equal(t, "json", OutputFormatJson)
}

func TestConstantTypes(t *testing.T) {
	// Test that all constants are strings
	assert.IsType(t, "", KUBESLICE_CONTROLLER_NAMESPACE)
	assert.IsType(t, "", ProjectObject)
	assert.IsType(t, "", ClusterObject)
	assert.IsType(t, "", SliceConfigObject)
	assert.IsType(t, "", ServiceExportConfigObject)
	assert.IsType(t, "", LicenseFileName)
	assert.IsType(t, "", Kind_Component)
	assert.IsType(t, "", Calico_Component)
	assert.IsType(t, "", Controller_Component)
	assert.IsType(t, "", Worker_registration_Component)
	assert.IsType(t, "", UI_install_Component)
	assert.IsType(t, "", Worker_Component)
	assert.IsType(t, "", Demo_Component)
	assert.IsType(t, "", CertManager_Component)
	assert.IsType(t, "", Prometheus_Component)
	assert.IsType(t, "", SecretObject)
	assert.IsType(t, "", OutputFormatYaml)
	assert.IsType(t, "", OutputFormatJson)
}

func TestNamespaceConstant(t *testing.T) {
	// Test namespace constant specifically
	namespace := KUBESLICE_CONTROLLER_NAMESPACE
	assert.NotEmpty(t, namespace)
	assert.Contains(t, namespace, "kubeslice")
	assert.Contains(t, namespace, "controller")
}

func TestObjectConstantFormat(t *testing.T) {
	// Test that object constants follow the expected format
	objectConstants := []string{
		ProjectObject,
		ClusterObject,
		SliceConfigObject,
		ServiceExportConfigObject,
	}

	for _, obj := range objectConstants {
		assert.Contains(t, obj, ".controller.kubeslice.io")
		assert.NotEmpty(t, obj)
	}
}

func TestComponentConstantUniqueness(t *testing.T) {
	// Test that all component constants are unique
	components := []string{
		Kind_Component,
		Calico_Component,
		Controller_Component,
		Worker_registration_Component,
		UI_install_Component,
		Worker_Component,
		Demo_Component,
		CertManager_Component,
		Prometheus_Component,
	}

	uniqueComponents := make(map[string]bool)
	for _, component := range components {
		assert.False(t, uniqueComponents[component], "Component %s is not unique", component)
		uniqueComponents[component] = true
	}

	assert.Equal(t, 9, len(uniqueComponents))
}

func TestOutputFormatValues(t *testing.T) {
	// Test that output format constants have valid values
	assert.Equal(t, "yaml", OutputFormatYaml)
	assert.Equal(t, "json", OutputFormatJson)
	
	// Test that they are different
	assert.NotEqual(t, OutputFormatYaml, OutputFormatJson)
}

func TestConstantNamingConvention(t *testing.T) {
	// Test that constants follow expected naming conventions
	
	// Component constants should end with "_Component"
	componentConstants := map[string]string{
		"Kind_Component":                Kind_Component,
		"Calico_Component":              Calico_Component,
		"Controller_Component":          Controller_Component,
		"Worker_registration_Component": Worker_registration_Component,
		"UI_install_Component":          UI_install_Component,
		"Worker_Component":              Worker_Component,
		"Demo_Component":                Demo_Component,
		"CertManager_Component":         CertManager_Component,
		"Prometheus_Component":          Prometheus_Component,
	}

	for name, value := range componentConstants {
		assert.Contains(t, name, "_Component")
		assert.NotEmpty(t, value)
	}

	// Object constants should end with "Object"
	objectConstants := map[string]string{
		"ProjectObject":             ProjectObject,
		"ClusterObject":             ClusterObject,
		"SliceConfigObject":         SliceConfigObject,
		"ServiceExportConfigObject": ServiceExportConfigObject,
		"SecretObject":              SecretObject,
	}

	for name, value := range objectConstants {
		assert.Contains(t, name, "Object")
		assert.NotEmpty(t, value)
	}
}

func TestLicenseFileName(t *testing.T) {
	// Test license file name constant
	assert.Equal(t, "kubeslice-license-file", LicenseFileName)
	assert.Contains(t, LicenseFileName, "kubeslice")
	assert.Contains(t, LicenseFileName, "license")
}

func TestConstantsNotEmpty(t *testing.T) {
	// Ensure all constants are not empty
	constants := []string{
		KUBESLICE_CONTROLLER_NAMESPACE,
		ProjectObject,
		ClusterObject,
		SliceConfigObject,
		ServiceExportConfigObject,
		LicenseFileName,
		Kind_Component,
		Calico_Component,
		Controller_Component,
		Worker_registration_Component,
		UI_install_Component,
		Worker_Component,
		Demo_Component,
		CertManager_Component,
		Prometheus_Component,
		SecretObject,
		OutputFormatYaml,
		OutputFormatJson,
	}

	for _, constant := range constants {
		assert.NotEmpty(t, constant, "Constant should not be empty")
	}
}

func TestKubernetesResourceConstants(t *testing.T) {
	// Test that Kubernetes resource constants follow API group format
	kubernetesResources := []string{
		ProjectObject,
		ClusterObject,
		SliceConfigObject,
		ServiceExportConfigObject,
	}

	for _, resource := range kubernetesResources {
		// Should contain the API group
		assert.Contains(t, resource, "controller.kubeslice.io")
		
		// Should have a resource name before the API group
		parts := splitString(resource, ".")
		assert.GreaterOrEqual(t, len(parts), 3) // resource.controller.kubeslice.io
	}
}

// Helper function to split string (simple implementation)
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	
	var result []string
	start := 0
	
	for i := 0; i <= len(s)-len(sep); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			if start <= i {
				result = append(result, s[start:i])
			}
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	
	if start < len(s) {
		result = append(result, s[start:])
	}
	
	return result
}
