package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "infra-cli",
	Short: "A developer-friendly CLI for managing containers and clusters",
	Long: `Infra-CLI abstracts the complexity of kubectl and docker commands
into a simplified, developer-friendly workflow.

It allows engineers to manage local development containers and remote
Kubernetes clusters through a unified set of high-level commands.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// RootCmd returns the root cobra.Command. It is exported for use in tests
// that need to inspect the registered subcommand tree.
func RootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	// Persistent flags are available to this command and all subcommands.
	rootCmd.PersistentFlags().StringP("environment", "e", "local", "Target environment: local or production")
	rootCmd.PersistentFlags().StringP("namespace", "n", "default", "Kubernetes namespace to operate in (kube-system is restricted)")
	rootCmd.PersistentFlags().Bool("force", false, "Bypass production confirmation prompts (for use in CI/CD pipelines)")
}
