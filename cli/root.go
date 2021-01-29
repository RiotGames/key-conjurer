package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	keyConjurerRcPath string
	// host of the API server. Don't use this. You probably meant to use newClient() instead.
	host string
	// This is set by the Makefile during build of the CLI. Don't use this.
	defaultHost  string
	authProvider string
	// config is a cache-like datastore for this application. It is loaded at app start-up.
	config Config
)

func init() {
	rootCmd.PersistentFlags().StringVar(&keyConjurerRcPath, "keyconjurer-rc-path", "~/.keyconjurerrc", "path to .keyconjurerrc file")
	rootCmd.PersistentFlags().StringVar(&host, "host", defaultHost, "The host of the KeyConjurer API")
	rootCmd.SetVersionTemplate(`{{printf "%s" .Version}}`)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(rolesCmd)
	rootCmd.AddCommand(providersCmd)
}

const versionString string = `
	Version: 		%s
	Client: 		%s
	Default Hostname:	%s
	Upgrade URL:		%s
`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "keyconjurer",
	Version: fmt.Sprintf(versionString, Version, ClientName, defaultHost, DownloadURL),
	Short:   "Retrieve temporary AWS API credentials.",
	Long: `Key Conjurer retrieves temporary credentials from the Key Conjurer API.

To get started run the following commands:
keyconjurer login # You will get prompted for your AD credentials
keyconjurer accounts
keyconjurer get <accountName>
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		host, err = parseHostname(host)
		if err != nil {
			return fmt.Errorf("invalid hostname: %w", err)
		}

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

func parseHostname(hostname string) (string, error) {
	uri, err := url.Parse(hostname)
	// Sometimes url.Parse is not smart enough to return an error but fails parsing all the same.
	// This enables us to self-heal if the user passes something like "idp.example.com" or "idp.example.com:4000"
	if err != nil {
		return "", err
	}

	// This indicate the user passed a URL with a path & a port *or* a hostname with a path and neither specified scheme.
	if strings.Contains(uri.Opaque, "/") || strings.Contains(uri.Path, "/") {
		return "", errHostnameCannotContainPath
	}

	// If the user passes something like foo.example.com, this will all be dumped inside the Path
	if uri.Host == "" && uri.Path != "" {
		uri.Scheme = "http"
		uri.Host = uri.Path
		uri.Path = ""
	}

	// If the user passes something that has the format %s:%d, Go is going to interpret %s as being the scheme and %d being the opaque portion.
	if uri.Opaque != "" && uri.Host == "" {
		uri.Host = net.JoinHostPort(uri.Scheme, uri.Opaque)
		uri.Scheme = "http"
		uri.Opaque = ""
	}

	if uri.Host == "" || err != nil {
		return "", err
	}

	if uri.Path != "" && uri.Path != "/" {
		return "", errHostnameCannotContainPath
	}

	return uri.String(), nil
}

func newClient() (Client, error) {
	return NewClient(host)
}
