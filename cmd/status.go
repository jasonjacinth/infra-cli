package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/jasonjacinth/infra-cli/internal/style"
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
		style.PrintError("Unknown environment: %s (use 'local' or 'production')", env)
		os.Exit(1)
	}
}

func runLocalStatus(app string) {
	if !shell.IsInstalled("docker") {
		style.PrintError("Docker is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	style.PrintHeader("Docker Container Status")
	fmt.Println(style.Divider())

	output, err := shell.Run("docker", "ps", "--format", "table {{.Names}}\t{{.Status}}")
	if err != nil {
		style.PrintError("Failed to get Docker status. Is Docker Desktop running?")
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
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
			style.PrintWarning("No containers found matching '%s'.", app)
		}
	} else {
		fmt.Println(output)
	}
}

func runK8sStatus(app string) {
	if !shell.IsInstalled("kubectl") {
		style.PrintError("kubectl is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	style.PrintHeader("Kubernetes Pod Status")
	fmt.Println(style.Divider())

	var output string
	var err error

	if app != "" {
		// Use label selector if app name is provided.
		output, err = shell.Run("kubectl", "get", "pods", "-l", fmt.Sprintf("app=%s", app))
	} else {
		output, err = shell.Run("kubectl", "get", "pods")
	}

	if err != nil {
		style.PrintError("Failed to get pod status.")
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		os.Exit(1)
	}

	if output == "" {
		style.PrintWarning("No pods found.")
	} else {
		fmt.Println(output)
	}
}
