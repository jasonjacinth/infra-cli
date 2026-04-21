package cmd

import (
	"fmt"
	"os"

	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/jasonjacinth/infra-cli/internal/style"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command.
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Verify that all required dependencies are installed",
	Long: `Check if Docker and kubectl are installed and available in your PATH.
This helps new team members onboard quickly by surfacing missing tools upfront.

Example:
  infra-cli setup`,
	Run: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) {
	style.PrintHeader("Checking system dependencies...")
	fmt.Println()

	allGood := true

	// --- Docker ---
	if shell.IsInstalled("docker") {
		version, err := shell.Run("docker", "--version")
		if err != nil {
			style.PrintWarning("  Docker is installed but not responding. Is Docker Desktop running?")
			allGood = false
		} else {
			fmt.Printf("  %s  %s\n", style.Success.Render("Docker: "), version)
		}
	} else {
		style.PrintError("  Docker:  not found in PATH")
		fmt.Fprintln(os.Stderr, style.Subtle.Render("     ->  Install from https://docs.docker.com/get-docker/"))
		allGood = false
	}

	// --- kubectl ---
	if shell.IsInstalled("kubectl") {
		version, err := shell.Run("kubectl", "version", "--client", "--short")
		if err != nil {
			// --short may not be supported in newer versions, fall back.
			version, err = shell.Run("kubectl", "version", "--client")
		}
		if err != nil {
			style.PrintWarning("  kubectl is installed but returned an error.")
			allGood = false
		} else {
			fmt.Printf("  %s %s\n", style.Success.Render("kubectl:"), version)
		}
	} else {
		style.PrintError("  kubectl: not found in PATH")
		fmt.Fprintln(os.Stderr, style.Subtle.Render("     ->  Install from https://kubernetes.io/docs/tasks/tools/"))
		allGood = false
	}

	// --- Summary ---
	fmt.Println()
	if allGood {
		style.PrintSuccess("All dependencies are installed. You're ready to go!")
	} else {
		style.PrintWarning("Some dependencies are missing. Please install them and re-run 'infra-cli setup'.")
	}
}
