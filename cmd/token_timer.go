package cmd

import (
	"github.com/dbschenker/vaultpal/timer"
	"github.com/spf13/cobra"
)

var (
	query bool
	bash  bool
	clear bool
)

func newTimerCmd() *cobra.Command {
	timerCmd := &cobra.Command{
		Use:   "timer",
		Short: "Display the remaining TTL of your vault token",
		Long: `Display the remaining TTL of your vault token.
Put it in your shell prompt to indicate, what vault instance you are currently using 
and how long your current token is valid`,
		Run: func(cmd *cobra.Command, args []string) {
			timer.Timer(bash, query, clear)
		},
	}

	timerCmd.Flags().BoolVarP(&query, "query", "q", false, "print a verbal representation of the TTL status. E.g. \"green\"")
	timerCmd.Flags().BoolVarP(&clear, "clear-cache", "x", false, "clear timer cache e.g. on token renewal")
	timerCmd.Flags().BoolVarP(&bash, "bash", "b", false, `Use in your bash prompt
Example:
source <(vaultpal timer -b)
Or, permanently:
echo "source <(vaultpal timer -b)" >> .bashrc
`)
	return timerCmd
}

func init() {
	rootCmd.AddCommand(newTimerCmd())
}
