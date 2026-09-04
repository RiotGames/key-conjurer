package command

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"log/slog"

	"github.com/coreos/go-oidc"
	"github.com/pkg/browser"
	"github.com/riotgames/key-conjurer/pkg/oauth2cli"
	"github.com/urfave/cli/v3"
	"golang.org/x/oauth2"
	"golang.org/x/term"
)

func init() {
	// Silence stdout/stderr from browsers
	browser.Stdout = io.Discard
	browser.Stderr = io.Discard
}

var (
	FlagURLOnly   = "url-only"
	FlagNoBrowser = "no-browser"
)

var loginCmd = &cli.Command{
	Name:        "login",
	Usage:       "Authenticate with KeyConjurer.",
	Description: "Login to KeyConjurer using OAuth2. You will be required to open the URL printed to the console or scan a QR code.",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var loginCmd LoginCommand
		if err := loginCmd.Parse(cmd, nil); err != nil {
			return err
		}

		return loginCmd.Execute(ctx, ConfigFromContext(ctx))
	},

	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    FlagURLOnly,
			Usage:   "Print only the URL to visit rather than a user-friendly message",
			Aliases: []string{"u"},
		},
		&cli.BoolFlag{
			Name:    FlagNoBrowser,
			Usage:   "Do not open a browser window, printing the URL instead",
			Aliases: []string{"b"},
		},
	},
}

// ShouldUseMachineOutput indicates whether or not we should write to standard output as if the user is a machine.
//
// What this means is implementation specific, but this usually indicates the user is trying to use this program in a script and we should avoid user-friendly output messages associated with values a user might find useful.
func ShouldUseMachineOutput(flags flagSet) bool {
	quiet := flags.Bool(FlagQuiet)
	fi, _ := os.Stdout.Stat()
	isPiped := fi.Mode()&os.ModeCharDevice == 0
	return isPiped || quiet
}

type LoginCommand struct {
	OIDCDomain    string
	ClientID      string
	MachineOutput bool
	NoBrowser     bool
}

type flagSet interface {
	String(name string) string
	Bool(name string) bool
	Uint(name string) uint
}

func (c *LoginCommand) Parse(flags flagSet, args []string) error {
	c.OIDCDomain = flags.String(FlagOIDCDomain)
	c.ClientID = flags.String(FlagClientID)
	c.NoBrowser = flags.Bool(FlagNoBrowser)
	urlOnly := flags.Bool(FlagURLOnly)
	c.MachineOutput = ShouldUseMachineOutput(flags) || urlOnly
	return nil
}

func (c LoginCommand) Execute(ctx context.Context, config *Config) error {
	if checkKeychainLocked() {
		// Don't go through the whole login flow if the keychain is locked, prompt the user to unlock it first
		return &ErrKeychainLocked{}
	}

	serveURL := openBrowserToURL
	if c.NoBrowser {
		if c.MachineOutput {
			serveURL = printURLToConsole
		} else {
			serveURL = friendlyPrintURLToConsole
		}
	}

	prov, err := oidc.NewProvider(ctx, c.OIDCDomain)
	if err != nil {
		return fmt.Errorf("discover provider: %w", err)
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

	cfg := oauth2.Config{
		ClientID:    c.ClientID,
		Endpoint:    prov.Endpoint(),
		Scopes:      []string{oidc.ScopeOpenID, "profile", "okta.apps.read", "okta.apps.sso"},
		RedirectURL: fmt.Sprintf("http://%s", net.JoinHostPort("localhost", port)),
	}

	handler := oauth2cli.NewAuthorizationCodeHandler(&cfg, serveURL)
	accessToken, err := handler.HandlePendingSession(ctx, sock)
	if err != nil {
		return err
	}

	// https://openid.net/specs/openid-connect-core-1_0.html#TokenResponse
	idToken, ok := accessToken.Extra("id_token").(string)
	if !ok {
		return fmt.Errorf("id_token not found in token response")
	}

	_, err = prov.Verifier(&oidc.Config{ClientID: c.ClientID}).Verify(ctx, idToken)
	if err != nil {
		return fmt.Errorf("validate id token: %w", err)
	}

	return putAccountCredentialInKeychain(accessToken, idToken)
}

type ErrNoPortsAvailable struct{}

func (e ErrNoPortsAvailable) Error() string {
	return "no ports available"
}

func (e ErrNoPortsAvailable) ExitCode() int {
	return ExitCodeConnectivityError
}

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

	return nil, &ErrNoPortsAvailable{}
}

func printURLToConsole(url string) error {
	fmt.Fprintln(os.Stdout, url)
	return nil
}

// hyperlink wraps text in a terminal hyperlink pointing at url, so supporting
// terminals (including tmux) render it as a clickable link. When stdout is not a
// terminal, the plain text is returned unchanged so piped or redirected output is
// never corrupted by escape sequences.
func hyperlink(url, text string) string {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return text
	}
	return osc8Hyperlink(url, text)
}

// osc8Hyperlink formats an OSC 8 terminal hyperlink escape sequence:
// ESC ]8;;<url> ESC \ <text> ESC ]8;; ESC \
func osc8Hyperlink(url, text string) string {
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}

func friendlyPrintURLToConsole(url string) error {
	fmt.Printf("Visit the following link in your terminal: %s\n", hyperlink(url, url))
	return nil
}

func openBrowserToURL(url string) error {
	slog.Debug("trying to open browser window", slog.String("url", url))
	return browser.OpenURL(url)
}
