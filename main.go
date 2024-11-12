package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"log/slog"

	"github.com/riotgames/key-conjurer/command"
	"github.com/spf13/cobra"
)

func main() {
	var opts slog.HandlerOptions
	if os.Getenv("DEBUG") == "1" {
		opts.Level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &opts)
	slog.SetDefault(slog.New(handler))

	args := os.Args[1:]
	if flag, ok := os.LookupEnv("KEYCONJURERFLAGS"); ok {
		args = append(args, strings.Split(flag, " ")...)
	}

	err := command.Execute(args)
	if isWindowsPortAccessError(err) {
		fmt.Fprintf(os.Stderr, "Encountered an issue when opening the port for KeyConjurer: %s\n", err)
		fmt.Fprintln(os.Stderr, "Consider running `net stop hns` and then `net start hns`")
		os.Exit(command.ExitCodeConnectivityError)
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

const (
	// wsaeacces is the Windows error code for attempting to access a socket that you don't have permission to access.
	//
	// This commonly occurs if the socket is in use or was not closed correctly, and can be resolved by restarting the hns service.
	wsaeacces = 10013
)

// isWindowsPortAccessError determines if the given error is the error wsaeacces.
func isWindowsPortAccessError(err error) bool {
	var syscallErr *syscall.Errno
	return errors.As(err, &syscallErr) && *syscallErr == wsaeacces
}
