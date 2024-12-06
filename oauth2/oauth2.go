package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

var ErrNoSAMLAssertion = errors.New("no saml assertion")

// stateBufSize is the size of the buffer used to generate the state parameter.
// 43 is a magic number - It generates states that are not too short or long for Okta's validation.
const stateBufSize = 43

func DiscoverConfig(ctx context.Context, domain, clientID string) (*oauth2.Config, error) {
	provider, err := oidc.NewProvider(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("couldn't discover OIDC configuration for %s: %w", domain, err)
	}

	cfg := oauth2.Config{
		ClientID: clientID,
		Endpoint: provider.Endpoint(),
		Scopes:   []string{"openid", "profile", "okta.apps.sso"},
	}

	return &cfg, nil
}

// OAuth2CallbackState encapsulates all of the information from an oauth2 callback.
//
// To retrieve the Code from the struct, you must use the Verify(string) function.
type OAuth2CallbackState struct {
	code             string
	state            string
	errorMessage     string
	errorDescription string
}

// FromRequest parses the given http.Request and populates the OAuth2CallbackState with those values.
func (o *OAuth2CallbackState) FromRequest(r *http.Request) {
	o.errorMessage = r.FormValue("error")
	o.errorDescription = r.FormValue("error_description")
	o.state = r.FormValue("state")
	o.code = r.FormValue("code")
}

type OAuth2Error struct {
	Reason      string
	Description string
}

func (e OAuth2Error) Error() string {
	return fmt.Sprintf("oauth2 error: %s (%s)", e.Description, e.Reason)
}

func generateState() string {
	stateBuf := make([]byte, stateBufSize)
	rand.Read(stateBuf)
	return base64.URLEncoding.EncodeToString(stateBuf)
}

type Session struct {
	url      string
	state    string
	verifier string

	Token chan *oauth2.Token
	Error chan error
}

func (s Session) URL() string {
	return s.url
}

type AuthorizationCodeHandler struct {
	Config *oauth2.Config

	sessions map[string]Session
	mu       sync.Mutex
}

func (h *AuthorizationCodeHandler) NewSession() Session {
	state := generateState()
	verifier := oauth2.GenerateVerifier()
	url := h.Config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
	s := Session{verifier: verifier, state: state, url: url, Token: make(chan *oauth2.Token)}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[state] = s
	return s
}

func (h *AuthorizationCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var info OAuth2CallbackState
	info.FromRequest(r)

	// This lock is manually released in both branches, because if we defer() it, then it will get released
	// after the Exchange() call. Exchange() can take a decent amount of time since it involves a remote call,
	// and we don't want to hold the mutex lock for that long.
	h.mu.Lock()
	session, ok := h.sessions[info.state]
	if !ok {
		h.mu.Unlock()
		http.Error(w, "no session", http.StatusBadRequest)
		return
	}
	// Delete the session early so we can release the lock.
	delete(h.sessions, info.state)
	h.mu.Unlock()

	token, err := h.Config.Exchange(r.Context(), info.code, oauth2.VerifierOption(session.verifier))
	if err != nil {
		session.Error <- err
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Make sure to respond to the user right away. If we don't,
	// the server may be closed before a response can be sent.
	fmt.Fprintln(w, "You may close this window now.")
	session.Token <- token
	close(session.Token)
}

func DiscoverConfigAndExchangeTokenForAssertion(ctx context.Context, accessToken, idToken, oidcDomain, clientID, applicationID string) (*saml.Response, string, error) {
	oauthCfg, err := DiscoverConfig(ctx, oidcDomain, clientID)
	if err != nil {
		return nil, "", Error{Message: "could not discover oauth2  config", InnerError: err}
	}

	tok, err := exchangeAccessTokenForWebSSOToken(ctx, oauthCfg, accessToken, idToken, applicationID)
	if err != nil {
		return nil, "", Error{Message: "error exchanging token", InnerError: err}
	}

	assertionBytes, err := exchangeWebSSOTokenForSAMLAssertion(ctx, oidcDomain, tok)
	if err != nil {
		return nil, "", Error{Message: "failed to fetch SAML assertion", InnerError: err}
	}

	response, err := saml.ParseEncodedResponse(string(assertionBytes))
	if err != nil {
		return nil, "", Error{Message: "failed to parse SAML response", InnerError: err}
	}

	return response, string(assertionBytes), nil
}

type Error struct {
	InnerError error
	Message    string
}

func (o Error) Unwrap() error {
	return o.InnerError
}

func (o Error) Error() string {
	return o.Message
}
