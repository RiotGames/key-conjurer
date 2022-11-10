package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func execute(cmd cobra.Command, args ...string) (string, error) {
	cmd.SetArgs(args)
	var buf bytes.Buffer
	cmd.SetOutput(&buf)
	err := cmd.Execute()
	return buf.String(), err
}

func TestVersionFlag(t *testing.T) {
	output, err := execute(*rootCmd, "--version")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	require.Equal(t, output, fmt.Sprintf("%s %s (Build Timestamp: %s - Client: %s)\n", appname, Version, buildTimestamp, ClientName))
}

func TestHelpLongInvalidArgs(t *testing.T) {
	_, err := execute(*rootCmd, "get", "-s")
	require.ErrorContains(t, err, "unknown shorthand flag: 's' in -s")
}

func TestInvalidCommand(t *testing.T) {
	_, err := execute(*rootCmd, "badcommand")
	require.ErrorContains(t, err, fmt.Sprintf("unknown command %q for %q", "badcommand", appname))
}
