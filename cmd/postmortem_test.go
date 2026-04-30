package cmd_test

import (
	"testing"

	"github.com/jasonjacinth/infra-cli/cmd"
	"github.com/spf13/cobra"
)

// TestPostmortemSubcommandRegistered verifies that "postmortem" has "create" with
// the required --title and --severity flags.
func TestPostmortemSubcommandRegistered(t *testing.T) {
	root := cmd.RootCmd()

	var postmortemCmd *cobra.Command
	for _, sub := range root.Commands() {
		if sub.Name() == "postmortem" {
			postmortemCmd = sub
			break
		}
	}

	if postmortemCmd == nil {
		t.Fatal("postmortem subcommand not found on root command")
	}

	var createFound bool
	for _, sub := range postmortemCmd.Commands() {
		if sub.Name() == "create" {
			createFound = true

			titleFlag := sub.Flags().Lookup("title")
			if titleFlag == nil {
				t.Error("expected --title flag on 'postmortem create'")
			}

			severityFlag := sub.Flags().Lookup("severity")
			if severityFlag == nil {
				t.Error("expected --severity flag on 'postmortem create'")
			}

			outputFlag := sub.Flags().Lookup("output")
			if outputFlag == nil {
				t.Error("expected --output flag on 'postmortem create'")
			}
		}
	}

	if !createFound {
		t.Error("expected 'create' subcommand to be registered under 'postmortem', but it was not")
	}
}
