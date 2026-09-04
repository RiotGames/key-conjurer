package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"log/slog"

	"github.com/riotgames/key-conjurer/command"
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
	args := os.Args
	if flag, ok := os.LookupEnv("KEYCONJURERFLAGS"); ok {
		args = append(args, strings.Split(flag, " ")...)
	}

	err := command.Execute(context.Background(), args)
	// err is only non-nil here if we reach this point with an error that did not implement cli.ExitCoder
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(command.ExitCodeUnknownError)
	}
}
