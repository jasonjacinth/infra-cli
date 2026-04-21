package guardrail_test

import (
	"testing"

	"github.com/jasonjacinth/infra-cli/internal/guardrail"
)

// TestCheckNamespace_AllowedNamespace verifies that a standard app namespace passes.
func TestCheckNamespace_AllowedNamespace(t *testing.T) {
	allowed := []string{"default", "staging", "production", "my-app"}
	for _, ns := range allowed {
		if err := guardrail.CheckNamespace(ns); err != nil {
			t.Errorf("expected namespace '%s' to be allowed, got error: %v", ns, err)
		}
	}
}

// TestCheckNamespace_RestrictedNamespaces verifies that cluster-critical namespaces are blocked.
func TestCheckNamespace_RestrictedNamespaces(t *testing.T) {
	restricted := []string{"kube-system", "kube-public", "kube-node-lease"}
	for _, ns := range restricted {
		err := guardrail.CheckNamespace(ns)
		if err == nil {
			t.Errorf("expected namespace '%s' to be blocked, but CheckNamespace returned nil", ns)
		}
	}
}

// TestCheckNamespace_ErrorMessageContainsNamespace verifies the error message is actionable.
func TestCheckNamespace_ErrorMessageContainsNamespace(t *testing.T) {
	err := guardrail.CheckNamespace("kube-system")
	if err == nil {
		t.Fatal("expected an error for kube-system, got nil")
	}

	msg := err.Error()
	if !contains(msg, "kube-system") {
		t.Errorf("expected error message to contain the namespace name, got: %q", msg)
	}
	if !contains(msg, "Guardrail Violation") {
		t.Errorf("expected error message to contain 'Guardrail Violation', got: %q", msg)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
