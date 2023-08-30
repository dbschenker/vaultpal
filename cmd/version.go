package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime"
)

func newVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version of vaultpal",
		Long: `Print version of vaultpal.

vaultpal version`,
		Run: func(cmd *cobra.Command, args []string) {
			var msg = fmt.Sprintf("%s (commit: %s), built %s, platform %s/%s",
				Version, Commit, BuildDate, runtime.GOOS, runtime.GOARCH)
			println(msg)
		},
	}

	return versionCmd
}

func init() {
	rootCmd.AddCommand(newVersionCmd())
}
