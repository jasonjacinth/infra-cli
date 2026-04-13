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

func init() {
	// Persistent flags are available to this command and all subcommands.
	rootCmd.PersistentFlags().StringP("environment", "e", "local", "Target environment: local or production")
}
