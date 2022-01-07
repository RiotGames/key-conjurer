package okta

import (
	"errors"
	"net/http"
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

func TestTranslateError(t *testing.T) {
	require.Nil(t, TranslateError(nil, nil), "expected nil since error is nil")
	oktaErr := TranslateError(nil, errors.New("error text"))
	require.True(t, errors.Is(oktaErr, ErrOktaInternalServerError), "expected internal server error")
	require.Contains(t, oktaErr.Error(), ErrOktaInternalServerError.Error(), "expected internal server error text")
	require.Contains(t, oktaErr.Error(), "error text", "expected the original error text")
	oktaErr = TranslateError(&http.Response{StatusCode: 400}, errors.New("failed"))
	require.True(t, errors.Is(oktaErr, ErrOktaBadRequest), "expected bad request error")
	require.Contains(t, oktaErr.Error(), ErrOktaBadRequest.Error(), "expected bad request error text")
	require.Contains(t, oktaErr.Error(), "failed", "expected the original error text")
}

func TestWrapError(t *testing.T) {
	oktaErr := WrapError(errors.New("access denied"), ErrOktaForbidden)
	require.True(t, errors.Is(oktaErr, ErrOktaForbidden), "expected forbidden error")
	require.Contains(t, oktaErr.Error(), ErrOktaForbidden.Error(), "expected forbidden text")
	require.Contains(t, oktaErr.Error(), "access denied", "expected the original error text")
}
