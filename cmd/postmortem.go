package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/jasonjacinth/infra-cli/internal/style"
	"github.com/spf13/cobra"
)

var validSeverities = map[string]bool{
	"critical": true,
	"major":    true,
	"minor":    true,
}

// postmortemCmd is the parent command for postmortem operations.
var postmortemCmd = &cobra.Command{
	Use:   "postmortem",
	Short: "Blameless postmortem document management",
	Long: `Generate structured, blameless postmortem documents following SRE best
practices. The generated document includes sections for timeline, impact analysis,
root cause (5 Whys), action items, and lessons learned.

Available subcommands:
  create    Generate a new postmortem document`,
}

// postmortemCreateCmd generates a blameless postmortem markdown file.
var postmortemCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate a blameless postmortem document",
	Long: `Create a structured postmortem markdown document pre-populated with:
  - Incident metadata (title, severity, date)
  - Blameless postmortem template sections
  - Auto-captured cluster context (pod status, recent events, deployments)

The cluster context capture is optional and degrades gracefully if the cluster
is unreachable.

Examples:
  infra-cli postmortem create --title "API Latency Spike" --severity critical
  infra-cli postmortem create --title "Pod OOMKill" --severity major -e production
  infra-cli postmortem create --title "DNS Timeout" --severity minor --output my-postmortem.md`,
	Run: runPostmortemCreate,
}

func init() {
	rootCmd.AddCommand(postmortemCmd)
	postmortemCmd.AddCommand(postmortemCreateCmd)

	postmortemCreateCmd.Flags().String("title", "", "Incident title (required)")
	postmortemCreateCmd.MarkFlagRequired("title")
	postmortemCreateCmd.Flags().String("severity", "", "Incident severity: critical, major, or minor (required)")
	postmortemCreateCmd.MarkFlagRequired("severity")
	postmortemCreateCmd.Flags().String("output", "", "Output file path (optional, auto-generated if omitted)")
}

func runPostmortemCreate(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	severity, _ := cmd.Flags().GetString("severity")
	outputPath, _ := cmd.Flags().GetString("output")
	namespace, _ := cmd.Flags().GetString("namespace")

	if !validSeverities[strings.ToLower(severity)] {
		style.PrintError("Invalid severity: '%s'. Must be one of: critical, major, minor.", severity)
		os.Exit(1)
	}
	severity = strings.ToLower(severity)

	now := time.Now()
	dateStr := now.Format("2006-01-02")
	timestampStr := now.Format("2006-01-02 15:04:05 MST")

	if outputPath == "" {
		slug := slugify(title)
		outputPath = fmt.Sprintf("postmortem-%s-%s.md", dateStr, slug)
	}

	style.PrintInfo("Generating postmortem document...")
	fmt.Println()

	clusterContext := captureClusterContext(namespace)

	content := buildPostmortemDocument(title, severity, dateStr, timestampStr, clusterContext)

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		style.PrintError("Failed to write postmortem file: %s", err)
		os.Exit(1)
	}

	absPath, _ := filepath.Abs(outputPath)
	style.PrintSuccess("Postmortem document created: %s", absPath)
}

func buildPostmortemDocument(title, severity, dateStr, timestampStr, clusterContext string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Postmortem: %s\n\n", title))
	b.WriteString("---\n\n")

	b.WriteString("## Incident Metadata\n\n")
	b.WriteString(fmt.Sprintf("| Field       | Value      |\n"))
	b.WriteString(fmt.Sprintf("|-------------|------------|\n"))
	b.WriteString(fmt.Sprintf("| **Title**   | %s         |\n", title))
	b.WriteString(fmt.Sprintf("| **Date**    | %s         |\n", dateStr))
	b.WriteString(fmt.Sprintf("| **Severity**| %s         |\n", strings.ToUpper(severity)))
	b.WriteString(fmt.Sprintf("| **Author**  | [your name]|\n"))
	b.WriteString(fmt.Sprintf("| **Status**  | Draft      |\n"))
	b.WriteString("\n---\n\n")

	b.WriteString("## Summary\n\n")
	b.WriteString("<!-- 2-3 sentence description of what happened, when, and the user impact. -->\n\n")
	b.WriteString("[TODO: Describe the incident in plain language.]\n\n")
	b.WriteString("---\n\n")

	b.WriteString("## Timeline\n\n")
	b.WriteString(fmt.Sprintf("| Time | Event |\n"))
	b.WriteString(fmt.Sprintf("|------|-------|\n"))
	b.WriteString(fmt.Sprintf("| %s | Incident began / alert fired |\n", timestampStr))
	b.WriteString("| [TODO] | First responder acknowledged |\n")
	b.WriteString("| [TODO] | Root cause identified |\n")
	b.WriteString("| [TODO] | Mitigation applied |\n")
	b.WriteString("| [TODO] | Service fully recovered |\n")
	b.WriteString("| [TODO] | Postmortem review meeting |\n")
	b.WriteString("\n---\n\n")

	b.WriteString("## Impact\n\n")
	b.WriteString("- **Duration**: [TODO: How long was the service degraded?]\n")
	b.WriteString("- **Users affected**: [TODO: Number or percentage of affected users.]\n")
	b.WriteString("- **Revenue impact**: [TODO: Estimated financial impact, if any.]\n")
	b.WriteString("- **SLO impact**: [TODO: Did this consume error budget? How much?]\n")
	b.WriteString("\n---\n\n")

	b.WriteString("## Root Cause Analysis (5 Whys)\n\n")
	b.WriteString("<!-- Ask \"why\" iteratively to find the true root cause, not just the symptom. -->\n\n")
	b.WriteString("1. **Why** did the service degrade?\n")
	b.WriteString("   - [TODO]\n")
	b.WriteString("2. **Why** did that happen?\n")
	b.WriteString("   - [TODO]\n")
	b.WriteString("3. **Why** did that happen?\n")
	b.WriteString("   - [TODO]\n")
	b.WriteString("4. **Why** did that happen?\n")
	b.WriteString("   - [TODO]\n")
	b.WriteString("5. **Why** did that happen?\n")
	b.WriteString("   - [TODO: This should be the root cause.]\n")
	b.WriteString("\n---\n\n")

	b.WriteString("## What Went Well\n\n")
	b.WriteString("- [TODO: What detection, response, or tooling worked as expected?]\n")
	b.WriteString("- [TODO]\n")
	b.WriteString("\n---\n\n")

	b.WriteString("## What Went Wrong\n\n")
	b.WriteString("- [TODO: What failed, was slow, or was missing?]\n")
	b.WriteString("- [TODO]\n")
	b.WriteString("\n---\n\n")

	b.WriteString("## Action Items\n\n")
	b.WriteString("| Action | Owner | Priority | Deadline |\n")
	b.WriteString("|--------|-------|----------|----------|\n")
	b.WriteString("| [TODO: Preventive action] | [owner] | P1 | [date] |\n")
	b.WriteString("| [TODO: Detection improvement] | [owner] | P2 | [date] |\n")
	b.WriteString("| [TODO: Process improvement] | [owner] | P3 | [date] |\n")
	b.WriteString("\n---\n\n")

	b.WriteString("## Lessons Learned\n\n")
	b.WriteString("- [TODO: What did the team learn from this incident?]\n")
	b.WriteString("- [TODO: What would you do differently next time?]\n")
	b.WriteString("\n---\n\n")

	b.WriteString("## Cluster Context Snapshot\n\n")
	b.WriteString(fmt.Sprintf("Captured at: %s\n\n", timestampStr))
	if clusterContext != "" {
		b.WriteString(clusterContext)
	} else {
		b.WriteString("_Cluster was not reachable at the time of document generation._\n")
	}

	return b.String()
}

// captureClusterContext gathers pod status, recent events, and deployment
// state from the cluster. Returns empty string if the cluster is unreachable.
func captureClusterContext(namespace string) string {
	if !shell.IsInstalled("kubectl") {
		return ""
	}

	var b strings.Builder

	podOutput, err := shell.Run("kubectl", "get", "pods", "-n", namespace)
	if err == nil && podOutput != "" {
		b.WriteString("### Pod Status\n\n")
		b.WriteString("```\n")
		b.WriteString(podOutput)
		b.WriteString("\n```\n\n")
	}

	eventOutput, err := shell.Run("kubectl", "get", "events", "--sort-by=.lastTimestamp", "-n", namespace, "--no-headers")
	if err == nil && eventOutput != "" {
		lines := strings.Split(eventOutput, "\n")
		limit := 10
		if len(lines) < limit {
			limit = len(lines)
		}
		b.WriteString("### Recent Events (last 10)\n\n")
		b.WriteString("```\n")
		b.WriteString(strings.Join(lines[len(lines)-limit:], "\n"))
		b.WriteString("\n```\n\n")
	}

	deployOutput, err := shell.Run("kubectl", "get", "deployments", "-n", namespace)
	if err == nil && deployOutput != "" {
		b.WriteString("### Deployments\n\n")
		b.WriteString("```\n")
		b.WriteString(deployOutput)
		b.WriteString("\n```\n\n")
	}

	return b.String()
}

// slugify converts a title string into a URL-safe, lowercase slug.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")

	var cleaned strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleaned.WriteRune(r)
		}
	}

	result := cleaned.String()
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	return strings.Trim(result, "-")
}
