package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/spf13/cobra"
)

// logsCmd represents the logs command.
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Tail logs for an application",
	Long: `Stream logs from a running Kubernetes pod.
Automatically resolves the pod name from the application name.

Examples:
  infra-cli logs --app nginx
  infra-cli logs --app nginx -e production`,
	Run: runLogs,
}

func init() {
	rootCmd.AddCommand(logsCmd)

	// Local flags for the logs command.
	logsCmd.Flags().StringP("app", "a", "", "Name of the application to tail logs for")
	logsCmd.MarkFlagRequired("app")
}

func runLogs(cmd *cobra.Command, args []string) {
	env, _ := cmd.Flags().GetString("environment")
	app, _ := cmd.Flags().GetString("app")

	if !shell.IsInstalled("kubectl") {
		fmt.Fprintln(os.Stderr, "❌ kubectl is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	switch env {
	case "local", "production":
		runK8sLogs(app)
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown environment: %s (use 'local' or 'production')\n", env)
		os.Exit(1)
	}
}

func runK8sLogs(app string) {
	// Auto-resolve: find a pod whose name starts with the app name.
	podName, err := findPodByApp(app)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("📋 Tailing logs for pod '%s'...\n\n", podName)

	// Stream logs directly to the terminal (Ctrl+C to stop).
	err = shell.RunLive("kubectl", "logs", "-f", podName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Failed to tail logs for pod '%s'.\n   %s\n", podName, err)
		os.Exit(1)
	}
}

// findPodByApp looks up a running pod whose name starts with the given app name.
// This abstracts away the need to remember full pod names like "my-service-7d4b8c6f5-xk9zn".
func findPodByApp(app string) (string, error) {
	output, err := shell.Run("kubectl", "get", "pods", "--no-headers", "-o", "custom-columns=:metadata.name")
	if err != nil {
		return "", fmt.Errorf("failed to list pods: %s", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if strings.HasPrefix(name, app) {
			return name, nil
		}
	}

	return "", fmt.Errorf("no pod found matching app name '%s'. Run 'infra-cli status -e production' to see available pods.", app)
}
