package command

import (
	"context"
	"net"
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
	assert.ErrorIs(t, ErrNoPortsAvailable, err)
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
	assert.ErrorIs(t, err, ErrNoPortsAvailable)
}
