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
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc"
	rootcerts "github.com/hashicorp/go-rootcerts"
	"golang.org/x/net/html"
	"golang.org/x/oauth2"
)

var (
	ErrInvalidDomain = errors.New("invalid domain")
	// ErrTokenExchangeNotSupported indicates that token exchange is not supported for the given application.
	//
	// This most commonly occurs when attempting to use token exchange on non-AWS applications with Okta.
	// Okta currently (2023-09-25) only supports the web sso grant type for AWS applications.
	ErrTokenExchangeNotSupported = errors.New("token exchange not supported")
)

// ErrOktaErrorResponse is returned when Okta returns a non-200 response that is not covered by other well-defined errors.
type ErrOktaErrorResponse struct {
	StatusCode int
	Response   *http.Response
}

func (e ErrOktaErrorResponse) Error() string {
	return fmt.Sprintf("bad response code: %d", e.StatusCode)
}

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
	uri, err := url.Parse(domain)
	if domain == "" || err != nil {
		return nil, ErrInvalidDomain
	}

	provider, err := oidc.NewProvider(ctx, uri.String())
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
	Addr       string
	errCh      chan error
	callbackCh chan OAuth2CallbackInfo
}

func NewOAuth2Listener() OAuth2Listener {
	return OAuth2Listener{
		// 5RIOT on a phone pad
		Addr:       ":57468",
		errCh:      make(chan error),
		callbackCh: make(chan OAuth2CallbackInfo),
	}
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

func (o OAuth2Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	info, err := ParseCallbackRequest(r)
	if err == nil {
		// The only errors that might occur would be incorreclty formatted requests, which we will silently drop.
		o.callbackCh <- info
	}

	// This is displayed to the end user in their browser.
	fmt.Fprintln(w, "You may close this window now.")
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

		if strings.Compare(info.State, state) != 0 {
			return "", OAuth2Error{Reason: "invalid_state", Description: "state mismatch"}
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

func GenerateCodeVerifierAndChallenge() (string, string, error) {
	codeVerifierBuf := make([]byte, stateBufSize)
	rand.Read(codeVerifierBuf)
	codeVerifier := base64.RawURLEncoding.EncodeToString(codeVerifierBuf)
	codeChallengeHash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(codeChallengeHash[:])
	return codeVerifier, codeChallenge, nil
}

func GenerateState() (string, error) {
	stateBuf := make([]byte, stateBufSize)
	rand.Read(stateBuf)
	return base64.URLEncoding.EncodeToString([]byte(stateBuf)), nil
}

func RedirectionFlow(ctx context.Context, oauthCfg *oauth2.Config, state, codeChallenge, codeVerifier string, outputMode LoginOutputMode) (*oauth2.Token, error) {
	listener := NewOAuth2Listener()
	go listener.Listen(ctx)
	oauthCfg.RedirectURL = "http://localhost:57468"
	url := oauthCfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
	)

	if err := outputMode.PrintURL(url); err != nil {
		// This is unlikely to ever happen
		return nil, fmt.Errorf("failed to display link: %w", err)
	}

	code, err := listener.WaitForAuthorizationCode(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization code: %w", err)
	}

	return oauthCfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
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

	switch resp.StatusCode {
	case http.StatusOK:
		var tok oauth2.Token
		return &tok, json.NewDecoder(resp.Body).Decode(&tok)
	case http.StatusBadRequest:
		// Unsupported application - This application probably hasn't been configured to support token exchange.
		// In other words, it's probably not an AWS application.
		return nil, ErrTokenExchangeNotSupported
	default:
		return nil, ErrOktaErrorResponse{resp.StatusCode, resp}
	}
}

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
		return nil, errors.New("could not find form")
	}

	saml, ok := form.Inputs["SAMLResponse"]
	if !ok {
		return nil, errors.New("no SAML assertion")
	}

	return []byte(saml), nil
}
