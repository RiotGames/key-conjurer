package okta

import (
	"errors"
	"testing"

	"github.com/riotgames/key-conjurer/api/core"
	"github.com/stretchr/testify/require"
)

func TestTranslateOktaError(t *testing.T) {
	require.Error(t, core.ErrBadRequest, translateOktaError(ErrOktaBadRequest, core.ErrUnspecified), "expected bad request error")
	require.Error(t, core.ErrAuthenticationFailed, translateOktaError(ErrOktaUnauthorized, core.ErrUnspecified), "expected authentication error")
	require.Error(t, core.ErrAccessDenied, translateOktaError(ErrOktaForbidden, core.ErrUnspecified), "expected access denied error")
	require.Error(t, core.ErrInternalError, translateOktaError(ErrOktaInternalServerError, core.ErrUnspecified), "expected internal server error")
	require.Error(t, core.ErrUnspecified, translateOktaError(errors.New("another error"), core.ErrUnspecified), "expected unspecified error")
	require.NoError(t, translateOktaError(nil, core.ErrUnspecified), "no error expected since no error passed")
}

func TestWrapOktaError(t *testing.T) {
	providerErr := wrapOktaError(ErrOktaBadRequest, core.ErrUnspecified)
	require.True(t, errors.Is(providerErr, core.ErrBadRequest), "expected bad request error")
	require.Contains(t, providerErr.Error(), ErrOktaBadRequest.Error(), "error message should contain the bad request error from Okta")
	err := errors.New("another error")
	providerErr = wrapOktaError(err, core.ErrUnspecified)
	require.True(t, errors.Is(providerErr, core.ErrUnspecified), "expected unspecified error")
	require.Contains(t, providerErr.Error(), err.Error(), "error message should contain the original")
}
