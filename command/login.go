package command

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"

	"log/slog"

	"github.com/pkg/browser"
	"github.com/riotgames/key-conjurer/oauth2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	FlagURLOnly   = "url-only"
	FlagNoBrowser = "no-browser"
)

func init() {
	loginCmd.Flags().BoolP(FlagURLOnly, "u", false, "Print only the URL to visit rather than a user-friendly message")
	loginCmd.Flags().BoolP(FlagNoBrowser, "b", false, "Do not open a browser window, printing the URL instead")
}

// ShouldUseMachineOutput indicates whether or not we should write to standard output as if the user is a machine.
//
// What this means is implementation specific, but this usually indicates the user is trying to use this program in a script and we should avoid user-friendly output messages associated with values a user might find useful.
func ShouldUseMachineOutput(flags *pflag.FlagSet) bool {
	quiet, _ := flags.GetBool(FlagQuiet)
	fi, _ := os.Stdout.Stat()
	isPiped := fi.Mode()&os.ModeCharDevice == 0
	return isPiped || quiet
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with KeyConjurer.",
	Long:  "Login to KeyConjurer using OAuth2. You will be required to open the URL printed to the console or scan a QR code.",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := ConfigFromCommand(cmd)
		if !HasTokenExpired(config.Tokens) {
			return nil
		}

		oidcDomain, _ := cmd.Flags().GetString(FlagOIDCDomain)
		clientID, _ := cmd.Flags().GetString(FlagClientID)
		urlOnly, _ := cmd.Flags().GetBool(FlagURLOnly)
		noBrowser, _ := cmd.Flags().GetBool(FlagNoBrowser)
		command := LoginCommand{
			Config:        config,
			OIDCDomain:    oidcDomain,
			ClientID:      clientID,
			MachineOutput: ShouldUseMachineOutput(cmd.Flags()) || urlOnly,
			NoBrowser:     noBrowser,
		}

		return command.Execute(cmd.Context())
	},
}

type LoginCommand struct {
	Config        *Config
	OIDCDomain    string
	ClientID      string
	MachineOutput bool
	NoBrowser     bool
}

func (c LoginCommand) Execute(ctx context.Context) error {
	oauthCfg, err := oauth2.DiscoverConfig(ctx, c.OIDCDomain, c.ClientID)
	if err != nil {
		return err
	}

	sock, err := findFirstFreePort(ctx, "127.0.0.1", CallbackPorts)
	if err != nil {
		return err
	}
	defer sock.Close()
	_, port, err := net.SplitHostPort(sock.Addr().String())
	if err != nil {
		// Failed to split the host and port. We need the port to continue, so bail
		return err
	}
	oauthCfg.RedirectURL = fmt.Sprintf("http://%s", net.JoinHostPort("localhost", port))

	handler := oauth2.RedirectionFlowHandler{
		Config:       oauthCfg,
		OnDisplayURL: openBrowserToURL,
	}

	if c.NoBrowser {
		if c.MachineOutput {
			handler.OnDisplayURL = printURLToConsole
		} else {
			handler.OnDisplayURL = friendlyPrintURLToConsole
		}
	}

	accessToken, err := handler.HandlePendingSession(ctx, sock, oauth2.GeneratePkceChallenge(), oauth2.GenerateState())
	if err != nil {
		return err
	}

	// https://openid.net/specs/openid-connect-core-1_0.html#TokenResponse
	idToken, ok := accessToken.Extra("id_token").(string)
	if !ok {
		return fmt.Errorf("id_token not found in token response")
	}

	return c.Config.SaveOAuthToken(accessToken, idToken)
}

var ErrNoPortsAvailable = errors.New("no ports available")

// findFirstFreePort will attempt to open a network listener for each port in turn, and return the first one that succeeded.
//
// If none succeed, ErrNoPortsAvailable is returned.
//
// This is useful for supporting OIDC servers that do not allow for ephemeral ports to be used in the loopback address, like Okta.
func findFirstFreePort(ctx context.Context, broadcastAddr string, ports []string) (net.Listener, error) {
	var lc net.ListenConfig
	for _, port := range ports {
		addr := net.JoinHostPort(broadcastAddr, port)
		slog.Debug("opening connection", slog.String("addr", addr))
		sock, err := lc.Listen(ctx, "tcp4", addr)
		if err == nil {
			slog.Debug("listening", slog.String("addr", addr))
			return sock, nil
		}
		slog.Debug("could not listen, trying a different addr", slog.String("addr", addr), slog.String("error", err.Error()))
	}

	return nil, ErrNoPortsAvailable
}

func printURLToConsole(url string) error {
	fmt.Fprintln(os.Stdout, url)
	return nil
}

func friendlyPrintURLToConsole(url string) error {
	fmt.Printf("Visit the following link in your terminal: %s\n", url)
	return nil
}

func openBrowserToURL(url string) error {
	slog.Debug("trying to open browser window", slog.String("url", url))
	return browser.OpenURL(url)
}
