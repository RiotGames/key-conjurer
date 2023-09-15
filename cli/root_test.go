package main

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Required to reset Cobra state between Test runs.
// If you add new tests that change state, you may need to
// add code here to reset the side effects of other tests
func resetCobra(cmd *cobra.Command) {
	cmd.Flags().Set("help", "false")
	cmd.Flags().Set("version", "false")
}

func execute(cmd *cobra.Command, args ...string) (string, error) {
	var buf bytes.Buffer
	cmd.SetArgs(args)
	cmd.SetOutput(&buf)
	err := cmd.Execute()
	return buf.String(), err
}

func TestVersionFlag(t *testing.T) {
	t.Cleanup(func() {
		resetCobra(rootCmd)
	})

	output, err := execute(rootCmd, "--version")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := fmt.Sprintf("keyconjurer-%s-%s TBD (date not set time not set zone not set)\n", runtime.GOOS, runtime.GOARCH)
	assert.Equal(t, output, expected)
}

func TestVersionShortFlag(t *testing.T) {
	t.Cleanup(func() {
		resetCobra(rootCmd)
	})

	output, err := execute(rootCmd, "-v")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := fmt.Sprintf("keyconjurer-%s-%s TBD (date not set time not set zone not set)\n", runtime.GOOS, runtime.GOARCH)
	assert.Equal(t, output, expected)
}

func TestHelpLongInvalidArgs(t *testing.T) {
	t.Cleanup(func() {
		resetCobra(rootCmd)
	})
	output, err := execute(rootCmd, "get", "-s")
	if err != nil {
		if err.Error() != "unknown shorthand flag: 's' in -s" {
			t.Errorf("Unexpected error: %v", err)
		}
	} else {
		t.Errorf("Unexpected non-error, output=: %v", output)
	}
}

func TestInvalidCommand(t *testing.T) {
	t.Cleanup(func() {
		resetCobra(rootCmd)
	})

	output, err := execute(rootCmd, "badcommand")
	if err != nil {
		if err.Error() != "unknown command \"badcommand\" for \"keyconjurer\"" {
			t.Errorf("Unexpected error: %v", err)
		}
	} else {
		t.Errorf("Unexpected non-error, output=: %v", output)
	}
}
