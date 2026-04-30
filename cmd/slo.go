package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/jasonjacinth/infra-cli/internal/style"
	"github.com/spf13/cobra"
)

const (
	sloMaxRestarts       = 5
	sloMinStabilityMins  = 5
)

// sloCmd is the parent command for SLO operations.
var sloCmd = &cobra.Command{
	Use:   "slo",
	Short: "Service Level Objective validation",
	Long: `Validate that your running workloads meet basic Service Level Objectives.

Available subcommands:
  validate    Check availability, restart budget, and pod stability`,
}

// sloValidateCmd checks three SLO targets against live cluster data.
var sloValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate SLOs for an application",
	Long: `Run three SLO checks against a running application:

  1. Availability   All desired replicas are running and ready.
  2. Restart Budget  No pod has restarted more than 5 times.
  3. Pod Stability   All pods have been running for at least 5 minutes.

Examples:
  infra-cli slo validate --app nginx -e production
  infra-cli slo validate --app nginx`,
	Run: runSLOValidate,
}

func init() {
	rootCmd.AddCommand(sloCmd)
	sloCmd.AddCommand(sloValidateCmd)

	sloValidateCmd.Flags().StringP("app", "a", "", "Name of the application to validate")
	sloValidateCmd.MarkFlagRequired("app")
}

// Kubernetes JSON structures used by kubectl -o json output.
type k8sDeployment struct {
	Spec struct {
		Replicas int `json:"replicas"`
	} `json:"spec"`
	Status struct {
		ReadyReplicas     int `json:"readyReplicas"`
		AvailableReplicas int `json:"availableReplicas"`
	} `json:"status"`
}

type k8sPodList struct {
	Items []k8sPod `json:"items"`
}

type k8sPod struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Status struct {
		Phase             string             `json:"phase"`
		StartTime         string             `json:"startTime"`
		ContainerStatuses []k8sContainerStatus `json:"containerStatuses"`
	} `json:"status"`
}

type k8sContainerStatus struct {
	RestartCount int  `json:"restartCount"`
	Ready        bool `json:"ready"`
}

func runSLOValidate(cmd *cobra.Command, args []string) {
	app, _ := cmd.Flags().GetString("app")
	namespace, _ := cmd.Flags().GetString("namespace")

	if !shell.IsInstalled("kubectl") {
		style.PrintError("kubectl is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	style.PrintHeader("SLO Validation: %s", app)
	fmt.Println(style.Divider())
	fmt.Println()

	passed := 0
	failed := 0

	if checkAvailabilitySLO(app, namespace) {
		passed++
	} else {
		failed++
	}

	if checkRestartBudgetSLO(app, namespace) {
		passed++
	} else {
		failed++
	}

	if checkPodStabilitySLO(app, namespace) {
		passed++
	} else {
		failed++
	}

	fmt.Println()
	fmt.Println(style.Divider())
	if failed == 0 {
		style.PrintSuccess("SLO Summary: %d/%d checks passed. All SLOs met.", passed, passed+failed)
	} else {
		style.PrintError("SLO Summary: %d/%d checks passed. %d SLO(s) violated.", passed, passed+failed, failed)
		os.Exit(1)
	}
}

// checkAvailabilitySLO verifies that all desired replicas are running and ready.
func checkAvailabilitySLO(app, namespace string) bool {
	output, err := shell.Run("kubectl", "get", "deployment", app, "-n", namespace, "-o", "json")
	if err != nil {
		style.PrintError("  [FAIL] Availability: could not fetch deployment '%s'.", app)
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("         %s", err)))
		return false
	}

	var deploy k8sDeployment
	if err := json.Unmarshal([]byte(output), &deploy); err != nil {
		style.PrintError("  [FAIL] Availability: could not parse deployment JSON.")
		return false
	}

	desired := deploy.Spec.Replicas
	ready := deploy.Status.ReadyReplicas

	if ready >= desired {
		style.PrintSuccess("  [PASS] Availability: %d/%d replicas ready.", ready, desired)
		return true
	}

	style.PrintError("  [FAIL] Availability: %d/%d replicas ready (expected %d).", ready, desired, desired)
	return false
}

// checkRestartBudgetSLO verifies that no pod has exceeded the restart threshold.
func checkRestartBudgetSLO(app, namespace string) bool {
	output, err := shell.Run("kubectl", "get", "pods", "-l", fmt.Sprintf("app=%s", app), "-n", namespace, "-o", "json")
	if err != nil {
		style.PrintError("  [FAIL] Restart Budget: could not fetch pods for '%s'.", app)
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("         %s", err)))
		return false
	}

	var podList k8sPodList
	if err := json.Unmarshal([]byte(output), &podList); err != nil {
		style.PrintError("  [FAIL] Restart Budget: could not parse pod JSON.")
		return false
	}

	maxRestarts := 0
	offender := ""

	for _, pod := range podList.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.RestartCount > maxRestarts {
				maxRestarts = cs.RestartCount
				offender = pod.Metadata.Name
			}
		}
	}

	if maxRestarts <= sloMaxRestarts {
		style.PrintSuccess("  [PASS] Restart Budget: max restarts = %d (threshold: %d).", maxRestarts, sloMaxRestarts)
		return true
	}

	style.PrintError("  [FAIL] Restart Budget: pod '%s' has %d restarts (threshold: %d).", offender, maxRestarts, sloMaxRestarts)
	return false
}

// checkPodStabilitySLO verifies that all pods have been running for at least
// the minimum stability duration.
func checkPodStabilitySLO(app, namespace string) bool {
	output, err := shell.Run("kubectl", "get", "pods", "-l", fmt.Sprintf("app=%s", app), "-n", namespace, "-o", "json")
	if err != nil {
		style.PrintError("  [FAIL] Pod Stability: could not fetch pods for '%s'.", app)
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("         %s", err)))
		return false
	}

	var podList k8sPodList
	if err := json.Unmarshal([]byte(output), &podList); err != nil {
		style.PrintError("  [FAIL] Pod Stability: could not parse pod JSON.")
		return false
	}

	threshold := time.Duration(sloMinStabilityMins) * time.Minute
	now := time.Now()

	for _, pod := range podList.Items {
		if pod.Status.StartTime == "" {
			style.PrintError("  [FAIL] Pod Stability: pod '%s' has no start time (may be pending).", pod.Metadata.Name)
			return false
		}

		startTime, err := time.Parse(time.RFC3339, pod.Status.StartTime)
		if err != nil {
			style.PrintError("  [FAIL] Pod Stability: could not parse start time for pod '%s'.", pod.Metadata.Name)
			return false
		}

		age := now.Sub(startTime)
		if age < threshold {
			style.PrintError("  [FAIL] Pod Stability: pod '%s' has only been running for %s (minimum: %s).", pod.Metadata.Name, age.Round(time.Second), threshold)
			return false
		}
	}

	style.PrintSuccess("  [PASS] Pod Stability: all pods running for at least %d minutes.", sloMinStabilityMins)
	return true
}
