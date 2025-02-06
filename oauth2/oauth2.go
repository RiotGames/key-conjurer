package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"

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
		Scopes:   []string{"openid", "profile", "okta.apps.read", "okta.apps.sso"},
	}

	return &cfg, nil
}

func generateState() string {
	stateBuf := make([]byte, stateBufSize)
	rand.Read(stateBuf)
	return base64.URLEncoding.EncodeToString(stateBuf)
}

type RedirectionFlowHandler struct {
	Config       *oauth2.Config
	OnDisplayURL func(url string) error
}

func (r RedirectionFlowHandler) HandlePendingSession(ctx context.Context, listener net.Listener) (*oauth2.Token, error) {
	if r.OnDisplayURL == nil {
		panic("OnDisplayURL must be set")
	}

	state := generateState()
	verifier := oauth2.GenerateVerifier()
	url := r.Config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
	handler := &handler{jobs: make(chan job), Exchanger: r.Config}
	// TODO: This error probably should not be ignored if it is not http.ErrServerClosed
	go http.Serve(listener, handler)
	defer handler.Close()

	if err := r.OnDisplayURL(url); err != nil {
		// This is unlikely to ever happen
		return nil, fmt.Errorf("failed to display link: %w", err)
	}

	return handler.Wait(ctx, state, verifier)
}

func DiscoverConfigAndExchangeTokenForAssertion(ctx context.Context, accessToken, idToken, oidcDomain, clientID, applicationID string) (*saml.Response, string, error) {
	oauthCfg, err := DiscoverConfig(ctx, oidcDomain, clientID)
	if err != nil {
		return nil, "", fmt.Errorf("discover oauth2 config: %w", err)
	}

	tok, err := exchangeAccessTokenForWebSSOToken(ctx, oauthCfg, accessToken, idToken, applicationID)
	if err != nil {
		return nil, "", fmt.Errorf("get websso token: %w", err)
	}

	assertionBytes, err := exchangeWebSSOTokenForSAMLAssertion(ctx, oidcDomain, tok)
	if err != nil {
		return nil, "", fmt.Errorf("get saml assertion: %w", err)
	}

	response, err := saml.ParseEncodedResponse(string(assertionBytes))
	if err != nil {
		return nil, "", fmt.Errorf("parse saml response: %w", err)
	}

	return response, string(assertionBytes), nil
}
