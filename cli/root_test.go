package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Required to reset Cobra state between Test runs.
// If you add new tests that change state, you may need to
// add code here to reset the side effects of other tests
func resetCobra(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	cmdShortVersionFlag = false
	cmdOneLineVersionFlag = false

	helpflag := cmd.Flags().Lookup("help")
	if helpflag != nil {
		helpflag.Value.Set("false")
	}

	verflag := cmd.Flags().Lookup("version")
	if verflag != nil {
		verflag.Value.Set("false")
	}
}

func execute(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()

	resetCobra(t, cmd)

	cmd.SetArgs(args)

	outputbuf := new(bytes.Buffer)

	cmd.SetOutput(outputbuf)

	err := cmd.Execute()

	return outputbuf.String(), err
}

func stringContains(t *testing.T, testTarget, shouldBeHere string) {
	t.Helper()
	if !strings.Contains(testTarget, shouldBeHere) {
		t.Errorf("Missing Content:\n   [%v]\nShould have been in here:\n   %v\n", shouldBeHere, testTarget)
	}
}

func stringOmits(t *testing.T, testTarget, shouldNotBeHere string) {
	t.Helper()
	if strings.Contains(testTarget, shouldNotBeHere) {
		t.Errorf("Extra Content that should not be found here:\n   [%v]\nBut it was:\n   %v", shouldNotBeHere, testTarget)
	}
}

func stringChecks(t *testing.T, testTarget string, shouldBeHere, shouldNotBeHere []string) {
	t.Helper()

	outputlogged := false
	grouplogged := false

	for _, v := range shouldBeHere {
		if !strings.Contains(testTarget, v) {
			if !outputlogged {
				outputlogged = true
				t.Logf("String being checked:\n===========\n%v\n===========\n", testTarget)
			}
			if !grouplogged {
				grouplogged = true
				t.Logf("Content missing that was expected:\n")
			}
			t.Errorf("Should be (%v)\n", v)
		}
	}

	grouplogged = false

	for _, v := range shouldNotBeHere {
		if strings.Contains(testTarget, v) {
			if !outputlogged {
				outputlogged = true
				t.Logf("String being checked:\n===========\n%v\n===========\n", testTarget)
			}
			if !grouplogged {
				grouplogged = true
				t.Logf("Content found that was not expected:\n")
			}
			t.Errorf("Should NOT be (%v)\n", v)
		}
	}
}

func TestHelpCommand(t *testing.T) {

	output, err := execute(t, rootCmd, "help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{rootCmd.Long, "-s, --short-version", "-1, --oneline-version"},
		[]string{})
}

func TestHelpFlag(t *testing.T) {

	output, err := execute(t, rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{rootCmd.Long, "-s, --short-version", "-1, --oneline-version"},
		[]string{})
}

func TestHelpShortFlag(t *testing.T) {

	output, err := execute(t, rootCmd, "-h")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{rootCmd.Long, "-s, --short-version", "-1, --oneline-version"},
		[]string{})
}

func TestHelpNoCommand(t *testing.T) {

	output, err := execute(t, rootCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{rootCmd.Long, "-s, --short-version", "-1, --oneline-version"},
		[]string{})
}

func TestVersionFlag(t *testing.T) {

	output, err := execute(t, rootCmd, "--version")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{"Version:", "Build Timestamp:", "Client:", "Default Hostname:", "Upgrade URL:", appname, Version, DownloadURL},
		[]string{})
}

func TestVersionShortFlag(t *testing.T) {

	output, err := execute(t, rootCmd, "-v")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{"Version:", "Build Timestamp:", "Client:", "Default Hostname:", "Upgrade URL:", appname, Version, DownloadURL},
		[]string{})
}

func TestOneLineVersionFlag(t *testing.T) {

	output, err := execute(t, rootCmd, "--oneline-version")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{appname, Version, "Client:", "(Build Timestamp:"},
		[]string{"Version:", "Default Hostname:", "Upgrade URL:", DownloadURL})
}

func TestOneLineVersionShortFlag(t *testing.T) {

	output, err := execute(t, rootCmd, "-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{appname, Version, "Client:", "(Build Timestamp:"},
		[]string{"Version:", "Default Hostname:", "Upgrade URL:", DownloadURL})
}

func TestShortVersionFlag(t *testing.T) {

	output, err := execute(t, rootCmd, "--short-version")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{appname, Version},
		[]string{"Version:", "Build Timestamp:", "Client:", "Default Hostname:", "Upgrade URL:", DownloadURL})
}

func TestShortVersionShortFlag(t *testing.T) {

	output, err := execute(t, rootCmd, "-s")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{appname, Version},
		[]string{"Version:", "Build Timestamp:", "Client:", "Default Hostname:", "Upgrade URL:", DownloadURL})
}

func TestShortVersionLongInvalidArgs(t *testing.T) {

	output, err := execute(t, rootCmd, "--short-version", "get")
	if err != nil {
		if err.Error() != "unknown flag: --short-version" {
			t.Errorf("Unexpected error: %v", err)
		}
	} else {
		t.Errorf("Unexpected non-error, output=: %v", output)
	}
}

func TestHelpLongInvalidArgs(t *testing.T) {

	output, err := execute(t, rootCmd, "get", "-s")
	if err != nil {
		if err.Error() != "unknown shorthand flag: 's' in -s" {
			t.Errorf("Unexpected error: %v", err)
		}
	} else {
		t.Errorf("Unexpected non-error, output=: %v", output)
	}
}

func TestInvalidCommand(t *testing.T) {

	output, err := execute(t, rootCmd, "badcommand")
	if err != nil {
		if err.Error() != "unknown command \"badcommand\" for \""+appname+"\"" {
			t.Errorf("Unexpected error: %v", err)
		}
	} else {
		t.Errorf("Unexpected non-error, output=: %v", output)
	}
}

// Runs multiple tests that set flags & state and ensures the resetCobra() function
// is properly re-initializing state so tests run as expected
func TestResetCobra(t *testing.T) {

	TestShortVersionShortFlag(t)
	TestOneLineVersionFlag(t)
	TestHelpShortFlag(t)
	TestVersionFlag(t)
	TestShortVersionShortFlag(t)
}
