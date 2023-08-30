package main

import (
	"errors"
	"fmt"
	"os"

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
	err := rootCmd.Execute()
	if err == nil {
		return
	}

	var usageErr *UsageError
	if errors.As(err, &usageErr) {
		fmt.Fprintln(os.Stderr, usageErr.Help)
		os.Exit(1)
		return
	}

	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
