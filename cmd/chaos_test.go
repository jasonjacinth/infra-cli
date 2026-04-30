package cmd_test

import (
	"testing"

	"github.com/jasonjacinth/infra-cli/cmd"
	"github.com/spf13/cobra"
)

// TestChaosSubcommandRegistered verifies that "chaos" has the "pod-kill" subcommand.
func TestChaosSubcommandRegistered(t *testing.T) {
	root := cmd.RootCmd()

	var chaosCmd *cobra.Command
	for _, sub := range root.Commands() {
		if sub.Name() == "chaos" {
			chaosCmd = sub
			break
		}
	}

	if chaosCmd == nil {
		t.Fatal("chaos subcommand not found on root command")
	}

	var podKillFound bool
	for _, sub := range chaosCmd.Commands() {
		if sub.Name() == "pod-kill" {
			podKillFound = true

			appFlag := sub.Flags().Lookup("app")
			if appFlag == nil {
				t.Error("expected --app flag to be registered on 'chaos pod-kill', but it was not found")
			}
		}
	}

	if !podKillFound {
		t.Error("expected 'pod-kill' subcommand to be registered under 'chaos', but it was not")
	}
}
