package main

import (
	"fmt"
	"strings"
)

const (
	ExitCodeTokensExpiredOrAbsent uint8 = 0x1
	ExitCodeUndisclosedOktaError        = 0x2
	ExitCodeAuthenticationError         = 0x3
	ExitCodeConnectivityError           = 0x4
	ExitCodeValueError                  = 0x5
	ExitCodeUnknownError                = 0x7D
)

var (
	ErrTokensExpiredOrAbsent = UsageError{
		Code:         ExitCodeTokensExpiredOrAbsent,
		DebugMessage: "tokens expired or absent",
		Description:  "Your session has expired. Please login again.",
	}
)

type codeError interface {
	Error() string
	Code() uint8
}

// UsageError indicates that the user used the program incorrectly
type UsageError struct {
	Code         uint8
	Description  string
	DebugMessage string
}

func (u UsageError) Error() string {
	return u.Description
}

type ValueError struct {
	Value       string
	ValidValues []string
}

func (v ValueError) Error() string {
	var quoted []string
	for _, v := range v.ValidValues {
		quoted = append(quoted, fmt.Sprintf("%q", v))
	}

	acceptable := strings.Join(quoted, ",")
	return fmt.Sprintf("provided value %s was not valid (accepted values: %s)", v.Value, acceptable)
}

func (v ValueError) Code() uint8 {
	return ExitCodeValueError
}
