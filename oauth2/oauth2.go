package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
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

// Verify safely compares the given state with the state from the OAuth2 callback.
//
// If they match, the code is returned, with a nil value. Otherwise, an empty string and an error is returned.
func (o OAuth2CallbackState) Verify(expectedState string) (string, error) {
	if o.errorMessage != "" {
		return "", OAuth2Error{Reason: o.errorMessage, Description: o.errorDescription}
	}

	if strings.Compare(o.state, expectedState) != 0 {
		return "", OAuth2Error{Reason: "invalid_state", Description: "state mismatch"}
	}

	return o.code, nil
}

// OAuth2CallbackHandler returns a http.Handler, channel and function triple.
//
// The http handler will accept exactly one request, which it will assume is an OAuth2 callback, parse it into an OAuth2CallbackState and then provide it to the given channel. Subsequent requests will be silently ignored.
//
// The function may be called to ensure that the channel is closed. The channel is closed when a request is received. In general, it is a good idea to ensure this function is called in a defer() block.
func OAuth2CallbackHandler() (http.Handler, <-chan OAuth2CallbackState, func()) {
	// TODO: It is possible for the caller to close a panic() if they execute the function in the triplet while the handler has not yet received a request.
	// That caller is us, so I don't care that much, but that probably indicates that this design is smelly.
	//
	// We should look at the Go SDK to see how they handle similar cases - channels that are not bound by a timer, or similar.

	ch := make(chan OAuth2CallbackState, 1)
	var reqHandle, closeHandle sync.Once
	closeFn := func() {
		closeHandle.Do(func() {
			close(ch)
		})
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		// This can sometimes be called multiple times, depending on the browser.
		// We will simply ignore any other requests and only serve the first.
		reqHandle.Do(func() {
			var state OAuth2CallbackState
			state.FromRequest(r)
			ch <- state
			closeFn()
		})

		// We still want to provide feedback to the end-user.
		fmt.Fprintln(w, "You may close this window now.")
	}

	return http.HandlerFunc(fn), ch, closeFn
}

type OAuth2Error struct {
	Reason      string
	Description string
}

func (e OAuth2Error) Error() string {
	return fmt.Sprintf("oauth2 error: %s (%s)", e.Description, e.Reason)
}

func GenerateState() string {
	stateBuf := make([]byte, stateBufSize)
	rand.Read(stateBuf)
	return base64.URLEncoding.EncodeToString(stateBuf)
}

type RedirectionFlowHandler struct {
	Config       *oauth2.Config
	OnDisplayURL func(url string) error
}

func (r RedirectionFlowHandler) HandlePendingSession(ctx context.Context, listener net.Listener, state string) (*oauth2.Token, error) {
	if r.OnDisplayURL == nil {
		panic("OnDisplayURL must be set")
	}

	verifier := oauth2.GenerateVerifier()
	url := r.Config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
	callbackHandler, ch, cancel := OAuth2CallbackHandler()
	// TODO: This error probably should not be ignored if it is not http.ErrServerClosed
	go http.Serve(listener, callbackHandler)
	defer cancel()

	if err := r.OnDisplayURL(url); err != nil {
		// This is unlikely to ever happen
		return nil, fmt.Errorf("failed to display link: %w", err)
	}

	select {
	case info := <-ch:
		code, err := info.Verify(state)
		if err != nil {
			return nil, fmt.Errorf("failed to get authorization code: %w", err)
		}
		return r.Config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	case <-ctx.Done():
		return nil, ctx.Err()
	}
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
