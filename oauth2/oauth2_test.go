package oauth2

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

type testExchanger struct {
	toks map[string]*oauth2.Token
}

func (te *testExchanger) AddToken(code string, token *oauth2.Token) {
	if te.toks == nil {
		te.toks = make(map[string]*oauth2.Token)
	}

	key := base64.RawStdEncoding.EncodeToString([]byte(code))
	te.toks[key] = token
}

func (te *testExchanger) Exchange(_ context.Context, code string, _ ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	// HACK: We cannot verify PKCE because params don't have a publicly accessible method for us to call.
	if te.toks == nil {
		te.toks = make(map[string]*oauth2.Token)
	}

	key := base64.RawStdEncoding.EncodeToString([]byte(code))
	tok, ok := te.toks[key]
	if !ok {
		return nil, errors.New("no token found for code")
	}
	return tok, nil
}

func Test_handler_YieldsCorrectlyFormattedState(t *testing.T) {
	var (
		ex            testExchanger
		handle        = &handler{Exchanger: &ex, jobs: make(chan job)}
		expectedToken = &oauth2.Token{
			AccessToken: "1234",
		}
		state    = "state goes here"
		code     = "code goes here"
		verifier = oauth2.GenerateVerifier()
		dl, _    = t.Deadline()
		ctx, _   = context.WithDeadline(context.Background(), dl)
		values   = url.Values{
			"state": []string{state},
			"code":  []string{code},
		}
		uri = url.URL{
			Scheme:   "http",
			Host:     "localhost",
			Path:     "/oauth2/callback",
			RawQuery: values.Encode(),
		}
		r = httptest.NewRequest("GET", uri.String(), nil)
		w = httptest.NewRecorder()
	)

	ex.AddToken(code, expectedToken)

	go handle.ServeHTTP(w, r)

	tok, err := handle.Wait(ctx, state, verifier)
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, tok)
}
