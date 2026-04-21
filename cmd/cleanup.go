package cmd

import (
	"fmt"
	"os"

	"github.com/jasonjacinth/infra-cli/internal/guardrail"
	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/jasonjacinth/infra-cli/internal/style"
	"github.com/spf13/cobra"
)

// cleanupCmd represents the cleanup command.
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove deployed workloads from the cluster",
	Long: `Delete all Kubernetes resources defined in the k8s/ manifests directory.
This is useful for tearing down test or sample workloads.

Cleanup in production requires confirmation unless --force is passed.

Examples:
  infra-cli cleanup
  infra-cli cleanup --dir k8s/
  infra-cli cleanup -e production --force`,
	Run: runCleanup,
}

func init() {
	rootCmd.AddCommand(cleanupCmd)

	cleanupCmd.Flags().StringP("dir", "d", "k8s/", "Path to the manifests directory to delete")
}

func runCleanup(cmd *cobra.Command, args []string) {
	if !shell.IsInstalled("kubectl") {
		style.PrintError("kubectl is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	env, _ := cmd.Flags().GetString("environment")
	namespace, _ := cmd.Flags().GetString("namespace")
	force, _ := cmd.Flags().GetBool("force")
	dir, _ := cmd.Flags().GetString("dir")

	// Guardrail: block operations against restricted namespaces.
	if err := guardrail.CheckNamespace(namespace); err != nil {
		style.PrintError("%s", err)
		os.Exit(1)
	}

	// Safety: cleanup is always destructive — require confirmation for production.
	if env == "production" {
		if err := guardrail.ConfirmProduction("cleanup", force); err != nil {
			style.PrintWarning("%s", err)
			os.Exit(0)
		}
	}

	style.PrintInfo("Removing resources defined in %s (namespace: %s)...", dir, namespace)
	fmt.Println()

	output, err := shell.Run("kubectl", "delete", "-f", dir, "-n", namespace)
	if err != nil {
		style.PrintError("Cleanup failed.")
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("\n  Make sure '%s' exists and the resources were previously deployed.", dir)))
		os.Exit(1)
	}

	fmt.Println(output)
	style.PrintSuccess("Cleanup complete. All resources removed.")
}
