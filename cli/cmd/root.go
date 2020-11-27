package cmd

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/riotgames/key-conjurer/cli/keyconjurer"

	"github.com/spf13/cobra"
)

var (
	keyConjurerRcPath string
	// The hostname of the API server. Don't use this. You probably meant to use newClient() instead.
	hostname string
	// This is set by the Makefile during build of the CLI. Don't use this.
	defaultHostname string
	authProvider    string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&keyConjurerRcPath, "keyconjurer-rc-path", "~/.keyconjurerrc", "path to .keyconjurerrc file")
	rootCmd.SetVersionTemplate(`{{printf "%s" .Version}}`)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(aliasCmd)
	rootCmd.AddCommand(unaliasCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(rolesCmd)
	rootCmd.AddCommand(providersCmd)

	rootCmd.PersistentFlags().StringVar(&hostname, "hostname", defaultHostname, "The hostname of the KeyConjurer API")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "keyconjurer",
	Version: fmt.Sprintf(versionString, keyconjurer.Version, keyconjurer.ClientName, defaultHostname, keyconjurer.DownloadURL),
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
		hostname, err = parseHostname(hostname)
		if err != nil {
			return fmt.Errorf("invalid hostname: %w", err)
		}

		return nil
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

func newClient() (keyconjurer.Client, error) {
	// hostname is guaranteed to be a valid URL thanks to our code in rootCmd.PersistentPreRunE
	return keyconjurer.New(hostname)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}
