package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"

	"log/slog"

	"github.com/riotgames/key-conjurer/command"
	"github.com/spf13/cobra"
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

	err := command.Execute(context.Background(), args)
	if IsWindowsPortAccessError(err) {
		fmt.Fprintf(os.Stderr, "Encountered an issue when opening the port for KeyConjurer: %s\n", err)
		fmt.Fprintln(os.Stderr, "Consider running `net stop hns` and then `net start hns`")
		os.Exit(command.ExitCodeConnectivityError)
	}

	if errors.Is(err, command.ErrKeychainLocked) && runtime.GOOS == "darwin" {
		fmt.Fprintln(os.Stderr, "The keychain used to store secrets is locked. It can be unlocked with the following command: `security unlock-keychain`. You may be asked to enter your password.")
		os.Exit(command.ExitCodeUnknownError)
	}

	if err != nil {
		cobra.CheckErr(err)

		errorCode, ok := command.GetExitCode(err)
		if !ok {
			errorCode = command.ExitCodeUnknownError
		}
		os.Exit(errorCode)
	}
}
