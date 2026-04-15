package cmd

import (
	"fmt"
	"os"

	"github.com/jasonjacinth/infra-cli/internal/shell"
	"github.com/spf13/cobra"
)

// cleanupCmd represents the cleanup command.
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove deployed workloads from the cluster",
	Long: `Delete all Kubernetes resources defined in the k8s/ manifests directory.
This is useful for tearing down test or sample workloads.

Examples:
  infra-cli cleanup
  infra-cli cleanup --dir k8s/`,
	Run: runCleanup,
}

func init() {
	rootCmd.AddCommand(cleanupCmd)

	cleanupCmd.Flags().StringP("dir", "d", "k8s/", "Path to the manifests directory to delete")
}

func runCleanup(cmd *cobra.Command, args []string) {
	if !shell.IsInstalled("kubectl") {
		fmt.Fprintln(os.Stderr, "kubectl is not installed. Run 'infra-cli setup' to check dependencies.")
		os.Exit(1)
	}

	dir, _ := cmd.Flags().GetString("dir")

	fmt.Printf("Removing resources defined in %s...\n\n", dir)

	output, err := shell.Run("kubectl", "delete", "-f", dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cleanup failed.\n   %s\n", err)
		fmt.Fprintf(os.Stderr, "\n   Make sure '%s' exists and the resources were previously deployed.\n", dir)
		os.Exit(1)
	}

	fmt.Println(output)
	fmt.Println("\nCleanup complete. All resources removed.")
}
