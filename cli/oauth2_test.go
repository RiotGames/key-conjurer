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

func sendRequestToListener(listener *OAuth2Listener, values url.Values) {
	uri := url.URL{
		Scheme:   "http",
		Host:     "localhost",
		Path:     "/oauth2/callback",
		RawQuery: values.Encode(),
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", uri.String(), nil)
	listener.ServeHTTP(w, req)
}

// This test ensures that the success flow of receiving an authorization code from a HTTP handler works correctly.
func Test_OAuth2Listener_WaitForAuthorizationCodeWorksCorrectly(t *testing.T) {
	listener := NewOAuth2Listener()
	expectedState := "state goes here"
	expectedCode := "code goes here"

	go sendRequestToListener(&listener, url.Values{
		"code":  []string{expectedCode},
		"state": []string{expectedState},
	})

	deadline, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	code, err := listener.WaitForAuthorizationCode(deadline, expectedState)
	assert.NoError(t, err)
	cancel()

	assert.Equal(t, expectedCode, code)
}

func Test_OAuth2Listener_ZeroValueNeverPanics(t *testing.T) {
	var listener OAuth2Listener
	deadline, _ := context.WithTimeout(context.Background(), 500*time.Millisecond)
	_, err := listener.WaitForAuthorizationCode(deadline, "")
	// This will timeout because the OAuth2Listener will forever listen on a nil channel
	assert.ErrorIs(t, context.DeadlineExceeded, err)
}

// Test_OAuth2Listener_MultipleRequestsDoesNotCausePanic prevents an issue where OAuth2Listener would send a request to a closed channel
func Test_OAuth2Listener_MultipleRequestsDoesNotCausePanic(t *testing.T) {
	listener := NewOAuth2Listener()
	expectedState := "state goes here"
	expectedCode := "code goes here"

	go sendRequestToListener(&listener, url.Values{
		"code":  []string{expectedCode},
		"state": []string{expectedState},
	})

	go sendRequestToListener(&listener, url.Values{
		"code":  []string{"not the expected code and should be discarded"},
		"state": []string{"not the expected state and should be discarded"},
	})

	deadlineCtx, _ := context.WithTimeout(context.Background(), 500*time.Millisecond)
	code, err := listener.WaitForAuthorizationCode(deadlineCtx, expectedState)
	assert.NoError(t, err)
	assert.Equal(t, expectedCode, code)
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
