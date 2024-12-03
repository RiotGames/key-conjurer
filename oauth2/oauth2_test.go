package oauth2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/oauth2"
)

func sendOAuth2CallbackRequest(handler http.Handler, values url.Values) {
	uri := url.URL{
		Scheme:   "http",
		Host:     "localhost",
		Path:     "/oauth2/callback",
		RawQuery: values.Encode(),
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", uri.String(), nil)
	handler.ServeHTTP(w, req)
}

type testCodeExchanger struct{}

func (t *testCodeExchanger) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return nil, nil
}

// Test_OAuth2Listener_MultipleRequestsDoesNotCausePanic prevents an issue where OAuth2Listener would send a request to a closed channel
func Test_OAuth2Listener_MultipleRequestsDoesNotCausePanic(t *testing.T) {
	ch := make(chan Callback, 2)
	defer close(ch)

	handler := OAuth2CallbackHandler(&testCodeExchanger{}, "state", "verifier", ch)

	go sendOAuth2CallbackRequest(handler, url.Values{
		// We send empty values because we don't care about processing in this test
		"code":  []string{""},
		"state": []string{""},
	})

	// We send this request synchronously to ensure that any panics are caught during the test.
	sendOAuth2CallbackRequest(handler, url.Values{
		"code":  []string{"not the expected code and should be discarded"},
		"state": []string{"not the expected state and should be discarded"},
	})
}
