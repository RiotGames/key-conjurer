//go:build darwin

package command

import (
	"errors"
	"os/exec"
)

func isKeychainLockedErr(err error) bool {
	var exitErr *exec.ExitError
	// keyring uses the 'security' binary, and that might exit with an error if it's not unlocked.
	return errors.As(err, &exitErr) && exitErr.ExitCode() == 36
}
