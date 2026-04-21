// Package guardrail enforces operational safety rules across all commands.
// It implements the "Platform as a Product" philosophy by acting as a policy
// layer between the user's intent and the underlying cluster operations.
package guardrail

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jasonjacinth/infra-cli/internal/style"
)

// restrictedNamespaces lists namespaces that are off-limits to application developers.
// These are reserved for cluster administrators only.
var restrictedNamespaces = map[string]bool{
	"kube-system":  true,
	"kube-public":  true,
	"kube-node-lease": true,
}

// CheckNamespace enforces namespace protection. It returns a fatal error if the
// provided namespace is in the restricted list. This prevents application developers
// from accidentally mutating cluster-critical workloads.
func CheckNamespace(namespace string) error {
	if restrictedNamespaces[namespace] {
		return fmt.Errorf(
			"[Guardrail Violation] Operations in the '%s' namespace are restricted to cluster administrators",
			namespace,
		)
	}
	return nil
}

// ConfirmProduction prompts the user for explicit consent before executing a
// destructive or irreversible operation against the production environment.
// If force is true, the prompt is skipped entirely — intended for CI/CD pipelines.
// Returns an error if the user declines.
func ConfirmProduction(operation string, force bool) error {
	if force {
		// --force bypasses the interactive prompt for non-interactive contexts (e.g. CI).
		style.PrintWarning("Production guardrail bypassed via --force flag.")
		return nil
	}

	style.PrintWarning("You are about to run '%s' against the PRODUCTION environment.", operation)
	fmt.Fprintf(os.Stderr, "%s ", style.Warning.Render("Are you sure? [y/N]:"))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input != "y" && input != "yes" {
		return fmt.Errorf("operation cancelled by user")
	}

	return nil
}
