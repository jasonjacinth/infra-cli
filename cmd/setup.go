package cmd

import (
	"fmt"

	"github.com/jasonjacinth/infra-cli/internal/shell"
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
	fmt.Println("🔍 Checking system dependencies...\n")

	allGood := true

	// --- Docker ---
	if shell.IsInstalled("docker") {
		version, err := shell.Run("docker", "--version")
		if err != nil {
			fmt.Println("  ⚠️  Docker is installed but not responding. Is Docker Desktop running?")
			allGood = false
		} else {
			fmt.Printf("  ✅ Docker:  %s\n", version)
		}
	} else {
		fmt.Println("  ❌ Docker:  not found in PATH")
		fmt.Println("     ➜  Install from https://docs.docker.com/get-docker/")
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
			fmt.Println("  ⚠️  kubectl is installed but returned an error.")
			allGood = false
		} else {
			fmt.Printf("  ✅ kubectl: %s\n", version)
		}
	} else {
		fmt.Println("  ❌ kubectl: not found in PATH")
		fmt.Println("     ➜  Install from https://kubernetes.io/docs/tasks/tools/")
		allGood = false
	}

	// --- Summary ---
	fmt.Println()
	if allGood {
		fmt.Println("🎉 All dependencies are installed. You're ready to go!")
	} else {
		fmt.Println("⚠️  Some dependencies are missing. Please install them and re-run 'infra-cli setup'.")
	}
}
