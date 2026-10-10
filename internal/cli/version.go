package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "1.0.0"
	GitCommit = "HEAD"
	BuildDate = "2026-10-10"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print PingForge version, commit, and runtime details",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("PingForge v%s (%s)\n", Version, GitCommit)
		fmt.Printf("  Built    : %s\n", BuildDate)
		fmt.Printf("  Go       : %s\n", runtime.Version())
		fmt.Printf("  OS/Arch  : %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
