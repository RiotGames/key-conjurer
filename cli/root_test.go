package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Required to reset Cobra state between Test runs.
// If you add new tests that change state, you may need to
// add code here to reset the side effects of other tests
func resetCobra(cmd *cobra.Command) {

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

func execute(cmd *cobra.Command, args ...string) (string, error) {

	cmd.SetArgs(args)

	outputbuf := new(bytes.Buffer)

	cmd.SetOutput(outputbuf)

	err := cmd.Execute()

	return outputbuf.String(), err
}

func stringChecks(t *testing.T, testTarget string, shouldBeHere, shouldNotBeHere []string) {
	t.Helper()
	expectedTermsMissing := []string{}
	notExpectedTermsFound := []string{}

	for _, v := range shouldBeHere {
		if !strings.Contains(testTarget, v) {
			expectedTermsMissing = append(expectedTermsMissing, v)
		}
	}
	for _, v := range shouldNotBeHere {
		if strings.Contains(testTarget, v) {
			notExpectedTermsFound = append(notExpectedTermsFound, v)
		}
	}

	missing := ""
	found := ""

	if len(expectedTermsMissing) > 0 {
		missing = fmt.Sprintf("Content missing that was expected:\n   %v\n", expectedTermsMissing)
	}
	if len(notExpectedTermsFound) > 0 {
		found = fmt.Sprintf("Content found that was not expected:\n   %v\n", notExpectedTermsFound)
	}
	if len(notExpectedTermsFound) > 0 || len(expectedTermsMissing) > 0 {
		t.Errorf("\n%v%vString being checked:\n---------------\n%v\n---------------\n", missing, found, testTarget)
	}
}

func TestHelpCommand(t *testing.T) {

	resetCobra(rootCmd)

	output, err := execute(rootCmd, "help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{rootCmd.Long, "-s, --short-version", "-1, --oneline-version"},
		[]string{})
}

func TestHelpFlag(t *testing.T) {

	resetCobra(rootCmd)

	output, err := execute(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{rootCmd.Long, "-s, --short-version", "-1, --oneline-version"},
		[]string{})
}

func TestHelpShortFlag(t *testing.T) {

	resetCobra(rootCmd)

	output, err := execute(rootCmd, "-h")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{rootCmd.Long, "-s, --short-version", "-1, --oneline-version"},
		[]string{})
}

func TestHelpNoCommand(t *testing.T) {

	resetCobra(rootCmd)

	output, err := execute(rootCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{rootCmd.Long, "-s, --short-version", "-1, --oneline-version"},
		[]string{})
}

func TestVersionFlag(t *testing.T) {

	resetCobra(rootCmd)

	output, err := execute(rootCmd, "--version")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{"Version:", "Build Timestamp:", "Client:", "Default Hostname:", "Upgrade URL:", appname, Version, DownloadURL},
		[]string{})
}

func TestVersionShortFlag(t *testing.T) {

	resetCobra(rootCmd)

	output, err := execute(rootCmd, "-v")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{"Version:", "Build Timestamp:", "Client:", "Default Hostname:", "Upgrade URL:", appname, Version, DownloadURL},
		[]string{})
}

func TestOneLineVersionFlag(t *testing.T) {

	resetCobra(rootCmd)

	output, err := execute(rootCmd, "--oneline-version")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{appname, Version, "Client:", "(Build Timestamp:"},
		[]string{"Version:", "Default Hostname:", "Upgrade URL:", DownloadURL})
}

func TestOneLineVersionShortFlag(t *testing.T) {

	resetCobra(rootCmd)

	output, err := execute(rootCmd, "-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{appname, Version, "Client:", "(Build Timestamp:"},
		[]string{"Version:", "Default Hostname:", "Upgrade URL:", DownloadURL})
}

func TestShortVersionFlag(t *testing.T) {

	resetCobra(rootCmd)

	output, err := execute(rootCmd, "--short-version")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{appname, Version},
		[]string{"Version:", "Build Timestamp:", "Client:", "Default Hostname:", "Upgrade URL:", DownloadURL})
}

func TestShortVersionShortFlag(t *testing.T) {

	resetCobra(rootCmd)

	output, err := execute(rootCmd, "-s")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stringChecks(t, output,
		[]string{appname, Version},
		[]string{"Version:", "Build Timestamp:", "Client:", "Default Hostname:", "Upgrade URL:", DownloadURL})
}

func TestShortVersionLongInvalidArgs(t *testing.T) {

	resetCobra(rootCmd)

	output, err := execute(rootCmd, "--short-version", "get")
	if err != nil {
		if err.Error() != "unknown flag: --short-version" {
			t.Errorf("Unexpected error: %v", err)
		}
	} else {
		t.Errorf("Unexpected non-error, output=: %v", output)
	}
}

func TestHelpLongInvalidArgs(t *testing.T) {

	resetCobra(rootCmd)

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

	resetCobra(rootCmd)

	output, err := execute(rootCmd, "badcommand")
	if err != nil {
		if err.Error() != "unknown command \"badcommand\" for \""+appname+"\"" {
			t.Errorf("Unexpected error: %v", err)
		}
	} else {
		t.Errorf("Unexpected non-error, output=: %v", output)
	}
}

// Runs multiple tests that set flags & state and ensures the resetCobra() function
// is properly re-initializing state so test run as expected
func TestResetCobra(t *testing.T) {

	TestShortVersionShortFlag(t)
	TestOneLineVersionFlag(t)
	TestHelpShortFlag(t)
	TestVersionFlag(t)
	TestShortVersionShortFlag(t)
}
