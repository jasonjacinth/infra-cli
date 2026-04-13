package shell

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// IsInstalled checks if a binary exists in the system PATH.
func IsInstalled(binary string) bool {
	_, err := exec.LookPath(binary)
	return err == nil
}

// Run executes a command and returns its combined stdout/stderr output.
// If the command fails, a user-friendly error is returned.
func Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Build a helpful error message.
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return "", fmt.Errorf("command '%s %s' failed:\n  %s", name, strings.Join(args, " "), stderrStr)
		}
		return "", fmt.Errorf("command '%s %s' failed: %w", name, strings.Join(args, " "), err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// RunLive executes a command with stdout and stderr connected directly
// to the terminal. Use this for streaming/interactive output (e.g. logs -f).
func RunLive(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command '%s %s' failed: %w", name, strings.Join(args, " "), err)
	}
	return nil
}
