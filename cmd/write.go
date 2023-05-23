package cmd

import (
	"github.com/dbschenker/vaultpal/aws"
	"github.com/dbschenker/vaultpal/kube"
	"github.com/spf13/cobra"
)

var validArgs = []string{"kubeconfig"}

func newWritCmd() *cobra.Command {
	writeCmd := &cobra.Command{
		Use:   "write",
		Short: "Write various types of resources",
		Long: `Write various types of resources.

Write requires a subcommand like kubeconfig, e.g.:

vaultpal write kubeconfig int webclaims-dev`,
		Run: nil,
	}

	kubeconfigCmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "Write a kubeconfig for a cluster",
		Long: `Write a kubeconfig created with vault secrets to access a kubernetes cluster.

Requires 2 arguments: [cluster-name] [role-name]
`,
		Args: cobra.ExactArgs(2),
		Example: `  # Write kubeconfig for cluster [int] with the vault role [webclaims-dev] (webclaims topic admin in the namespace webclaims-dev)
  vaultpal write kubeconfig int webclaims-dev`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return kube.WriteKubeconfig(args[0], args[1])

		}}
	writeCmd.AddCommand(kubeconfigCmd)

	awscredsCmd := &cobra.Command{
		Use:   "awscreds",
		Short: "Write a awscreds file for the aws-cli",
		Long: `Write a awscreds file created with vault secrets to access the AWS API from the CLI.

Requires 2 arguments: [role-name] [aws-profile-name]
`,
		Args: cobra.ExactArgs(2),
		Example: `  # Write awscreds for role [topic_owner_tsc] with profile [np]
  vaultpal write awscreds topic_owner_tsc np`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return aws.WriteAWSCreds(args[0], args[1])

		}}
	writeCmd.AddCommand(awscredsCmd)

	return writeCmd
}

func init() {
	rootCmd.AddCommand(newWritCmd())
}
