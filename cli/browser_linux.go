package main

import (
	"context"
	"os/exec"
)

func OpenBrowser(url string) error {
	proc := exec.CommandContext(context.Background(), "xdg-open", url)
	return proc.Run()
}
