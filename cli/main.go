package main

import (
	"fmt"
	"os"

	"github.com/riotgames/key-conjurer/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
