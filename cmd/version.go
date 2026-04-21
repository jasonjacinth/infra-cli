package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Build-time variables injected via ldflags during compilation.
// These are populated by the Makefile using:
//
//	go build -ldflags "-X cmd.Version=... -X cmd.GitCommit=... -X cmd.BuildTime=..."
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version, build commit, and build time",
	Long: `Display build metadata for infra-cli.

Version, Git commit hash, and build timestamp are injected at compile
time via linker flags. If you are running from source, values will
show as "dev" / "unknown".

Example:
  infra-cli version`,
	Run: runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("infra-cli %s\n", Version)
	fmt.Printf("  commit:    %s\n", GitCommit)
	fmt.Printf("  built:     %s\n", BuildTime)
	fmt.Printf("  go:        %s\n", runtime.Version())
	fmt.Printf("  os/arch:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
