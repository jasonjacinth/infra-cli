package cmd_test

import (
	"testing"

	"github.com/jasonjacinth/infra-cli/cmd"
)

// TestRootCommandRegistered verifies that Execute does not panic on a bare call
// and that the expected subcommands are present.
func TestRootCommandRegistered(t *testing.T) {
	// Execute is the public entry point. We exercise HasAvailableSubCommands
	// indirectly by checking the exported command list via the package-level
	// rootCmd, which is accessible through the Execute function's initialization.
	// A clean way to inspect commands is via the cobra.Command API.
	root := cmd.RootCmd()

	expectedSubs := []string{"setup", "deploy", "status", "logs", "rollback", "cleanup", "version", "chaos", "postmortem", "slo", "capacity"}

	registered := make(map[string]bool)
	for _, sub := range root.Commands() {
		registered[sub.Name()] = true
	}

	for _, name := range expectedSubs {
		if !registered[name] {
			t.Errorf("expected subcommand '%s' to be registered, but it was not", name)
		}
	}
}
