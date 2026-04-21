package cmd

import (
	"fmt"
	"os"

	"github.com/jasonjacinth/infra-cli/internal/guardrail"
	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/jasonjacinth/infra-cli/internal/style"
	"github.com/spf13/cobra"
)

// rollbackCmd represents the rollback command.
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Revert the last deployment change",
	Long: `Rollback the most recent deployment for an application.
Uses kubectl rollout undo (production) or restarts the previous Docker image (local).

Rolling back in production requires confirmation unless --force is passed.

Examples:
  infra-cli rollback --app my-service
  infra-cli rollback --app my-service -e production
  infra-cli rollback --app my-service -e production --force`,
	Run: runRollback,
}

func init() {
	rootCmd.AddCommand(rollbackCmd)

	// Local flags for the rollback command.
	rollbackCmd.Flags().StringP("app", "a", "", "Name of the application to rollback")
	rollbackCmd.MarkFlagRequired("app")
}

func runRollback(cmd *cobra.Command, args []string) {
	env, _ := cmd.Flags().GetString("environment")
	app, _ := cmd.Flags().GetString("app")
	namespace, _ := cmd.Flags().GetString("namespace")
	force, _ := cmd.Flags().GetBool("force")

	// Guardrail: block operations against restricted namespaces.
	if err := guardrail.CheckNamespace(namespace); err != nil {
		style.PrintError("%s", err)
		os.Exit(1)
	}

	switch env {
	case "local":
		runLocalRollback(app)
	case "production":
		// Safety: require explicit confirmation before rolling back production.
		if err := guardrail.ConfirmProduction("rollback", force); err != nil {
			style.PrintWarning("%s", err)
			os.Exit(0)
		}
		runK8sRollback(app, namespace)
	default:
		style.PrintError("Unknown environment: %s (use 'local' or 'production')", env)
		os.Exit(1)
	}
}

func runLocalRollback(app string) {
	if !shell.IsInstalled("docker") {
		style.PrintError("Docker is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	style.PrintInfo("Rolling back '%s' locally...", app)
	fmt.Println()

	// For Docker, rollback means restarting the container with its previous image.
	_, err := shell.Run("docker", "restart", app)
	if err != nil {
		style.PrintError("Local rollback failed.")
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		fmt.Fprintln(os.Stderr, style.Subtle.Render("\n  Make sure the container exists. Check with: infra-cli status"))
		os.Exit(1)
	}

	style.PrintSuccess("Container '%s' has been restarted.", app)
}

func runK8sRollback(app, namespace string) {
	if !shell.IsInstalled("kubectl") {
		style.PrintError("kubectl is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	style.PrintInfo("Rolling back deployment '%s' on Kubernetes (namespace: %s)...", app, namespace)
	fmt.Println()

	output, err := shell.Run("kubectl", "rollout", "undo", "deployment/"+app, "-n", namespace)
	if err != nil {
		style.PrintError("Kubernetes rollback failed.")
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		fmt.Fprintln(os.Stderr, style.Subtle.Render("\n  Make sure the deployment exists. Check with: infra-cli status -e production"))
		os.Exit(1)
	}

	fmt.Println(output)
	style.PrintSuccess("Deployment '%s' rolled back successfully.", app)
}
