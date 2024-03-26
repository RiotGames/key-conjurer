package main

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func Test_OAuth2CallbackHandler_YieldsCorrectlyFormattedState(t *testing.T) {
	handler, ch, cancel := OAuth2CallbackHandler()
	t.Cleanup(func() {
		cancel()
	})

	expectedState := "state goes here"
	expectedCode := "code goes here"

	go sendOAuth2CallbackRequest(handler, url.Values{
		"code":  []string{expectedCode},
		"state": []string{expectedState},
	})

	callbackState := <-ch
	code, err := callbackState.Verify(expectedState)
	assert.NoError(t, err)
	assert.Equal(t, expectedCode, code)
}

func Test_OAuth2CallbackState_VerifyWorksCorrectly(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		expectedState := "state goes here"
		expectedCode := "code goes here"
		callbackState := OAuth2CallbackState{
			code:  expectedCode,
			state: expectedState,
		}
		code, err := callbackState.Verify(expectedState)
		assert.NoError(t, err)
		assert.Equal(t, expectedCode, code)
	})

	t.Run("unhappy path", func(t *testing.T) {
		expectedState := "state goes here"
		expectedCode := "code goes here"
		callbackState := OAuth2CallbackState{
			code:  expectedCode,
			state: expectedState,
		}
		_, err := callbackState.Verify("mismatching state")
		var oauthErr OAuth2Error
		assert.ErrorAs(t, err, &oauthErr)
		assert.Equal(t, "invalid_state", oauthErr.Reason)
	})
}

// Test_OAuth2Listener_MultipleRequestsDoesNotCausePanic prevents an issue where OAuth2Listener would send a request to a closed channel
func Test_OAuth2Listener_MultipleRequestsDoesNotCausePanic(t *testing.T) {
	handler, ch, cancel := OAuth2CallbackHandler()
	t.Cleanup(func() {
		cancel()
	})

	go sendOAuth2CallbackRequest(handler, url.Values{
		// We send empty values because we don't care about processing in this test
		"code":  []string{""},
		"state": []string{""},
	})

	// We drain the channel of the first request so the handler completes.
	// Without this step, we would get 'stuck' in the sync.Once().
	<-ch

	// We send this request synchronously to ensure that any panics are caught during the test.
	sendOAuth2CallbackRequest(handler, url.Values{
		"code":  []string{"not the expected code and should be discarded"},
		"state": []string{"not the expected state and should be discarded"},
	})
}

// Test_ListenAnyPort_WorksCorrectly is going to be flaky because processes may open ports outside of our control.
func Test_ListenAnyPort_WorksCorrectly(t *testing.T) {
	ports := []string{"58080", "58081", "58082", "58083"}
	socket, err := net.Listen("tcp4", net.JoinHostPort("127.0.0.1", ports[0]))
	t.Cleanup(func() {
		socket.Close()
	})
	require.NoError(t, err, "Could not open socket on port: %s", ports[0])

	listenFunc := ListenAnyPort("127.0.0.1", ports)
	openedSocket, err := listenFunc(context.Background())

	assert.NoError(t, err)
	_, port, err := net.SplitHostPort(openedSocket.Addr().String())
	assert.NoError(t, err)
	// There is no guarantee on which port FindFirstFreePort will choose, but it must pick one from the given list.
	assert.Contains(t, ports, port)
	openedSocket.Close()
}

func Test_ListenAnyPort_RejectsIfNoPortsAvailable(t *testing.T) {
	var ports []string
	listenFunc := ListenAnyPort("127.0.0.1", ports)
	_, err := listenFunc(context.Background())
	assert.ErrorIs(t, ErrNoPortsAvailable, err)
}

func Test_ListenAnyPort_RejectsIfAllProvidedPortsExhausted(t *testing.T) {
	ports := []string{"58080", "58081", "58082", "58083"}
	var sockets []net.Listener
	var activePorts []string
	// This exhausts all sockets in 'ports' and dumps them into 'activePorts'.
	for _, port := range ports {
		socket, err := net.Listen("tcp4", net.JoinHostPort("127.0.0.1", port))
		if err == nil {
			sockets = append(sockets, socket)
			activePorts = append(activePorts, port)
		}
	}

	require.NotEmpty(t, activePorts, "could not open any sockets")

	t.Cleanup(func() {
		for _, socket := range sockets {
			socket.Close()
		}
	})

	_, err := ListenAnyPort("127.0.0.1", activePorts)(context.Background())
	assert.ErrorIs(t, err, ErrNoPortsAvailable)
}
