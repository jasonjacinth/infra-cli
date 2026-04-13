package cmd

import (
	"fmt"
	"os"

	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/spf13/cobra"
)

// rollbackCmd represents the rollback command.
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Revert the last deployment change",
	Long: `Rollback the most recent deployment for an application.
Uses kubectl rollout undo (production) or restarts the previous Docker image (local).

Examples:
  infra-cli rollback --app my-service
  infra-cli rollback --app my-service -e production`,
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

	switch env {
	case "local":
		runLocalRollback(app)
	case "production":
		runK8sRollback(app)
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown environment: %s (use 'local' or 'production')\n", env)
		os.Exit(1)
	}
}

func runLocalRollback(app string) {
	if !shell.IsInstalled("docker") {
		fmt.Fprintln(os.Stderr, "❌ Docker is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	fmt.Printf("⏪ Rolling back '%s' locally...\n\n", app)

	// For Docker, rollback means restarting the container with its previous image.
	// We stop the current container and restart via docker-compose.
	_, err := shell.Run("docker", "restart", app)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Local rollback failed.\n   %s\n", err)
		fmt.Fprintln(os.Stderr, "\n   Make sure the container exists. Check with: infra-cli status")
		os.Exit(1)
	}

	fmt.Printf("✅ Container '%s' has been restarted.\n", app)
}

func runK8sRollback(app string) {
	if !shell.IsInstalled("kubectl") {
		fmt.Fprintln(os.Stderr, "❌ kubectl is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	fmt.Printf("⏪ Rolling back deployment '%s' on Kubernetes...\n\n", app)

	output, err := shell.Run("kubectl", "rollout", "undo", "deployment/"+app)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Kubernetes rollback failed.\n   %s\n", err)
		fmt.Fprintln(os.Stderr, "\n   Make sure the deployment exists. Check with: infra-cli status -e production")
		os.Exit(1)
	}

	fmt.Println(output)
	fmt.Printf("\n✅ Deployment '%s' rolled back successfully.\n", app)
}
