package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	rootcerts "github.com/hashicorp/go-rootcerts"
	"golang.org/x/oauth2"
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
	codeVerifierBuf := make([]byte, 43)
	rand.Read(codeVerifierBuf)
	codeVerifier := base64.RawURLEncoding.EncodeToString(codeVerifierBuf)
	codeChallengeHash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(codeChallengeHash[:])
	return codeVerifier, codeChallenge, nil
}

func GenerateState() (string, error) {
	stateBuf := make([]byte, 43)
	rand.Read(stateBuf)
	return base64.URLEncoding.EncodeToString([]byte(stateBuf)), nil
}
