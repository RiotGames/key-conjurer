package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapError(t *testing.T) {
	err := errors.New("verification failed")
	providerErr := WrapError(ErrFactorVerificationFailed, err)
	require.True(t, errors.Is(providerErr, ErrFactorVerificationFailed), "expected factor verification error")
	require.Contains(t, providerErr.Error(), err.Error(), "expected the original error message")
}
