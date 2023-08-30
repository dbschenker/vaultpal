package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const custom_completion = `__vaultpal_parse_list() {
    local vaultpal_output out
    if vaultpal_output=$(vault list $1 2>/dev/null); then
        local vaultout=(${vaultpal_output})
        COMPREPLY=( $( compgen -W "${vaultout[*]:2}" -- "$cur" ) )
    fi
}

__vaultpal_list_cluster_pki_roles() {
    local vault_pki clustername
    clustername=$1
    if vault_pki=$(vault kv get -field=pki kv/vaultpal/k8s/clusters/"${clustername}" 2>/dev/null); then
        if vaultpal_output=$(vault list "${vault_pki}"/roles 2>/dev/null); then
          local vaultout=(${vaultpal_output})
          COMPREPLY=( $( compgen -W "${vaultout[*]:2}" -- "$cur" ) )
        fi
    fi
}

__vaultpal_list_kubeconfig() {
    if [[ ${#nouns[@]} -eq 0 ]]; then
        __vaultpal_list_items "kv/metadata/vaultpal/k8s/clusters"
    elif [[ ${#nouns[@]} -eq 1 ]]; then
        __vaultpal_list_cluster_pki_roles ${nouns[0]}
    else
        return 0
    fi
}

__vaultpal_list_items() {
    __vaultpal_parse_list $1
    if [[ $? -eq 0 ]]; then
        return 0
    fi
}

__vaultpal_custom_func() {
    case ${last_command} in
        vaultpal_export_awssts)
            if [[ ${#nouns[@]} -ge 1 ]]; then
              return
            fi
            __vaultpal_list_items "aws/roles"
            return
            ;;
        vaultpal_switch_role)
            if [[ ${#nouns[@]} -ge 1 ]]; then
              return
            fi
            __vaultpal_list_items "auth/token/roles"
            return
            ;;
        vaultpal_write_kubeconfig)
            __vaultpal_list_kubeconfig
            return
            ;;
        *)
            ;;
    esac
}`

var cfgFile string

// The verbose flag value
var v string

// Version can be set with go build -ldflags="-X github.com/dbschenker/vaultpal/cmd.Version=<VERSION>"
var Version = "v.latest"
var Commit = "unknown"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:                    "vaultpal",
	Short:                  "vault~Pal 👍 will assist you",
	Long:                   `vault~Pal 👍 will help you using vault in your daily work.`,
	Run:                    runHelp,
	BashCompletionFunction: custom_completion,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vaultpal.yaml)")

	/*
		// Cobra also supports local flags, which will only run
		// when this action is called directly.
		rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	*/

	// logging
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := setUpLogs(os.Stdout, v); err != nil {
			return err
		}
		return nil
	}
	rootCmd.PersistentFlags().StringVarP(&v, "verbosity", "v", logrus.InfoLevel.String(), "Log level (debug, info, warn, error, fatal, panic")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".vaultpal" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".vaultpal")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func printWelcome() {
	var msg = fmt.Sprintf("%s\n%s %s %s\n", "          { vault~Pal 👍 }", "         °- (🕶 )", Version, "-°")
	println(msg)
}

func runHelp(cmd *cobra.Command, args []string) {
	printWelcome()
	cmd.Help()
}

// setUpLogs set the log output ans the log level
func setUpLogs(out io.Writer, level string) error {
	logrus.SetOutput(out)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	return nil
}
