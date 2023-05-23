package cmd

import (
	"github.com/dbschenker/vaultpal/token"
	"github.com/spf13/cobra"
)

func newSwitchCmd() *cobra.Command {
	switchCmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch between roles",
		Long: `Switch between vault roles.

Switch requires a subcommand like roles, e.g.:

vaultpal switch role k8s-admin`,
		RunE: nil,
	}

	switchCmd.AddCommand(
		&cobra.Command{
			Use:   "role",
			Short: "Switch to a token role",
			Long: `Switch with current token to another token role.

Requires 1 arguments: [role-name]
`,
			Args: cobra.ExactArgs(1),
			Example: `  # Switch with current token to another token role [k8s-admin]
  vaultpal switch role k8s-admin`,
			RunE: func(cmd *cobra.Command, args []string) error {
				return token.SwitchRole(args[0])

			}})

	return switchCmd
}

func init() {
	rootCmd.AddCommand(newSwitchCmd())
}
