package okta

import (
	"errors"
	"testing"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/stretchr/testify/require"
)

func TestGetMessage(t *testing.T) {
	oktaErr := okta.Error{ErrorCode: "E01", ErrorSummary: "okta error"}
	require.Equal(t, "okta error", GetMessage(&oktaErr), "expected only error summary")
	err := errors.New("error description")
	require.Equal(t, "error description", GetMessage(err))
}

func TestNewSAMLError(t *testing.T) {
	err := errors.New("error description")
	samlErr := NewSAMLError(err)
	require.True(t, errors.Is(samlErr, ErrOktaSAMLError))
	require.Contains(t, samlErr.Error(), ErrOktaSAMLError.Error(), "expected SAML error")
	require.Contains(t, samlErr.Error(), err.Error(), "expected error description")
}
