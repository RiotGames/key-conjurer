package oauth2

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/oauth2"
)

var cfg = &oauth2.Config{
	ClientID: "client-id",
	Endpoint: oauth2.Endpoint{
		TokenURL: "http://localhost/oauth2/token",
	},
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

var client = http.Client{
	Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/oauth2/token" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       nil,
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       nil,
		}, nil
	}),
}

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

// Test_AuthorizationCodeHandler_IsReentrant prevents an issue where AuthorizationCodeHandler would send a request to a closed channel
func Test_AuthorizationCodeHandler_IsReentrant(t *testing.T) {
	handler := NewAuthorizationCodeHandler(cfg)

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
	// If we reach here with no panics, it should pass
}
