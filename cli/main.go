package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/riotgames/key-conjurer/cli/cmd"
)

func main() {
	err := cmd.Execute()
	if err == nil {
		return
	}

	var usageErr *cmd.UsageError
	if errors.As(err, &usageErr) {
		fmt.Fprintln(os.Stderr, usageErr.Help)
		os.Exit(1)
		return
	}

	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
