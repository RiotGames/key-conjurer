package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	rootcerts "github.com/hashicorp/go-rootcerts"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	ClientID   = os.Getenv("OKTA_CLIENT_ID")
	OktaDomain = os.Getenv("OKTA_DOMAIN")
)

func NewOAuth2Client(ctx context.Context, cfg *oauth2.Config, tok *oauth2.Token) *http.Client {
	// Some Darwin systems require certs to be loaded from the system certificate store or attempts to verify SSL certs on internal websites may fail.
	transport := http.DefaultTransport
	if certs, err := rootcerts.LoadSystemCAs(); err == nil {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certs,
			},
		}
	}

	// The following Oauth2 code is copied from the OAuth2 package with modifications to allow us to use our custom transport with root CAs on Darwin systems.
	src := oauth2.ReuseTokenSource(tok, cfg.TokenSource(ctx, tok))
	return &http.Client{
		Transport: &oauth2.Transport{Base: transport, Source: src},
		Timeout:   time.Second * time.Duration(clientHttpTimeoutSeconds),
	}
}

type OAuth2CallbackInfo struct {
	Code             string
	State            string
	Error            string
	ErrorDescription string
}

type OAuth2Listener struct {
	Addr       string
	errCh      chan error
	callbackCh chan OAuth2CallbackInfo
}

func NewOAuth2Listener() OAuth2Listener {
	return OAuth2Listener{
		Addr:       ":8080",
		errCh:      make(chan error),
		callbackCh: make(chan OAuth2CallbackInfo),
	}
}

func ParseCallbackRequest(r *http.Request) (OAuth2CallbackInfo, error) {
	info := OAuth2CallbackInfo{
		Error:            r.FormValue("error"),
		ErrorDescription: r.FormValue("error_description"),
		State:            r.FormValue("state"),
	}

	return info, nil
}

func (o OAuth2Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	info, err := ParseCallbackRequest(r)
	if err == nil {
		// The only errors that might occur would be incorreclty formatted requests, which we will silently drop.
		o.callbackCh <- info
	}
}

func (o OAuth2Listener) Listen(ctx context.Context) {
	server := http.Server{Addr: o.Addr, Handler: o}
	go func() {
		<-ctx.Done()
		server.Close()
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		o.errCh <- err
	}

	close(o.callbackCh)
	close(o.errCh)
}

func (o OAuth2Listener) WaitForAuthorizationCode(ctx context.Context, state string) (string, error) {
	select {
	case info := <-o.callbackCh:
		if info.Error != "" {
			return "", OAuth2Error{Reason: info.Error, Description: info.ErrorDescription}
		}

		return info.Code, nil
	case err := <-o.errCh:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

type OAuth2Error struct {
	Reason      string
	Description string
}

func (e OAuth2Error) Error() string {
	return fmt.Sprintf("oauth2 error: %s (%s)", e.Description, e.Reason)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with KeyConjurer.",
	Long:  "Login using your AD creds. This stores encrypted credentials on the local system.",
	// Example: appname + " login",
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := oidc.NewProvider(cmd.Context(), OktaDomain)
		if err != nil {
			return fmt.Errorf("couldn't discover OIDC configuration for %s: %w", OktaDomain, err)
		}

		oauthCfg := oauth2.Config{
			ClientID: ClientID,
			Endpoint: provider.Endpoint(),
			Scopes:   []string{"openid", "email", "okta.apps.read"},
			// TODO: Only use a redirect URL to localhost if the user has supplied the `--open-browser` flag.
			// If they don't, the default behavior should be to display a QR code that the user should scan and follow the device authorization flow instead.
			RedirectURL: "http://localhost:8080",
		}

		state := "TODO: CREATE RANDOM STATE HERE"
		codeChallenge := "TODO: INSERT CODE CHALLENGE HERE"
		codeVerifier := "TODO: GENERATE CODE VERIFIER HERE"
		url := oauthCfg.AuthCodeURL(state,
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		)

		listener := NewOAuth2Listener()
		go listener.Listen(cmd.Context())

		fmt.Printf("Visit the following link in your terminal: %s\n", url)

		// TODO: Allow cancellation with CTRL+C
		// Cobra might already do this for us..
		code, err := listener.WaitForAuthorizationCode(cmd.Context(), state)
		if err != nil {
			return fmt.Errorf("failed to get authorization code: %w", err)
		}

		token, err := oauthCfg.Exchange(cmd.Context(), code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
		if err != nil {
			return fmt.Errorf("failed to exchange code for token: %w", err)
		}

		// TODO: Stash the token somewhere.
		// TODO: Grab the email from the id token
		b, _ := json.Marshal(token)
		fmt.Printf("token: %s", b)
		return nil
	},
}
