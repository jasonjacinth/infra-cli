package cmd

import (
	"fmt"
	"os"

	"github.com/jasonjacinth/infra-cli/internal/guardrail"
	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/jasonjacinth/infra-cli/internal/style"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command.
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy an application to the target environment",
	Long: `Deploy an application to Kubernetes using the manifests in k8s/.

For local: applies k8s/deployment.yaml and k8s/service.yaml to your local cluster.
For production: applies k8s/deployment.yaml to the production cluster.

Deploying to production requires confirmation unless --force is passed.

Examples:
  infra-cli deploy --app nginx
  infra-cli deploy --app nginx -e production
  infra-cli deploy --app nginx -e production -n staging
  infra-cli deploy --app nginx -e production --force`,
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
	namespace, _ := cmd.Flags().GetString("namespace")
	force, _ := cmd.Flags().GetBool("force")

	if !shell.IsInstalled("kubectl") {
		style.PrintError("kubectl is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	// Guardrail: block operations against restricted namespaces.
	if err := guardrail.CheckNamespace(namespace); err != nil {
		style.PrintError("%s", err)
		os.Exit(1)
	}

	switch env {
	case "local":
		runLocalDeploy(app, namespace)
	case "production":
		// Safety: require explicit confirmation before mutating production.
		if err := guardrail.ConfirmProduction("deploy", force); err != nil {
			style.PrintWarning("%s", err)
			os.Exit(0)
		}
		runK8sDeploy(app, namespace)
	default:
		style.PrintError("Unknown environment: %s (use 'local' or 'production')", env)
		os.Exit(1)
	}
}

func runLocalDeploy(app, namespace string) {
	style.PrintInfo("Deploying '%s' to local Kubernetes cluster (namespace: %s)...", app, namespace)
	fmt.Println()

	output, err := shell.Run("kubectl", "apply", "-f", "k8s/deployment.yaml", "-f", "k8s/service.yaml", "-n", namespace)
	if err != nil {
		style.PrintError("Local deploy failed.")
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		fmt.Fprintln(os.Stderr, style.Subtle.Render("\n  Make sure k8s/deployment.yaml and k8s/service.yaml exist and your cluster is running."))
		os.Exit(1)
	}

	fmt.Println(output)
	style.PrintSuccess("'%s' deployed to local cluster.", app)
}

func runK8sDeploy(app, namespace string) {
	manifest := "k8s/deployment.yaml"
	style.PrintInfo("Deploying '%s' to Kubernetes (production) using %s (namespace: %s)...", app, manifest, namespace)
	fmt.Println()

	output, err := shell.Run("kubectl", "apply", "-f", manifest, "-n", namespace)
	if err != nil {
		style.PrintError("Kubernetes deploy failed.")
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("\n  Make sure '%s' exists and your cluster is reachable.", manifest)))
		os.Exit(1)
	}

	fmt.Println(output)
	style.PrintSuccess("'%s' deployed to Kubernetes.", app)
}
