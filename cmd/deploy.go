package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jasonjacinth/infra-cli/internal/guardrail"
	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/jasonjacinth/infra-cli/internal/style"
	"github.com/spf13/cobra"
)

// unhealthyPodStates are Kubernetes pod statuses that indicate a failed deployment.
var unhealthyPodStates = []string{
	"CrashLoopBackOff",
	"ImagePullBackOff",
	"ErrImagePull",
	"Error",
	"InvalidImageName",
	"CreateContainerConfigError",
}

// deployCmd represents the deploy command.
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy an application to the target environment",
	Long: `Deploy an application to Kubernetes using the manifests in k8s/.

For local: applies k8s/deployment.yaml and k8s/service.yaml to your local cluster.
For production: applies k8s/deployment.yaml to the production cluster.

Deploying to production requires confirmation unless --force is passed.

Strategies:
  rolling   Default. Applies the manifest and waits for rollout completion.
  canary    Scales to 1 replica first, validates health, then prompts to
            promote to the full replica count or abort with a rollback.

Examples:
  infra-cli deploy --app nginx
  infra-cli deploy --app nginx -e production
  infra-cli deploy --app nginx -e production --strategy canary
  infra-cli deploy --app nginx -e production --force`,
	Run: runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringP("app", "a", "", "Name of the application to deploy")
	deployCmd.MarkFlagRequired("app")
	deployCmd.Flags().String("strategy", "rolling", "Deployment strategy: rolling or canary")
}

func runDeploy(cmd *cobra.Command, args []string) {
	env, _ := cmd.Flags().GetString("environment")
	app, _ := cmd.Flags().GetString("app")
	namespace, _ := cmd.Flags().GetString("namespace")
	force, _ := cmd.Flags().GetBool("force")
	strategy, _ := cmd.Flags().GetString("strategy")

	if strategy != "rolling" && strategy != "canary" {
		style.PrintError("Unknown strategy: %s (use 'rolling' or 'canary')", strategy)
		os.Exit(1)
	}

	if !shell.IsInstalled("kubectl") {
		style.PrintError("kubectl is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	if err := guardrail.CheckNamespace(namespace); err != nil {
		style.PrintError("%s", err)
		os.Exit(1)
	}

	switch env {
	case "local":
		if strategy == "canary" {
			style.PrintWarning("Canary strategy is only supported for production. Falling back to rolling.")
		}
		runLocalDeploy(app, namespace)
	case "production":
		if err := guardrail.ConfirmProduction("deploy", force); err != nil {
			style.PrintWarning("%s", err)
			os.Exit(0)
		}
		if strategy == "canary" {
			runCanaryDeploy(app, namespace)
		} else {
			runK8sDeploy(app, namespace)
		}
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

	if !validateDeployHealth(app, namespace) {
		style.PrintWarning("Auto-rolling back '%s' due to unhealthy pods...", app)
		shell.Run("kubectl", "rollout", "undo", "deployment/"+app, "-n", namespace)
		style.PrintError("Deploy failed health validation. Rollback complete.")
		os.Exit(1)
	}

	style.PrintSuccess("'%s' deployed to local cluster.", app)
}

func runK8sDeploy(app, namespace string) {
	manifest := "k8s/deployment.yaml"
	style.PrintInfo("Deploying '%s' to Kubernetes (production, rolling) using %s (namespace: %s)...", app, manifest, namespace)
	fmt.Println()

	output, err := shell.Run("kubectl", "apply", "-f", manifest, "-n", namespace)
	if err != nil {
		style.PrintError("Kubernetes deploy failed.")
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("\n  Make sure '%s' exists and your cluster is reachable.", manifest)))
		os.Exit(1)
	}

	fmt.Println(output)

	if !validateDeployHealth(app, namespace) {
		style.PrintWarning("Auto-rolling back '%s' due to unhealthy pods...", app)
		shell.Run("kubectl", "rollout", "undo", "deployment/"+app, "-n", namespace)
		style.PrintError("Deploy failed health validation. Rollback complete.")
		os.Exit(1)
	}

	style.PrintSuccess("'%s' deployed to Kubernetes.", app)
}

// runCanaryDeploy implements the canary deployment strategy:
// 1. Capture the current replica count.
// 2. Apply the new manifest.
// 3. Scale to 1 replica (the canary).
// 4. Validate health of the single canary pod.
// 5. Prompt to promote (scale to original) or abort (rollback + restore).
func runCanaryDeploy(app, namespace string) {
	manifest := "k8s/deployment.yaml"
	style.PrintInfo("Deploying '%s' to Kubernetes (production, canary) using %s (namespace: %s)...", app, manifest, namespace)
	fmt.Println()

	originalReplicas := getCurrentReplicaCount(app, namespace)

	output, err := shell.Run("kubectl", "apply", "-f", manifest, "-n", namespace)
	if err != nil {
		style.PrintError("Kubernetes deploy failed.")
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		os.Exit(1)
	}
	fmt.Println(output)

	style.PrintInfo("Scaling to 1 replica for canary validation...")
	_, err = shell.Run("kubectl", "scale", fmt.Sprintf("deployment/%s", app), "--replicas=1", "-n", namespace)
	if err != nil {
		style.PrintError("Failed to scale deployment for canary.")
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		os.Exit(1)
	}

	if !validateDeployHealth(app, namespace) {
		style.PrintWarning("Canary pod is unhealthy. Aborting and rolling back...")
		shell.Run("kubectl", "rollout", "undo", "deployment/"+app, "-n", namespace)
		shell.Run("kubectl", "scale", fmt.Sprintf("deployment/%s", app), fmt.Sprintf("--replicas=%d", originalReplicas), "-n", namespace)
		style.PrintError("Canary deploy aborted. Rolled back to previous version with %d replicas.", originalReplicas)
		os.Exit(1)
	}

	style.PrintSuccess("Canary pod is healthy.")
	fmt.Println()

	if err := guardrail.ConfirmCanaryPromotion(originalReplicas); err != nil {
		style.PrintWarning("Promotion declined. Rolling back canary...")
		shell.Run("kubectl", "rollout", "undo", "deployment/"+app, "-n", namespace)
		shell.Run("kubectl", "scale", fmt.Sprintf("deployment/%s", app), fmt.Sprintf("--replicas=%d", originalReplicas), "-n", namespace)
		style.PrintInfo("Canary aborted. Restored previous version with %d replicas.", originalReplicas)
		os.Exit(0)
	}

	style.PrintInfo("Promoting canary: scaling to %d replicas...", originalReplicas)
	_, err = shell.Run("kubectl", "scale", fmt.Sprintf("deployment/%s", app), fmt.Sprintf("--replicas=%d", originalReplicas), "-n", namespace)
	if err != nil {
		style.PrintError("Failed to promote canary.")
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		os.Exit(1)
	}

	style.PrintSuccess("Canary promoted. '%s' is now running %d replicas with the new configuration.", app, originalReplicas)
}

// validateDeployHealth waits for the rollout to complete and checks pod
// status for unhealthy states. Returns true if all pods are healthy.
func validateDeployHealth(app, namespace string) bool {
	style.PrintInfo("Waiting for rollout to complete...")

	_, err := shell.Run("kubectl", "rollout", "status", fmt.Sprintf("deployment/%s", app), "-n", namespace, "--timeout=120s")
	if err != nil {
		style.PrintError("Rollout did not complete within 120 seconds.")
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		return false
	}

	time.Sleep(3 * time.Second)

	output, err := shell.Run("kubectl", "get", "pods", "-l", fmt.Sprintf("app=%s", app), "-n", namespace, "--no-headers")
	if err != nil {
		style.PrintWarning("Could not verify pod health: %s", err)
		return true
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		for _, badState := range unhealthyPodStates {
			if strings.Contains(line, badState) {
				style.PrintError("Unhealthy pod detected: %s", strings.TrimSpace(line))
				return false
			}
		}
	}

	style.PrintSuccess("All pods are healthy.")
	return true
}

// getCurrentReplicaCount queries the cluster for the current replica count
// of a deployment. Returns 2 as a default if the deployment does not exist.
func getCurrentReplicaCount(app, namespace string) int {
	output, err := shell.Run("kubectl", "get", "deployment", app, "-n", namespace, "-o", "jsonpath={.spec.replicas}")
	if err != nil {
		return 2
	}
	count, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return 2
	}
	return count
}
