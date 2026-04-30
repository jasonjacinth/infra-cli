package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/jasonjacinth/infra-cli/internal/guardrail"
	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/jasonjacinth/infra-cli/internal/style"
	"github.com/spf13/cobra"
)

// chaosCmd is the parent command for chaos engineering operations.
var chaosCmd = &cobra.Command{
	Use:   "chaos",
	Short: "Chaos engineering tools for resilience testing",
	Long: `Inject controlled failures into your Kubernetes workloads to verify
that self-healing mechanisms (ReplicaSets, liveness probes, readiness probes)
are working correctly.

Available subcommands:
  pod-kill    Randomly kill a pod belonging to an application`,
}

// podKillCmd kills a randomly selected pod for a given application.
var podKillCmd = &cobra.Command{
	Use:   "pod-kill",
	Short: "Kill a random pod to test self-healing",
	Long: `Delete a randomly selected pod belonging to the target application.
Kubernetes should automatically recreate the pod via the ReplicaSet controller.
This verifies that your deployment's self-healing behavior is working.

The command waits after deletion and reports whether recovery succeeded.

Examples:
  infra-cli chaos pod-kill --app nginx -e production
  infra-cli chaos pod-kill --app nginx -e production --force`,
	Run: runPodKill,
}

func init() {
	rootCmd.AddCommand(chaosCmd)
	chaosCmd.AddCommand(podKillCmd)

	podKillCmd.Flags().StringP("app", "a", "", "Name of the application to target")
	podKillCmd.MarkFlagRequired("app")
}

func runPodKill(cmd *cobra.Command, args []string) {
	env, _ := cmd.Flags().GetString("environment")
	app, _ := cmd.Flags().GetString("app")
	namespace, _ := cmd.Flags().GetString("namespace")
	force, _ := cmd.Flags().GetBool("force")

	if !shell.IsInstalled("kubectl") {
		style.PrintError("kubectl is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	if err := guardrail.CheckNamespace(namespace); err != nil {
		style.PrintError("%s", err)
		os.Exit(1)
	}

	if env == "production" {
		if err := guardrail.ConfirmProduction("chaos pod-kill", force); err != nil {
			style.PrintWarning("%s", err)
			os.Exit(0)
		}
	}

	output, err := shell.Run("kubectl", "get", "pods", "-l", fmt.Sprintf("app=%s", app), "-n", namespace, "--no-headers", "-o", "custom-columns=:metadata.name")
	if err != nil {
		style.PrintError("Failed to list pods for app '%s'.", app)
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		os.Exit(1)
	}

	pods := parsePodNames(output)
	if len(pods) == 0 {
		style.PrintError("No pods found for app '%s' in namespace '%s'.", app, namespace)
		os.Exit(1)
	}

	target := pods[rand.Intn(len(pods))]

	style.PrintHeader("Chaos Engineering: Pod Kill")
	fmt.Println(style.Divider())
	fmt.Printf("  Target app:   %s\n", app)
	fmt.Printf("  Namespace:    %s\n", namespace)
	fmt.Printf("  Pods found:   %d\n", len(pods))
	fmt.Printf("  Victim pod:   %s\n", style.Warning.Render(target))
	fmt.Println()

	style.PrintInfo("Killing pod '%s'...", target)
	_, err = shell.Run("kubectl", "delete", "pod", target, "-n", namespace)
	if err != nil {
		style.PrintError("Failed to delete pod '%s'.", target)
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		os.Exit(1)
	}

	style.PrintSuccess("Pod '%s' deleted.", target)
	fmt.Println()

	style.PrintInfo("Waiting 5 seconds for Kubernetes to recreate the pod...")
	time.Sleep(5 * time.Second)

	style.PrintHeader("Recovery Status")
	fmt.Println(style.Divider())

	recoveryOutput, err := shell.Run("kubectl", "get", "pods", "-l", fmt.Sprintf("app=%s", app), "-n", namespace)
	if err != nil {
		style.PrintWarning("Could not check recovery status: %s", err)
		os.Exit(1)
	}

	fmt.Println(recoveryOutput)
	fmt.Println()

	newPods := parsePodNames(recoveryOutput)
	if len(newPods) >= len(pods) {
		style.PrintSuccess("Self-healing verified: Kubernetes recreated the pod. %d/%d pods running.", len(newPods), len(pods))
	} else {
		style.PrintWarning("Recovery may still be in progress. Expected %d pods, found %d. Re-check with 'infra-cli status -e production'.", len(pods), len(newPods))
	}
}

func parsePodNames(output string) []string {
	var pods []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		name := strings.TrimSpace(line)
		if name != "" {
			pods = append(pods, name)
		}
	}
	return pods
}
