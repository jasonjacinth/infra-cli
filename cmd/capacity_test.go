package cmd_test

import (
	"testing"

	"github.com/jasonjacinth/infra-cli/cmd"
)

// TestCapacityCommandRegistered verifies that "capacity" has --app registered.
func TestCapacityCommandRegistered(t *testing.T) {
	root := cmd.RootCmd()

	var capacityFound bool
	for _, sub := range root.Commands() {
		if sub.Name() != "capacity" {
			continue
		}
		capacityFound = true

		appFlag := sub.Flags().Lookup("app")
		if appFlag == nil {
			t.Error("expected --app flag on 'capacity'")
		}
	}

	if !capacityFound {
		t.Fatal("capacity subcommand not found on root command")
	}
}
