package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kubeslice/kubeslice-cli/util"
)

// TestVerificationResult verifies that verificationResult correctly identifies
// each numeric status code and maps it to the expected outcome. Status 0 must
// not call Fatalf, status 1 and 2 trigger a fatal exit so we only test status 0
// and the branch paths by examining the formatted output message.
func TestVerificationResult_SuccessStatus(t *testing.T) {
	t.Parallel()

	// Redirect stdout so we can inspect the printed output without side effects.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	verificationResult(0, "kubectl")

	w.Close()
	os.Stdout = origStdout

	buf := make([]byte, 256)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if output == "" {
		t.Error("verificationResult(0, ...) produced no output, expected a success message")
	}
}

// TestVerifyBinary_NotFoundReturnsOne checks that _verifyBinary returns 1
// when the executable does not exist on the system path and the override
// environment variable is not set.
func TestVerifyBinary_NotFoundReturnsOne(t *testing.T) {
	t.Parallel()

	// Use a binary name that is guaranteed not to exist.
	got := _verifyBinary("kubeslice-nonexistent-binary-xyz", "KUBESLICE_NONEXISTENT_PATH", []string{"version"})
	if got != 1 {
		t.Errorf("_verifyBinary(): want 1 (not found), got %d", got)
	}
}

// TestVerifyBinary_EnvOverrideNotFound checks that when the environment
// variable override points to a non-existent path, the function still returns 1.
func TestVerifyBinary_EnvOverrideNotFound(t *testing.T) {
	t.Parallel()

	const envVar = "KUBESLICE_TEST_OVERRIDE_PATH_XYZ"
	t.Setenv(envVar, "/nonexistent/path/to/binary")

	got := _verifyBinary("kubectl", envVar, []string{"version", "--client=true"})
	if got != 1 {
		t.Errorf("_verifyBinary() with bad env override: want 1 (not found), got %d", got)
	}
}

// TestVerifyBinary_ValidBinaryReturnsZero checks that _verifyBinary returns 0
// when pointing to a real executable that runs successfully. We use a small
// shell script (or the system 'true' binary) so this test stays hermetic.
func TestVerifyBinary_ValidBinaryReturnsZero(t *testing.T) {
	t.Parallel()

	// Create a temporary script that always exits 0.
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "fake-binary")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Skipf("skipping: cannot create temp script: %v", err)
	}

	// Pre-populate ExecutablePaths so RunCommandCustomIO doesn't panic on lookup.
	util.ExecutablePaths = map[string]string{
		"fake-binary": scriptPath,
	}

	const envVar = "KUBESLICE_FAKE_BINARY_PATH"
	t.Setenv(envVar, scriptPath)

	got := _verifyBinary("fake-binary", envVar, []string{})
	if got != 0 {
		t.Errorf("_verifyBinary() with valid binary: want 0 (success), got %d", got)
	}
}

// TestVerifyBinary_FoundButNotExecutableReturnsTwo checks that _verifyBinary
// returns 2 when the binary is found but its verify command fails (e.g., the
// command exits non-zero). We point to /bin/false (always exits 1) as the
// executable under test.
func TestVerifyBinary_FoundButCommandFailsReturnsTwo(t *testing.T) {
	t.Parallel()

	// Locate 'false' binary (exits 1 by design) — available on Linux/macOS CI.
	falsePath, err := os.Stat("/bin/false")
	if err != nil || falsePath.IsDir() {
		t.Skip("skipping: /bin/false not available on this platform")
	}

	const envVar = "KUBESLICE_FALSE_BINARY_PATH"
	t.Setenv(envVar, "/bin/false")

	// Any argument list will cause /bin/false to exit non-zero.
	got := _verifyBinary("false-binary", envVar, []string{"--some-arg"})
	if got != 2 {
		t.Errorf("_verifyBinary() with failing command: want 2 (not executable), got %d", got)
	}
}
