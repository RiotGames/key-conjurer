// Package core which contains logic that is shared among different compilation units.
//
// This is primarily used to prevent cyclical compilation (forbidden in Go) between compilation units.
//
// Forgive the horizontal slicing, it's not great and is an anti-pattern in Go, but it's also very quick and does the job.
package core

import (
	"errors"
	"fmt"
)

// Credentials is a struct which contains the username and password for a user.
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Encrypted indicates whether or not the credentials are encrypted
func (c Credentials) Encrypted() bool {
	return c.Username == "encrypted"
}

// A User represents a user from an Authentication Provider.
// This struct is used to guarantee that the user has originated from an authenticator, rather than being a POGO
type User struct {
	ID string
}

// An Application is some SAML-enabled service that a user is entitled to.
type Application struct {
	// LegacyID is used to enable legacy support for the old key-conjurer clients.
	// This is not used past KeyConjurer version 2
	LegacyID uint   `json:"id"`
	ID       string `json:"@id"`
	Name     string `json:"name"`

	// Href is a link that can be visited to kick off the authentication flow for this application.
	Href string `json:"href"`
}

// A Role is something a user can 'assume' when accessing an application.
//
// This stems from AWS terminology with their AssumeRolePolicy; it's possible this concept does not translate well with alternative cloud providers.
type Role struct {
	ID          string
	RoleName    string
	AccountName string
}

// A list of standard errors that can be returned by an authentication provider.
var (
	ErrBadRequest                    = errors.New("bad request")
	ErrApplicationNotFound           = errors.New("application not found")
	ErrAuthenticationFailed          = errors.New("unauthorized")
	ErrAccessDenied                  = errors.New("access denied")
	ErrFactorVerificationFailed      = errors.New("factor verification failed")
	ErrCouldNotSendMfaPush           = errors.New("could not send MFA push")
	ErrSubmitChallengeResponseFailed = errors.New("submit challenge response failed")
	ErrCouldNotCreateSession         = errors.New("could not create a session")
	ErrSAMLError                     = errors.New("failed to process SAML")
	ErrInternalError                 = errors.New("internal error")
	ErrUnspecified                   = errors.New("unspecified")
)

// WrapError wraps an error into a standard authentication provider error.
func WrapError(standardErr error, nestedErr error) error {
	return fmt.Errorf("%w: %s", standardErr, nestedErr.Error())
}
