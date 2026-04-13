package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the health status of running services",
	Long: `Display the health and status of containers (local) or pods (production).

Examples:
  infra-cli status
  infra-cli status --app my-service
  infra-cli status -e production`,
	Run: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Local flags for the status command.
	statusCmd.Flags().StringP("app", "a", "", "Filter status by application name (optional)")
}

func runStatus(cmd *cobra.Command, args []string) {
	env, _ := cmd.Flags().GetString("environment")
	app, _ := cmd.Flags().GetString("app")

	switch env {
	case "local":
		runLocalStatus(app)
	case "production":
		runK8sStatus(app)
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown environment: %s (use 'local' or 'production')\n", env)
		os.Exit(1)
	}
}

func runLocalStatus(app string) {
	if !shell.IsInstalled("docker") {
		fmt.Fprintln(os.Stderr, "❌ Docker is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	fmt.Println("📦 Docker Container Status")
	fmt.Println(strings.Repeat("─", 50))

	output, err := shell.Run("docker", "ps", "--format", "table {{.Names}}\t{{.Status}}")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to get Docker status. Is Docker Desktop running?\n   %s\n", err)
		os.Exit(1)
	}

	if app != "" {
		// Filter output lines to show header + lines matching the app name.
		lines := strings.Split(output, "\n")
		if len(lines) > 0 {
			fmt.Println(lines[0]) // Header row
		}
		found := false
		for _, line := range lines[1:] {
			if strings.Contains(line, app) {
				fmt.Println(line)
				found = true
			}
		}
		if !found {
			fmt.Printf("\nNo containers found matching '%s'.\n", app)
		}
	} else {
		fmt.Println(output)
	}
}

func runK8sStatus(app string) {
	if !shell.IsInstalled("kubectl") {
		fmt.Fprintln(os.Stderr, "❌ kubectl is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	fmt.Println("☸️  Kubernetes Pod Status")
	fmt.Println(strings.Repeat("─", 50))

	var output string
	var err error

	if app != "" {
		// Use label selector if app name is provided.
		output, err = shell.Run("kubectl", "get", "pods", "-l", fmt.Sprintf("app=%s", app))
	} else {
		output, err = shell.Run("kubectl", "get", "pods")
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to get pod status.\n   %s\n", err)
		os.Exit(1)
	}

	if output == "" {
		fmt.Println("No pods found.")
	} else {
		fmt.Println(output)
	}
}
