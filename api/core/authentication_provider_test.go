package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewBadRequestError(t *testing.T) {
	err := NewBadRequestError("error message")
	require.True(t, errors.Is(err, ErrBadRequest), "expected bad request error")
	require.Contains(t, err.Error(), "error message", "expected the original error message")
}

func TestNewInternalError(t *testing.T) {
	err := NewInternalError("error message")
	require.True(t, errors.Is(err, ErrInternalError), "expected internal error")
	require.Contains(t, err.Error(), "error message", "expected the original error message")
}

func TestNewAuthenticationProviderError(t *testing.T) {
	err := errors.New("verification failed")
	providerErr := NewAuthenticationProviderError(ErrFactorVerificationFailed, err)
	require.True(t, errors.Is(providerErr, ErrFactorVerificationFailed), "expected factor verification error")
	require.Contains(t, providerErr.Error(), err.Error(), "expected the original error message")
}
