package cmd_test

import (
	"testing"

	"github.com/jasonjacinth/infra-cli/cmd"
)

// TestDeployRequiresAppFlag confirms that the deploy command has --app registered.
func TestDeployRequiresAppFlag(t *testing.T) {
	root := cmd.RootCmd()

	// Find the deploy subcommand in the registered command tree.
	var deployFound bool
	for _, sub := range root.Commands() {
		if sub.Name() != "deploy" {
			continue
		}
		deployFound = true

		appFlag := sub.Flags().Lookup("app")
		if appFlag == nil {
			t.Fatal("expected --app flag to be registered on the deploy command, but it was not found")
		}
	}

	if !deployFound {
		t.Fatal("deploy subcommand not found on root command")
	}
}

// TestDeployHelpContainsExamples verifies that the deploy command's help text
// includes usage examples, which are important for developer-facing CLIs.
func TestDeployHelpContainsExamples(t *testing.T) {
	root := cmd.RootCmd()

	for _, sub := range root.Commands() {
		if sub.Name() != "deploy" {
			continue
		}
		if sub.Long == "" {
			t.Error("expected deploy command to have a Long description with examples")
		}
	}
}
