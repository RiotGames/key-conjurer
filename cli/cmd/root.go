package cmd

import (
	"fmt"

	"github.com/riotgames/key-conjurer/cli/keyconjurer"

	"github.com/spf13/cobra"
)

var (
	keyConjurerRcPath string
	devFlag           bool
)

func init() {
	rootCmd.PersistentFlags().StringVar(&keyConjurerRcPath, "keyconjurer-rc-path", "~/.keyconjurerrc", "path to .keyconjurerrc file")
	rootCmd.PersistentFlags().BoolVarP(&devFlag, "dev", "d", false, "flag to use dev server")
	rootCmd.SetVersionTemplate(`{{printf "%s" .Version}}`)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(aliasCmd)
	rootCmd.AddCommand(unaliasCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(versionCmd)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "keyconjurer",
	Version: fmt.Sprintf(versionString, keyconjurer.Version, keyconjurer.Client, keyconjurer.ProdAPI, keyconjurer.DevAPI, keyconjurer.DownloadURL),
	Short:   "Retrieve temporary AWS API credentials.",
	Long: `Key Conjurer retrieves temporary credentials from the Key Conjurer API.

To get started run the following commands:
keyconjurer login # You will get prompted for your AD credentials
keyconjurer accounts
keyconjurer get <accountName>
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		keyconjurer.Dev = devFlag
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}
