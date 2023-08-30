package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	//  Config json storage location
	keyConjurerRcPath string
	// host of the API server. Don't use this. You probably meant to use newClient() instead.
	host string
	// This is set by the Makefile during build of the CLI. Don't use this.
	defaultHost string
	// config is a cache-like datastore for this application. It is loaded at app start-up.
	config                   Config
	quiet                    bool
	buildTimestamp           string = BuildDate + " " + BuildTime + " " + BuildTimeZone
	cmdShortVersionFlag      bool   = false
	cmdOneLineVersionFlag    bool   = false
	cloudAws                        = "aws"
	cloudTencent                    = "tencent"
	clientHttpTimeoutSeconds int    = 120
)

func init() {
	rootCmd.PersistentFlags().IntVar(&clientHttpTimeoutSeconds, "http-timeout", 120, "the amount of time in seconds to wait for keyconjurer to respond")
	rootCmd.PersistentFlags().StringVar(&keyConjurerRcPath, "keyconjurer-rc-path", "~/.keyconjurerrc", "path to .keyconjurerrc file")
	rootCmd.PersistentFlags().StringVar(&host, "host", defaultHost, "The host of the KeyConjurer API")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "tells the CLI to be quiet; stdout will not contain human-readable informational messages")
	rootCmd.SetVersionTemplate(`{{printf "%s" .Version}}`)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(&switchCmd)
	rootCmd.AddCommand(&aliasCmd)
	rootCmd.AddCommand(&unaliasCmd)
	rootCmd.AddCommand(&rolesCmd)
	rootCmd.Flags().BoolVarP(&cmdShortVersionFlag, "short-version", "s", false, "version for "+appname+" (short format)")
	rootCmd.Flags().BoolVarP(&cmdOneLineVersionFlag, "oneline-version", "1", false, "version for "+appname+" (single line format)")
}

// hack to remove the leading blank line in the --version output
const versionString string = "" +
	"	Version: 		%s\n" +
	"	Build Timestamp:	%s\n" +
	"	Client: 		%s\n" +
	"	Default Hostname:	%s\n" +
	"	Upgrade URL:		%s\n"

func alternateVersions(cmd *cobra.Command, short, oneline bool) {
	if oneline {
		cmd.Printf("%s %s (Build Timestamp:%s - Client:%s)\n", appname, Version, buildTimestamp, ClientName)
	} else {
		cmd.Printf("%s %s\n", appname, Version)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     appname,
	Version: fmt.Sprintf(versionString, Version, buildTimestamp, ClientName, defaultHost, DownloadURL),
	Short:   "Retrieve temporary cloud credentials.",
	Long: `Key Conjurer retrieves temporary credentials from the Key Conjurer API.

To get started run the following commands:
  ` + appname + ` login # You will get prompted for your AD credentials
  ` + appname + ` accounts
  ` + appname + ` get <accountName>
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		fp := keyConjurerRcPath
		if expanded, err := homedir.Expand(fp); err == nil {
			fp = expanded
		}

		if err := os.MkdirAll(filepath.Dir(fp), os.ModeDir|os.FileMode(0755)); err != nil {
			return err
		}

		file, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if os.IsNotExist(err) {
			return nil
		} else if err != nil {
			return err
		}

		return config.Read(file)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if cmdShortVersionFlag || cmdOneLineVersionFlag {
			alternateVersions(cmd, cmdShortVersionFlag, cmdOneLineVersionFlag)
		} else {
			cmd.Help()
		}
	},
	PersistentPostRunE: func(*cobra.Command, []string) error {
		var fp string
		if expanded, err := homedir.Expand(keyConjurerRcPath); err == nil {
			fp = expanded
		}

		dir := filepath.Dir(fp)
		if err := os.MkdirAll(dir, os.ModeDir|os.FileMode(0755)); err != nil {
			return err
		}

		file, err := os.Create(fp)
		if err != nil {
			return fmt.Errorf("unable to create %s reason: %w", fp, err)
		}

		defer file.Close()
		return config.Write(file)
	},
}

var errHostnameCannotContainPath = errors.New("hostname must not contain a path")
