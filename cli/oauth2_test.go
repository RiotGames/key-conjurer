package main

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// This test ensures that the success flow of receiving an authorization code from a HTTP handler works correctly.
func Test_OAuth2Listener_WaitForAuthorizationCodeWorksCorrectly(t *testing.T) {
	// TODO: Initializing private fields means we should probably refactor OAuth2Listener's constructor to not require net.Listener,
	// but instead to have net.Listener be a thing in Listen();
	//
	// Users should be able to use ServeHTTP.
	listener := OAuth2Listener{
		callbackCh: make(chan OAuth2CallbackInfo),
	}

	expectedState := "state goes here"
	expectedCode := "code goes here"

	go func() {
		w := httptest.NewRecorder()
		values := url.Values{
			"code":  []string{expectedCode},
			"state": []string{expectedState},
		}

		uri := url.URL{
			Scheme:   "http",
			Host:     "localhost",
			Path:     "/oauth2/callback",
			RawQuery: values.Encode(),
		}

		req := httptest.NewRequest("GET", uri.String(), nil)
		listener.ServeHTTP(w, req)
	}()

	deadline, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	code, err := listener.WaitForAuthorizationCode(deadline, expectedState)
	assert.NoError(t, err)
	cancel()

	assert.Equal(t, expectedCode, code)
	assert.NoError(t, listener.Close())
}

func Test_OAuth2Listener_ZeroValueNeverPanics(t *testing.T) {
	var listener OAuth2Listener
	deadline, _ := context.WithTimeout(context.Background(), 500*time.Millisecond)
	_, err := listener.WaitForAuthorizationCode(deadline, "")
	assert.ErrorIs(t, context.DeadlineExceeded, err)
	assert.NoError(t, listener.Close())
}
