package main

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

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

	err := rootCmd.Execute()
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
