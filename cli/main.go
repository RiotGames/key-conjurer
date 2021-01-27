package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	err := Execute()
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
