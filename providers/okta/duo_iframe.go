package okta

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/riotgames/key-conjurer/providers/duo"
	"github.com/tidwall/gjson"
)

type DuoIframe struct {
	Host        string
	SignedToken SignedToken
	CallbackURL string
	StateHandle string
	Method      string
	StateToken  StateToken
}

func (f DuoIframe) Upgrade(ctx context.Context, client *http.Client) (StateToken, error) {
	duo := duo.New()
	tok, err := duo.SendPush(f.SignedToken.Tx, string(f.StateToken), f.CallbackURL, f.Host)
	if err != nil {
		return f.StateToken, err
	}

	// Post the state token back to the callbackURL.
	vals := map[string]any{
		"credentials": map[string]string{
			"signatureData": f.SignedToken.ConcatenateAuthSignature(tok),
		},
		"stateHandle": f.StateHandle,
	}

	// The following request will redirect us back to Okta when we receive a successful response.
	reqBuf, _ := json.Marshal(vals)
	req, _ := http.NewRequestWithContext(ctx, f.Method, f.CallbackURL, bytes.NewReader(reqBuf))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return f.StateToken, err
	}

	defer resp.Body.Close()
	bodyBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		return f.StateToken, err
	}

	// HACK: This HTTP response will always HTTP 500.
	//
	// The SP-initiated flow using an application link always yields a HTTP 500 on the last redirect.
	// This has been raised with Okta and it's not clear why this occurs.
	//
	// Because the SP-initiated flow works so well in DuoFrameless, and the SP-initiated flow with DuoIframe will only be around for a short while, we're okay with this hack.
	//
	// The server will always return HTTP 500 here, but it will still upgrade the users session and set the cookie. Without this step, the session won't be upgraded correctly and a user won't be able to auth.
	req, _ = http.NewRequestWithContext(ctx, "GET", gjson.GetBytes(bodyBuf, "success.href").Str, nil)
	_, err = client.Do(req)
	return f.StateToken, err
}
