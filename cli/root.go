package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	//  Config json storage location
	// host of the API server. Don't use this. You probably meant to use newClient() instead.
	host  string
	quiet bool
	// This is set by the Makefile during build of the CLI. Don't use this.
	defaultHost      string
	identityProvider string
	// config is a cache-like datastore for this application. It is loaded at app start-up.
	config         Config
	buildTimestamp string = strings.Join([]string{BuildDate, BuildTime, BuildTimeZone}, " ")
	cloudAws              = "aws"
	cloudTencent          = "tencent"
)

func init() {
	pflag.StringVar(&host, "host", defaultHost, "The host of the KeyConjurer API")
	pflag.BoolVar(&quiet, "quiet", false, "tells the CLI to be quiet; stdout will not contain human-readable informational messages")
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(&switchCmd)
	rootCmd.AddCommand(&providersCmd)
	rootCmd.AddCommand(&aliasCmd)
	rootCmd.AddCommand(&unaliasCmd)
	rootCmd.AddCommand(&rolesCmd)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     appname,
	Version: fmt.Sprintf("%s %s (Build Timestamp: %s - Client: %s)", appname, Version, buildTimestamp, ClientName),
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
		var err error
		host, err = parseHostname(host)
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

func newClient() (Client, error) {
	return NewClient(host)
}
