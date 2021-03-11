package main

import (
	"fmt"
	"os"
)

func logInfo(msg string, args ...interface{}) {
	if quiet {
		return
	}

	fmt.Fprintf(os.Stdout, msg, args...)
	fmt.Fprint(os.Stdout, "\n")
}
