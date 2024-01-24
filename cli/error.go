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
	ExitCodeAWSError                    = 0x6
	ExitCodeUnknownError                = 0x7D
)

var (
	ErrTokensExpiredOrAbsent = UsageError{
		ExitCode:     ExitCodeTokensExpiredOrAbsent,
		DebugMessage: "tokens expired or absent",
		Description:  "Your session has expired. Please login again.",
	}
)

type genericError struct {
	Message  string
	ExitCode uint8
}

func (e genericError) Error() string {
	return e.Message
}

func (e genericError) Code() uint8 {
	return e.ExitCode
}

type codeError interface {
	Error() string
	Code() uint8
}

// UsageError indicates that the user used the program incorrectly
type UsageError struct {
	ExitCode     uint8
	Description  string
	DebugMessage string
}

func (u UsageError) Error() string {
	return u.Description
}

func (u UsageError) Code() uint8 {
	return u.ExitCode
}

func UnknownRoleError(role, applicationID string) error {
	return genericError{
		Message:  fmt.Sprintf("You do not have access to the role %s on application %s", role, applicationID),
		ExitCode: ExitCodeValueError,
	}
}

func UnknownAccountError(accountID, bypassCacheFlag string) error {
	return genericError{
		Message:  fmt.Sprintf("%q is not a known account name in your account cache. Your cache can be refreshed by entering executing `keyconjurer accounts`. If the value provided is an Okta application ID, you may provide --%s as an option to this command and try again.", accountID, bypassCacheFlag),
		ExitCode: ExitCodeValueError,
	}
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

type OktaError struct {
	InnerError error
	Message    string
}

func (o OktaError) Unwrap() error {
	return o.InnerError
}

func (o OktaError) Error() string {
	return o.Message
}

func (o OktaError) Code() uint8 {
	return ExitCodeUndisclosedOktaError
}

type AWSError struct {
	InnerError error
	Message    string
}

func (o AWSError) Unwrap() error {
	return o.InnerError
}

func (o AWSError) Error() string {
	return o.Message
}

func (o AWSError) Code() uint8 {
	return ExitCodeAWSError
}
