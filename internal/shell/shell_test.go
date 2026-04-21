package shell_test

import (
	"testing"

	"github.com/jasonjacinth/infra-cli/internal/shell"
)

// TestIsInstalled_KnownBinary checks that a universally available binary is detected.
func TestIsInstalled_KnownBinary(t *testing.T) {
	// "echo" is available on every Unix-like system.
	if !shell.IsInstalled("echo") {
		t.Error("expected 'echo' to be found in PATH, but it was not")
	}
}

// TestIsInstalled_MissingBinary checks that a non-existent binary returns false.
func TestIsInstalled_MissingBinary(t *testing.T) {
	if shell.IsInstalled("infra_cli_nonexistent_binary_xyz") {
		t.Error("expected non-existent binary to not be found, but IsInstalled returned true")
	}
}

// TestRun_Success verifies that a simple command runs and returns its output.
func TestRun_Success(t *testing.T) {
	out, err := shell.Run("echo", "hello")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out != "hello" {
		t.Errorf("expected output 'hello', got: %q", out)
	}
}

// TestRun_Failure verifies that a failing command returns a non-nil error.
func TestRun_Failure(t *testing.T) {
	// "ls" on a path that does not exist will exit non-zero.
	_, err := shell.Run("ls", "/this/path/does/not/exist/infra_cli_test")
	if err == nil {
		t.Error("expected an error for a failing command, but got nil")
	}
}

// TestRun_OutputTrimmed verifies that shell.Run trims leading/trailing whitespace.
func TestRun_OutputTrimmed(t *testing.T) {
	// printf without newline, then check that no extra whitespace is present.
	out, err := shell.Run("sh", "-c", "printf '  hello  '")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello" {
		t.Errorf("expected trimmed output 'hello', got: %q", out)
	}
}
