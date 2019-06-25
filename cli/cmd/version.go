package cmd

import (
	"fmt"
	"keyconjurer-cli/keyconjurer"

	"github.com/spf13/cobra"
)

const versionString string = `    Version:      %s
    Client:       %s
    Prod URL:     %s
    Dev  URL:     %s
    Upgrade URL:  %s
`

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Shows Key Conjurer version information.",
	Long:    "Shows Key Conjurer version information.",
	Example: "keyconjurer version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(versionString, keyconjurer.Version, keyconjurer.Client, keyconjurer.ProdAPI, keyconjurer.DevAPI, keyconjurer.DownloadURL)
	}}
