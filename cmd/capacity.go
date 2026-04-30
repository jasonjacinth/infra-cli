package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/jasonjacinth/infra-cli/internal/style"
	"github.com/spf13/cobra"
)

// capacityCmd displays resource usage, requests, and limits for an application.
var capacityCmd = &cobra.Command{
	Use:   "capacity",
	Short: "Analyze resource usage and capacity for an application",
	Long: `Display a table of CPU and memory usage alongside configured requests and
limits for each pod in a deployment. Shows utilization percentages to help
with capacity planning decisions.

Actual usage data requires the Kubernetes metrics-server to be running.
If metrics-server is not available, only requests and limits are shown.

Examples:
  infra-cli capacity --app nginx -e production
  infra-cli capacity --app nginx`,
	Run: runCapacity,
}

func init() {
	rootCmd.AddCommand(capacityCmd)

	capacityCmd.Flags().StringP("app", "a", "", "Name of the application to analyze")
	capacityCmd.MarkFlagRequired("app")
}

type k8sDeploymentFull struct {
	Spec struct {
		Replicas int `json:"replicas"`
		Template struct {
			Spec struct {
				Containers []k8sContainer `json:"containers"`
			} `json:"spec"`
		} `json:"template"`
	} `json:"spec"`
}

type k8sContainer struct {
	Name      string         `json:"name"`
	Resources k8sResources   `json:"resources"`
}

type k8sResources struct {
	Requests k8sResourceValues `json:"requests"`
	Limits   k8sResourceValues `json:"limits"`
}

type k8sResourceValues struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type podUsage struct {
	Name   string
	CPUm   int
	MemMi  int
}

func runCapacity(cmd *cobra.Command, args []string) {
	app, _ := cmd.Flags().GetString("app")
	namespace, _ := cmd.Flags().GetString("namespace")

	if !shell.IsInstalled("kubectl") {
		style.PrintError("kubectl is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	style.PrintHeader("Capacity Analysis: %s", app)
	fmt.Println(style.Divider())
	fmt.Println()

	output, err := shell.Run("kubectl", "get", "deployment", app, "-n", namespace, "-o", "json")
	if err != nil {
		style.PrintError("Could not fetch deployment '%s'.", app)
		fmt.Fprintln(os.Stderr, style.Subtle.Render(fmt.Sprintf("  %s", err)))
		os.Exit(1)
	}

	var deploy k8sDeploymentFull
	if err := json.Unmarshal([]byte(output), &deploy); err != nil {
		style.PrintError("Could not parse deployment JSON: %s", err)
		os.Exit(1)
	}

	if len(deploy.Spec.Template.Spec.Containers) == 0 {
		style.PrintError("No containers found in deployment '%s'.", app)
		os.Exit(1)
	}

	container := deploy.Spec.Template.Spec.Containers[0]
	cpuReq := container.Resources.Requests.CPU
	cpuLim := container.Resources.Limits.CPU
	memReq := container.Resources.Requests.Memory
	memLim := container.Resources.Limits.Memory

	fmt.Printf("  Deployment:     %s\n", app)
	fmt.Printf("  Replicas:       %d\n", deploy.Spec.Replicas)
	fmt.Printf("  Container:      %s\n", container.Name)
	fmt.Printf("  CPU Request:    %s\n", cpuReq)
	fmt.Printf("  CPU Limit:      %s\n", cpuLim)
	fmt.Printf("  Memory Request: %s\n", memReq)
	fmt.Printf("  Memory Limit:   %s\n", memLim)
	fmt.Println()

	usageData, metricsAvailable := getActualUsage(app, namespace)

	if metricsAvailable && len(usageData) > 0 {
		cpuReqM := parseMillicores(cpuReq)
		cpuLimM := parseMillicores(cpuLim)
		memReqMi := parseMebibytes(memReq)
		memLimMi := parseMebibytes(memLim)

		fmt.Printf("  %-40s %10s %10s %10s %10s %10s %10s\n",
			"POD", "CPU USED", "CPU REQ", "CPU LIM", "MEM USED", "MEM REQ", "MEM LIM")
		fmt.Printf("  %s\n", strings.Repeat("-", 100))

		var totalCPU, totalMem int

		for _, pu := range usageData {
			fmt.Printf("  %-40s %9dm %9dm %9dm %8dMi %8dMi %8dMi\n",
				truncate(pu.Name, 40), pu.CPUm, cpuReqM, cpuLimM, pu.MemMi, memReqMi, memLimMi)
			totalCPU += pu.CPUm
			totalMem += pu.MemMi
		}

		fmt.Println()
		podCount := len(usageData)

		if cpuReqM > 0 {
			avgCPU := totalCPU / podCount
			cpuUtilReq := (avgCPU * 100) / cpuReqM
			cpuUtilLim := 0
			if cpuLimM > 0 {
				cpuUtilLim = (avgCPU * 100) / cpuLimM
			}
			style.PrintInfo("  CPU utilization (avg): %d%% of request, %d%% of limit", cpuUtilReq, cpuUtilLim)
		}

		if memReqMi > 0 {
			avgMem := totalMem / podCount
			memUtilReq := (avgMem * 100) / memReqMi
			memUtilLim := 0
			if memLimMi > 0 {
				memUtilLim = (avgMem * 100) / memLimMi
			}
			style.PrintInfo("  Mem utilization (avg): %d%% of request, %d%% of limit", memUtilReq, memUtilLim)
		}
	} else {
		style.PrintWarning("Actual usage data is unavailable (metrics-server may not be running).")
		style.PrintSubtle("  To enable: minikube addons enable metrics-server")
		style.PrintSubtle("  Or on Docker Desktop: metrics-server is included by default.")
		fmt.Println()
		style.PrintInfo("Showing configured requests and limits only (see above).")
	}

	fmt.Println()
	style.PrintSuccess("Capacity analysis complete.")
}

// getActualUsage runs kubectl top pods and parses the output.
// Returns the per-pod usage data and whether metrics were available.
func getActualUsage(app, namespace string) ([]podUsage, bool) {
	output, err := shell.Run("kubectl", "top", "pods", "-l", fmt.Sprintf("app=%s", app), "-n", namespace, "--no-headers")
	if err != nil {
		return nil, false
	}

	var results []podUsage
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		results = append(results, podUsage{
			Name:  fields[0],
			CPUm:  parseMillicores(fields[1]),
			MemMi: parseMebibytes(fields[2]),
		})
	}

	return results, true
}

// parseMillicores converts a Kubernetes CPU string (e.g. "100m", "0.5", "1")
// to integer millicores.
func parseMillicores(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if strings.HasSuffix(s, "m") {
		val, err := strconv.Atoi(strings.TrimSuffix(s, "m"))
		if err != nil {
			return 0
		}
		return val
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int(val * 1000)
}

// parseMebibytes converts a Kubernetes memory string (e.g. "128Mi", "1Gi", "256000Ki")
// to integer mebibytes.
func parseMebibytes(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if strings.HasSuffix(s, "Gi") {
		val, err := strconv.Atoi(strings.TrimSuffix(s, "Gi"))
		if err != nil {
			return 0
		}
		return val * 1024
	}
	if strings.HasSuffix(s, "Mi") {
		val, err := strconv.Atoi(strings.TrimSuffix(s, "Mi"))
		if err != nil {
			return 0
		}
		return val
	}
	if strings.HasSuffix(s, "Ki") {
		val, err := strconv.Atoi(strings.TrimSuffix(s, "Ki"))
		if err != nil {
			return 0
		}
		return val / 1024
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val / (1024 * 1024)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
