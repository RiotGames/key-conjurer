package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

var (
	//  Config json storage location
	// host of the API server. Don't use this. You probably meant to use newClient() instead.
	host string
	// This is set by the Makefile during build of the CLI. Don't use this.
	defaultHost      string
	identityProvider string
	// config is a cache-like datastore for this application. It is loaded at app start-up.
	config                Config
	quiet                 bool
	buildTimestamp        string = BuildDate + " " + BuildTime + " " + BuildTimeZone
	cmdShortVersionFlag   bool   = false
	cmdOneLineVersionFlag bool   = false
	cloudAws                     = "aws"
	cloudTencent                 = "tencent"
)

func init() {
	rootCmd.PersistentFlags().StringVar(&host, "host", defaultHost, "The host of the KeyConjurer API")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "tells the CLI to be quiet; stdout will not contain human-readable informational messages")
	rootCmd.SetVersionTemplate(`{{printf "%s" .Version}}`)
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
		var err error
		host, err = parseHostname(host)
		if err != nil {
			return fmt.Errorf("invalid hostname: %w", err)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if cmdShortVersionFlag || cmdOneLineVersionFlag {
			alternateVersions(cmd, cmdShortVersionFlag, cmdOneLineVersionFlag)
		} else {
			cmd.Help()
		}
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
