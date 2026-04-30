package cmd_test

import (
	"testing"

	"github.com/jasonjacinth/infra-cli/cmd"
	"github.com/spf13/cobra"
)

// TestSLOSubcommandRegistered verifies that "slo" has "validate" with --app.
func TestSLOSubcommandRegistered(t *testing.T) {
	root := cmd.RootCmd()

	var sloCmd *cobra.Command
	for _, sub := range root.Commands() {
		if sub.Name() == "slo" {
			sloCmd = sub
			break
		}
	}

	if sloCmd == nil {
		t.Fatal("slo subcommand not found on root command")
	}

	var validateFound bool
	for _, sub := range sloCmd.Commands() {
		if sub.Name() == "validate" {
			validateFound = true

			appFlag := sub.Flags().Lookup("app")
			if appFlag == nil {
				t.Error("expected --app flag on 'slo validate'")
			}
		}
	}

	if !validateFound {
		t.Error("expected 'validate' subcommand to be registered under 'slo', but it was not")
	}
}
