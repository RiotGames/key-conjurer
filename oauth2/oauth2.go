package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

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

type callbackState struct {
	code             string
	state            string
	errorMessage     string
	errorDescription string
}

func parseOAuth2CallbackState(r *http.Request, info *callbackState) error {
	err := r.ParseForm()
	info.errorMessage = r.FormValue("error")
	info.errorDescription = r.FormValue("error_description")
	info.state = r.FormValue("state")
	info.code = r.FormValue("code")
	return err
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

// NewAuthorizationCodeHandler creates a new AuthorizationCodeHandler.
func NewAuthorizationCodeHandler(config *oauth2.Config) *AuthorizationCodeHandler {
	return &AuthorizationCodeHandler{
		config:   config,
		sessions: make(map[string]Session),
	}
}

// AuthorizationCodeHandler is an http.Handler that handles the OAuth2 authorization code flow.
//
// It is intended to be used by CLIs that need to authenticate with an OAuth2 provider.
//
// Sessions can be created using NewSession, and those sessions can be used to retrieve the OAuth2 token.
type AuthorizationCodeHandler struct {
	config   *oauth2.Config
	sessions map[string]Session
	mu       sync.Mutex
}

func (a *AuthorizationCodeHandler) NewSession() Session {
	state := generateState()
	verifier := oauth2.GenerateVerifier()
	url := a.config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
	// A channel capacity of 1 is used to prevent requests from blocking if they are not actively being awaited on.
	s := Session{verifier: verifier, state: state, url: url, Token: make(chan *oauth2.Token, 1)}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sessions[state] = s
	return s
}

func (a *AuthorizationCodeHandler) removeSessionIfExists(state string) (Session, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s, ok := a.sessions[state]
	if ok {
		delete(a.sessions, state)
	}
	return s, ok
}

func (a *AuthorizationCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var st callbackState
	if err := parseOAuth2CallbackState(r, &st); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session, ok := a.removeSessionIfExists(st.state)
	if !ok {
		http.Error(w, "no session", http.StatusBadRequest)
		return
	}

	token, err := a.config.Exchange(r.Context(), st.code, oauth2.VerifierOption(session.verifier))
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
