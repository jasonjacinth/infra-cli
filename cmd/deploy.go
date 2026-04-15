package cmd

import (
	"fmt"
	"os"

	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command.
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy an application to the target environment",
	Long: `Deploy an application to Kubernetes using the manifests in k8s/.

For local: applies k8s/deployment.yaml and k8s/service.yaml to your local cluster.
For production: applies k8s/deployment.yaml to the production cluster.

Examples:
  infra-cli deploy --app nginx
  infra-cli deploy --app nginx -e production`,
	Run: runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)

	// Local flags for the deploy command.
	deployCmd.Flags().StringP("app", "a", "", "Name of the application to deploy")
	deployCmd.MarkFlagRequired("app")
}

func runDeploy(cmd *cobra.Command, args []string) {
	env, _ := cmd.Flags().GetString("environment")
	app, _ := cmd.Flags().GetString("app")

	if !shell.IsInstalled("kubectl") {
		fmt.Fprintln(os.Stderr, "kubectl is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	switch env {
	case "local":
		runLocalDeploy(app)
	case "production":
		runK8sDeploy(app)
	default:
		fmt.Fprintf(os.Stderr, "Unknown environment: %s (use 'local' or 'production')\n", env)
		os.Exit(1)
	}
}

func runLocalDeploy(app string) {
	fmt.Printf("Deploying '%s' to local Kubernetes cluster...\n\n", app)

	output, err := shell.Run("kubectl", "apply", "-f", "k8s/deployment.yaml", "-f", "k8s/service.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Local deploy failed.\n   %s\n", err)
		fmt.Fprintln(os.Stderr, "\n   Make sure k8s/deployment.yaml and k8s/service.yaml exist and your cluster is running.")
		os.Exit(1)
	}

	fmt.Println(output)
	fmt.Printf("\n'%s' deployed to local cluster.\n", app)
}

func runK8sDeploy(app string) {
	manifest := "k8s/deployment.yaml"
	fmt.Printf("Deploying '%s' to Kubernetes (production) using %s...\n\n", app, manifest)

	output, err := shell.Run("kubectl", "apply", "-f", manifest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Kubernetes deploy failed.\n   %s\n", err)
		fmt.Fprintf(os.Stderr, "\n   Make sure '%s' exists and your cluster is reachable.\n", manifest)
		os.Exit(1)
	}

	fmt.Println(output)
	fmt.Printf("\n'%s' deployed to Kubernetes.\n", app)
}
