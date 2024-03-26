package main

import (
	"context"
	"net"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// This test is going to be flaky because processes may open ports outside of our control.
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
