package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"log/slog"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

const (
	// WSAEACCES is the Windows error code for attempting to access a socket that you don't have permission to access.
	//
	// This commonly occurs if the socket is in use or was not closed correctly, and can be resolved by restarting the hns service.
	WSAEACCES = 10013
)

// IsWindowsPortAccessError determines if the given error is the error WSAEACCES.
func IsWindowsPortAccessError(err error) bool {
	var syscallErr *syscall.Errno
	return errors.As(err, &syscallErr) && *syscallErr == WSAEACCES
}

func init() {
	var opts slog.HandlerOptions
	if os.Getenv("DEBUG") == "1" {
		opts.Level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &opts)
	slog.SetDefault(slog.New(handler))
}

func main() {
	args := os.Args[1:]
	if flag, ok := os.LookupEnv("KEYCONJURERFLAGS"); ok {
		args = append(args, strings.Split(flag, " ")...)
	}
	rootCmd.SetArgs(args)

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, NewHTTPClient())
	err := rootCmd.ExecuteContext(ctx)
	if IsWindowsPortAccessError(err) {
		fmt.Fprintf(os.Stderr, "Encountered an issue when opening the port for KeyConjurer: %s\n", err)
		fmt.Fprintln(os.Stderr, "Consider running `net stop hns` and then `net start hns`")
		os.Exit(ExitCodeConnectivityError)
	}

	var codeErr codeError
	if errors.As(err, &codeErr) {
		cobra.CheckErr(codeErr)
		os.Exit(int(codeErr.Code()))
	} else if err != nil {
		// Probably a cobra error.
		cobra.CheckErr(err)
		os.Exit(ExitCodeUnknownError)
	}
}
