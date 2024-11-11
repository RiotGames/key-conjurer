package command

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

const (
	ExitCodeTokensExpiredOrAbsent int = 0x1
	ExitCodeUndisclosedOktaError  int = 0x2
	ExitCodeAuthenticationError   int = 0x3
	ExitCodeConnectivityError     int = 0x4
	ExitCodeValueError            int = 0x5
	ExitCodeAWSError              int = 0x6
	ExitCodeUnknownError          int = 0x7D
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
	ExitCode int
}

func (e genericError) Error() string {
	return e.Message
}

func (e genericError) Code() int {
	return e.ExitCode
}

type codeError interface {
	Error() string
	Code() int
}

// UsageError indicates that the user used the program incorrectly
type UsageError struct {
	ExitCode     int
	Description  string
	DebugMessage string
}

func (u UsageError) Error() string {
	return u.Description
}

func (u UsageError) Code() int {
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

func (v ValueError) Code() int {
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

func (o OktaError) Code() int {
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
	return fmt.Sprintf("%s: %s", o.Message, o.InnerError)
}

func (o AWSError) Code() int {
	return ExitCodeAWSError
}

type TimeToLiveError struct {
	MaxDuration       time.Duration
	RequestedDuration time.Duration
}

func (o TimeToLiveError) Code() int {
	return ExitCodeValueError
}

func (o TimeToLiveError) Error() string {
	if o.MaxDuration == 0 && o.RequestedDuration == 0 {
		// Duration is ambiguous/was not specified by AWS, so we return a generic message instead.
		return "the TTL you requested exceeds the maximum TTL for this configuration"
	}

	// We cast to int to discard decimal places
	return fmt.Sprintf("you requested a TTL of %d hours, but the maximum for this configuration is %d hours", int(o.RequestedDuration.Hours()), int(o.MaxDuration.Hours()))
}

// tryParseTimeToLiveError attempts to parse an error related to the DurationSeconds field in the STS request.
//
// If the given error does relate to the specified DurationSeconds being larger than MaxDurationSeconds, this function will return a more specific error than the one the AWS SDK provides, and returns true.
// Returns nil and false in all other situations.
func tryParseTimeToLiveError(err error) (error, bool) {
	var awsErr awserr.Error
	if errors.As(err, &awsErr) && awsErr.Code() == "ValidationError" {
		var providedValue, maxValue time.Duration
		// This is no more specific type than this, and yes, unfortunately the error message includes the count.
		formatOne := "1 validation error detected: Value '%d' at 'durationSeconds' failed to satisfy constraint: Member must have value less than or equal to %d"
		if n, parseErr := fmt.Sscanf(awsErr.Message(), formatOne, &providedValue, &maxValue); parseErr == nil && n == 2 {
			return TimeToLiveError{MaxDuration: maxValue * time.Second, RequestedDuration: providedValue * time.Second}, true
		}

		formatAmbiguousMaximum := "The requested DurationSeconds exceeds the MaxSessionDuration set for this role."
		if strings.Compare(awsErr.Message(), formatAmbiguousMaximum) == 0 {
			return TimeToLiveError{MaxDuration: 0, RequestedDuration: 0}, true
		}
	}

	return nil, false
}

func GetExitCode(err error) (int, bool) {
	var codeError codeError
	if errors.As(err, &codeError) {
		return codeError.Code(), true
	}
	return 0, false
}
