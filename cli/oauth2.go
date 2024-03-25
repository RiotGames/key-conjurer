package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/coreos/go-oidc"
	rootcerts "github.com/hashicorp/go-rootcerts"
	"golang.org/x/net/html"
	"golang.org/x/oauth2"
)

var ErrNoSAMLAssertion = errors.New("no saml assertion")

// stateBufSize is the size of the buffer used to generate the state parameter.
// 43 is a magic number - It generates states that are not too short or long for Okta's validation.
const stateBufSize = 43

func NewHTTPClient() *http.Client {
	// Some Darwin systems require certs to be loaded from the system certificate store or attempts to verify SSL certs on internal websites may fail.
	tr := http.DefaultTransport
	if certs, err := rootcerts.LoadSystemCAs(); err == nil {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certs,
			},
		}
	}

	return &http.Client{Transport: LogRoundTripper{tr}}
}

func DiscoverOAuth2Config(ctx context.Context, domain, clientID string) (*oauth2.Config, error) {
	provider, err := oidc.NewProvider(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("couldn't discover OIDC configuration for %s: %w", domain, err)
	}

	cfg := oauth2.Config{
		ClientID: clientID,
		Endpoint: provider.Endpoint(),
		Scopes:   []string{"openid", "profile", "okta.apps.read", "okta.apps.sso"},
	}

	return &cfg, nil
}

type OAuth2CallbackInfo struct {
	Code             string
	State            string
	Error            string
	ErrorDescription string
}

type OAuth2Listener struct {
	Socket     net.Listener
	callbackCh chan OAuth2CallbackInfo
}

func ParseCallbackRequest(r *http.Request) (OAuth2CallbackInfo, error) {
	info := OAuth2CallbackInfo{
		Error:            r.FormValue("error"),
		ErrorDescription: r.FormValue("error_description"),
		State:            r.FormValue("state"),
		Code:             r.FormValue("code"),
	}

	return info, nil
}

func NewOAuth2Listener(socket net.Listener) OAuth2Listener {
	return OAuth2Listener{
		Socket:     socket,
		callbackCh: make(chan OAuth2CallbackInfo),
	}
}

func (o OAuth2Listener) Close() error {
	if o.callbackCh != nil {
		close(o.callbackCh)
	}
	if o.Socket != nil {
		return o.Socket.Close()
	}
	return nil
}

func (o OAuth2Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	info, err := ParseCallbackRequest(r)
	if err == nil {
		// The only errors that might occur would be incorreclty formatted requests, which we will silently drop.
		o.callbackCh <- info
	}

	// This is displayed to the end user in their browser.
	fmt.Fprintln(w, "You may close this window now.")
}

func (o OAuth2Listener) Listen() error {
	err := http.Serve(o.Socket, o)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (o OAuth2Listener) WaitForAuthorizationCode(ctx context.Context, state string) (string, error) {
	select {
	case info := <-o.callbackCh:
		if info.Error != "" {
			return "", OAuth2Error{Reason: info.Error, Description: info.ErrorDescription}
		}

		if strings.Compare(info.State, state) != 0 {
			return "", OAuth2Error{Reason: "invalid_state", Description: "state mismatch"}
		}

		return info.Code, nil
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

func GeneratePkceChallenge() PkceChallenge {
	codeVerifierBuf := make([]byte, stateBufSize)
	rand.Read(codeVerifierBuf)
	codeVerifier := base64.RawURLEncoding.EncodeToString(codeVerifierBuf)
	codeChallengeHash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(codeChallengeHash[:])
	return PkceChallenge{Verifier: codeVerifier, Challenge: codeChallenge}
}

func GenerateState() string {
	stateBuf := make([]byte, stateBufSize)
	rand.Read(stateBuf)
	return base64.URLEncoding.EncodeToString(stateBuf)
}

type PkceChallenge struct {
	Challenge string
	Verifier  string
}

func printURLToConsole(url string) error {
	fmt.Fprintln(os.Stdout, url)
	return nil
}

type RedirectionFlowHandler struct {
	Config       *oauth2.Config
	OnDisplayURL func(url string) error

	// Listen is a function that can be provided to override how the redirection flow handler opens a network socket.
	// If this is not specified, the handler will attempt to create a connection that listens to 0.0.0.0:57468 on IPv4.
	Listen func() (net.Listener, error)
}

func (r RedirectionFlowHandler) HandlePendingSession(ctx context.Context, challenge PkceChallenge, state string) (*oauth2.Token, error) {
	if r.OnDisplayURL == nil {
		r.OnDisplayURL = printURLToConsole
	}

	if r.Listen == nil {
		r.Listen = func() (net.Listener, error) {
			var lc net.ListenConfig
			sock, err := lc.Listen(ctx, "tcp4", net.JoinHostPort("0.0.0.0", "57468"))
			return sock, err
		}
	}

	sock, err := r.Listen()
	if err != nil {
		return nil, err
	}

	r.Config.RedirectURL = "http://localhost:57468"
	url := r.Config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", challenge.Challenge),
	)

	listener := NewOAuth2Listener(sock)
	defer listener.Close()
	// This error can be ignored.
	go listener.Listen()

	if err := r.OnDisplayURL(url); err != nil {
		// This is unlikely to ever happen
		return nil, fmt.Errorf("failed to display link: %w", err)
	}

	code, err := listener.WaitForAuthorizationCode(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization code: %w", err)
	}

	return r.Config.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", challenge.Verifier))
}

func ExchangeAccessTokenForWebSSOToken(ctx context.Context, client *http.Client, oauthCfg *oauth2.Config, token *TokenSet, applicationID string) (*oauth2.Token, error) {
	if client == nil {
		client = http.DefaultClient
	}
	// https://datatracker.ietf.org/doc/html/rfc8693
	data := url.Values{
		"client_id":          {oauthCfg.ClientID},
		"actor_token":        {token.AccessToken},
		"actor_token_type":   {"urn:ietf:params:oauth:token-type:access_token"},
		"subject_token":      {token.IDToken},
		"subject_token_type": {"urn:ietf:params:oauth:token-type:id_token"},
		"grant_type":         {"urn:ietf:params:oauth:grant-type:token-exchange"},
		// https://www.linkedin.com/pulse/oktas-aws-cli-app-mysterious-case-powerful-okta-apis-chaim-sanders/
		"requested_token_type": {"urn:okta:oauth:token-type:web_sso_token"},
		"audience":             {fmt.Sprintf("urn:okta:apps:%s", applicationID)},
	}
	body := strings.NewReader(data.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, oauthCfg.Endpoint.TokenURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// TODO: The response can indicate a failure, we should check that for this function
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var tok oauth2.Token
	return &tok, json.NewDecoder(resp.Body).Decode(&tok)
}

// TODO: This is actually an Okta-specific API
func ExchangeWebSSOTokenForSAMLAssertion(ctx context.Context, client *http.Client, issuer string, token *oauth2.Token) ([]byte, error) {
	if client == nil {
		client = http.DefaultClient
	}

	data := url.Values{"token": {token.AccessToken}}
	uri := fmt.Sprintf("%s/login/token/sso?%s", issuer, data.Encode())
	req, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	req.Header.Add("Accept", "text/html")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusInternalServerError {
		return nil, errors.New("internal okta error occurred")
	}

	doc, _ := html.Parse(resp.Body)
	form, ok := FindFirstForm(doc)
	if !ok {
		return nil, ErrNoSAMLAssertion
	}

	saml, ok := form.Inputs["SAMLResponse"]
	if !ok {
		return nil, ErrNoSAMLAssertion
	}

	return []byte(saml), nil
}

func DiscoverConfigAndExchangeTokenForAssertion(ctx context.Context, client *http.Client, toks *TokenSet, oidcDomain, clientID, applicationID string) (*saml.Response, string, error) {
	oauthCfg, err := DiscoverOAuth2Config(ctx, oidcDomain, clientID)
	if err != nil {
		return nil, "", OktaError{Message: "could not discover oauth2  config", InnerError: err}
	}

	tok, err := ExchangeAccessTokenForWebSSOToken(ctx, client, oauthCfg, toks, applicationID)
	if err != nil {
		return nil, "", OktaError{Message: "error exchanging token", InnerError: err}
	}

	assertionBytes, err := ExchangeWebSSOTokenForSAMLAssertion(ctx, client, oidcDomain, tok)
	if err != nil {
		return nil, "", OktaError{Message: "failed to fetch SAML assertion", InnerError: err}
	}

	response, err := ParseBase64EncodedSAMLResponse(string(assertionBytes))
	if err != nil {
		return nil, "", OktaError{Message: "failed to parse SAML response", InnerError: err}
	}

	return response, string(assertionBytes), nil
}
