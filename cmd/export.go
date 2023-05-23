package cmd

import (
	"errors"
	"github.com/dbschenker/vaultpal/aws"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var bashAWSSTSAlias = `function _vpalsts(){pal_sts_result=$(vaultpal export awssts $1); if [ $? -eq 0 ]; then echo "STS success"; eval "$pal_sts_result"; else echo "--- FAILED STS ---"; echo "${pal_sts_result}"; fi};_vpalsts`

func newExportCmd() *cobra.Command {
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export various types of resources to shell",
		Long: `Export various types of resources to shell

Write requires a subcommand like awssts, e.g.:

vaultpal export awssts tsc-vpc-manager`,
		Run: nil,
	}

	awsstsCmd := &cobra.Command{
		Use:   "awssts",
		Short: "Export AWS STS credentials to shell",
		Long: `Get AWS STS credentials from vault for an AWS STS role and export them to shell.

Requires 1 argument: [role-name]

Hint: Use a bash alias, that will export the received AWS STS credentials for direct usage. awssts provides an alias
function to use with flag -a/-alias. See examples.
`,
		Args: cobra.MaximumNArgs(1),
		Example: `  # Export AWS STS credentials for engine (aws) with role [tsc-vpc-manager]
  vaultpal export awssts tsc-vpc-manager

  # Print bash alias function to use for awssts command 
  vaultpal export awssts -a
  # or direct alias definition with vaultpal
  alias palsts="$(vaultpal export awssts -a)"`,
		RunE: func(cmd *cobra.Command, args []string) error {

			aliasF, err := cmd.Flags().GetBool("alias")
			if err != nil {
				log.Fatalf("cannot read alias flag: %s", err)
			}
			if aliasF {
				_, _ = os.Stdout.Write([]byte(bashAWSSTSAlias))
				return nil
			}

			if len(args) != 1 {
				return errors.New("missing argument: role name to use for aws sts")
			}
			pathF, err := cmd.Flags().GetString("path")
			if err != nil {
				log.Fatalf("cannot read aws engine path: %s", err)
			}
			return aws.ExportSTSCredentials(pathF, args[0])

		}}

	setAWSEngineFlag(awsstsCmd)
	awsstsCmd.Flags().BoolP("alias", "a", false, "Print an alias function to use in bash for awssts command (default: false)")
	exportCmd.AddCommand(awsstsCmd)

	awsConsoleCmd := &cobra.Command{
		Use:   "awsconsole",
		Short: "Export AWS web console URL to shell",
		Long: `Get AWS STS credentials from vault for an STS role and generate a sign-in URL for the web console using these credentials.

Requires 1 argument: [role-name]
`,
		Args: cobra.MaximumNArgs(1),
		Example: `  # Generate AWS console URL for engine (aws) with role [tsc-vpc-manager]
  vaultpal export awsconsole tsc-vpc-manager`,
		RunE: func(cmd *cobra.Command, args []string) error {

			suppressBrowserF, err := cmd.Flags().GetBool("suppress-open")
			if err != nil {
				log.Fatalf("cannot read suppress-open flag: %s", err)
			}

			if len(args) != 1 {
				return errors.New("missing argument: role name to use for aws sts")
			}
			pathF, err := cmd.Flags().GetString("path")
			if err != nil {
				log.Fatalf("cannot read aws engine path: %s", err)
			}
			return aws.GenerateConsoleURL(pathF, suppressBrowserF, args[0])

		}}
	setAWSEngineFlag(awsConsoleCmd)
	awsConsoleCmd.Flags().BoolP("suppress-open", "s", false, "Suppress opening URL in default browser (default: false)")
	exportCmd.AddCommand(awsConsoleCmd)

	return exportCmd
}

func setAWSEngineFlag(awsCmd *cobra.Command) *string {
	return awsCmd.Flags().StringP("path", "p", "aws", "Name of the vault secret engine to be used for creating aws sts credentials (default: aws)")
}

func init() {
	rootCmd.AddCommand(newExportCmd())
}
