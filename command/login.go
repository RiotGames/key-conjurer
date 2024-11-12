package command

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"log/slog"

	"github.com/coreos/go-oidc"
	"github.com/pkg/browser"
	"github.com/riotgames/key-conjurer/oauth2"
)

var (
	FlagURLOnly   = "url-only"
	FlagNoBrowser = "no-browser"
)

func isPiped() bool {
	fi, _ := os.Stdout.Stat()
	return fi.Mode()&os.ModeCharDevice == 0
}

type LoginCommand struct {
	URLOnly bool `help:"Print only the URL to visit rather than a user-friendly message." short:"u"`
	Browser bool `help:"Open the browser to the Okta URL. If false, a URL will be printed to the command line instead." default:"true" negatable:"" short:"b"`
}

func (c LoginCommand) Help() string {
	return "Login to KeyConjurer using OAuth2. You will be required to open the URL printed to the console or scan a QR code."
}

func (c LoginCommand) RunContext(ctx context.Context, globals *Globals, config *Config) error {
	if !HasTokenExpired(config.Tokens) {
		return nil
	}

	client := &http.Client{Transport: LogRoundTripper{http.DefaultTransport}}
	oauthCfg, err := oauth2.DiscoverConfig(oidc.ClientContext(ctx, client), globals.OIDCDomain, globals.ClientID)
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

	if !c.Browser {
		if isPiped() || globals.Quiet {
			handler.OnDisplayURL = printURLToConsole
		} else {
			handler.OnDisplayURL = friendlyPrintURLToConsole
		}
	}

	accessToken, idToken, err := handler.HandlePendingSession(ctx, sock, oauth2.GenerateState())
	if err != nil {
		return err
	}

	return config.SaveOAuthToken(accessToken, idToken)
}

func (c LoginCommand) Run(globals *Globals, config *Config) error {
	return c.RunContext(context.Background(), globals, config)
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
