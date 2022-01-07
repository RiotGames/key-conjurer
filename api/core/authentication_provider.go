// Package core which contains logic that is shared among different compilation units.
//
// This is primarily used to prevent cyclical compilation (forbidden in Go) between compilation units.
//
// Forgive the horizontal slicing, it's not great and is an anti-pattern in Go, but it's also very quick and does the job.
package core

import (
	"context"
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
}

// A Role is something a user can 'assume' when accessing an application.
//
// This stems from AWS terminology with their AssumeRolePolicy; it's possible this concept does not translate well with alternative cloud providers.
type Role struct {
	ID          string
	RoleName    string
	AccountName string
}

// An AuthenticationProvider is a component which will verify user credentials, list the applications a user is entitled to, the roles the user may assume within that application and generate SAML assertions for federation.
type AuthenticationProvider interface {
	// Authenticate should validate that the provided credentials are correct for a user.
	Authenticate(ctx context.Context, credentials Credentials) (User, error)
	// ListApplications should list all the applications the given user is entitled to access.
	ListApplications(ctx context.Context, user User) ([]Application, error)
	// GenerateSAMLAssertion should generate a SAML assertion that the user may exchange with the target application in order to gain access to it.
	GenerateSAMLAssertion(ctx context.Context, credentials Credentials, appID string) (*SAMLResponse, error)
}

// ProviderError is an error returned by an authentication provider.
type ProviderError error

// A list of standard errors that can be returned by an authentication provider.
var (
	ErrBadRequest                    ProviderError = errors.New("bad request")
	ErrApplicationNotFound           ProviderError = errors.New("application not found")
	ErrAuthenticationFailed          ProviderError = errors.New("unauthorized")
	ErrAccessDenied                  ProviderError = errors.New("access denied")
	ErrFactorVerificationFailed      ProviderError = errors.New("factor verification failed")
	ErrCouldNotSendMfaPush           ProviderError = errors.New("could not send MFA push")
	ErrSubmitChallengeResponseFailed ProviderError = errors.New("submit challenge response failed")
	ErrCouldNotCreateSession         ProviderError = errors.New("could not create a session")
	ErrSAMLError                     ProviderError = errors.New("failed to process SAML")
	ErrInternalError                 ProviderError = errors.New("internal error")
	ErrUnspecified                   ProviderError = errors.New("unspecified")
)

// NewBadRequestError creates a ErrBadRequest error with a specified message.
func NewBadRequestError(message string) error {
	return fmt.Errorf("%w: %s", ErrBadRequest, message)
}

// NewInternalError creates a ErrInternalError error with a specified message.
func NewInternalError(message string) error {
	return fmt.Errorf("%w: %s", ErrInternalError, message)
}

// NewAuthenticationProviderError wraps an error into a standard authentication provider error.
func NewAuthenticationProviderError(standardErr ProviderError, nestedErr error) error {
	return fmt.Errorf("%w: %s", standardErr, nestedErr.Error())
}
