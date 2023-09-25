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

	if err := rootCmd.Execute(); err == nil {
		return
	}

	os.Exit(1)
}
