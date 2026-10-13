package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nimbleflux/fluxbase/internal/buildinfo"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show CLI version information",
	Long:  `Display the version, commit hash, and build date of the Fluxbase CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Fluxbase CLI %s\n", buildinfo.Version())
		fmt.Printf("Commit: %s\n", buildinfo.Commit())
		fmt.Printf("Build Date: %s\n", buildinfo.BuildDate())
	},
}
