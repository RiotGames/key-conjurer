package command

import (
	"context"
	"net"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_findFirstFreePort_WorksCorrectly(t *testing.T) {
	ports := []string{"58080", "58081", "58082", "58083"}
	socket, err := net.Listen("tcp4", net.JoinHostPort("127.0.0.1", ports[0]))
	t.Cleanup(func() {
		socket.Close()
	})
	require.NoError(t, err, "Could not open socket on port: %s", ports[0])

	openedSocket, err := findFirstFreePort(context.Background(), "127.0.0.1", ports)
	assert.NoError(t, err)
	_, port, err := net.SplitHostPort(openedSocket.Addr().String())
	assert.NoError(t, err)
	// There is no guarantee on which port FindFirstFreePort will choose, but it must pick one from the given list.
	assert.Contains(t, ports, port)
	openedSocket.Close()
}

func Test_findFirstFreePort_RejectsIfNoPortsAvailable(t *testing.T) {
	var ports []string
	_, err := findFirstFreePort(context.Background(), "127.0.0.1", ports)
	assert.ErrorIs(t, errNoPortsAvailable, err)
}

func Test_findFirstFreePort_RejectsIfAllProvidedPortsExhausted(t *testing.T) {
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

	_, err := findFirstFreePort(context.Background(), "127.0.0.1", activePorts)
	assert.ErrorIs(t, err, errNoPortsAvailable)
}

func Test_osc8Hyperlink_ProducesExactOSC8Sequence(t *testing.T) {
	url := "https://example.okta.com/oauth2/v1/authorize?client_id=abc&scope=openid"
	got := osc8Hyperlink(url, url)
	want := "\x1b]8;;" + url + "\x1b\\" + url + "\x1b]8;;\x1b\\"
	assert.Equal(t, want, got)
	assert.True(t, strings.HasPrefix(got, "\x1b]8;;"))
	assert.True(t, strings.HasSuffix(got, "\x1b]8;;\x1b\\"))
}

func Test_osc8Hyperlink_DegradesToPlainText(t *testing.T) {
	url := "https://example.okta.com/authorize?x=1&y=2"
	got := osc8Hyperlink(url, url)
	// An OSC-8-aware terminal that chooses not to render the link still
	// consumes the OSC control sequence up to its ST terminator, leaving
	// only the link text visible.
	osc8 := regexp.MustCompile("\x1b\\]8;;[^\x1b]*\x1b\\\\")
	visible := osc8.ReplaceAllString(got, "")
	assert.Equal(t, url, visible)
}

func Test_osc8Hyperlink_RawOutputContainsFullURL(t *testing.T) {
	url := "https://example.okta.com/authorize?x=1&y=2"
	got := osc8Hyperlink(url, url)
	// Even if a terminal renders escapes literally, the URL text between
	// the OSC wrappers is emitted verbatim and contiguous.
	assert.Contains(t, got, "\x1b\\"+url+"\x1b]8;;")
}

func Test_hyperlink_ReturnsPlainTextWhenNotATerminal(t *testing.T) {
	// Under `go test`, stdout is not a terminal, so hyperlink must return the
	// plain text unchanged — no OSC 8 escape bytes that would corrupt piped or
	// redirected output.
	url := "https://example.okta.com/authorize?x=1&y=2"
	got := hyperlink(url, url)
	assert.Equal(t, url, got)
	assert.NotContains(t, got, "\x1b")
}
