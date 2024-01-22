package main

import (
	"os"
	"strings"

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
	if e, ok := err.(codeError); ok {
		rootCmd.PrintErrln(e.Error())
		os.Exit(int(e.Code()))
	} else if err != nil {
		rootCmd.PrintErrf("An unexpected error occurred: %s", err.Error())
		os.Exit(ExitCodeUnknownError)
	}
}
