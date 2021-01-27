package main

import (
	"fmt"
	"strings"
)

func invalidValueError(value string, validValues []string) error {
	var quoted []string
	for _, v := range validValues {
		quoted = append(quoted, fmt.Sprintf("%q", v))
	}

	acceptable := strings.Join(quoted, ",")
	help := fmt.Sprintf("provided value %s was not valid (accepted values: %s)", value, acceptable)
	return &UsageError{
		ShortMessage: "invalid_value",
		Help:         help,
	}
}

// UsageError indicates that the user used the program incorrectly
type UsageError struct {
	// ShortMessage is not currently used, but is intended to be used when debugging. It is not displayed to users.
	ShortMessage string
	// Help is displayed to the user when the error message occurs over a tty. It should be one sentence and inform the user how to resolve the problem.
	Help string
}

func (u *UsageError) Error() string {
	return u.ShortMessage
}

var (
	// ErrNoCredentials indicates the user attempted to use a command that requires credentials to be stored on disk but had not logged in beforehand.
	ErrNoCredentials  error = &UsageError{ShortMessage: "no credentials", Help: "You must log in using `keyconjurer login` before using this command"}
	ErrNoRoleProvided error = &UsageError{ShortMessage: "no role provided", Help: "The --role flag must be specified when using this command"}
)
