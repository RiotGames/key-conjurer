package command

import (
	"encoding/json"
	"errors"
	"runtime"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

// ErrKeychainLocked indicates that the operating system keychain is locked and needs to be unlocked.
//
// This usually only occurs on Darwin systems.
type ErrKeychainLocked struct{}

func (e ErrKeychainLocked) Error() string {
	if runtime.GOOS == "darwin" {
		return "The keychain used to store secrets is locked. It can be unlocked with the following command: `security unlock-keychain`. You may be asked to enter your password."
	} else {
		return "Keychain locked"
	}
}

func (e ErrKeychainLocked) ExitCode() int {
	return ExitCodeUnknownError
}

var _ cli.ExitCoder = &ErrKeychainLocked{}

// keyringToken is a token stored in the operating system keyring.
//
// *oauth2.Token is not stored directly because it does not preserve the extra data (the id token)
// It's not generally recommended to store id tokens, but we need the id token to do our websso login wizardry.
type keyringToken struct {
	oauth2.Token
	IDToken string `json:"id_token"`
}

func checkKeychainLocked() bool {
	_, err := getAccountCredentialFromKeychain()
	return isKeychainLockedErr(err)
}

func getAccountCredentialFromKeychain() (*oauth2.Token, error) {
	buf, err := keyring.Get("keyconjurer", "accounts-credential")
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, ErrTokensExpiredOrAbsent
	} else if err != nil {
		return nil, err
	}

	var tok keyringToken
	if err := json.Unmarshal([]byte(buf), &tok); err != nil {
		// bad JSON format
		return nil, ErrTokensExpiredOrAbsent
	}

	if tok.Expiry.Before(time.Now()) {
		// Expired. Caught by Dan's patch review; without this an expired
		// cached token was silently returned as if still valid.
		return nil, ErrTokensExpiredOrAbsent
	}

	// This is how we expect to find the ID token in the access token.
	// Hacky, but this is also how OAuth2 APIs communicate it
	extra := map[string]any{"id_token": tok.IDToken}
	return tok.WithExtra(extra), nil
}

func putAccountCredentialInKeychain(tok *oauth2.Token, idToken string) error {
	tk := keyringToken{
		Token:   *tok,
		IDToken: idToken,
	}
	buf, _ := json.Marshal(tk)
	err := keyring.Set("keyconjurer", "accounts-credential", string(buf))
	if isKeychainLockedErr(err) {
		return &ErrKeychainLocked{}
	}
	return err
}

type keychainTokenSource struct{}

func (k *keychainTokenSource) Token() (*oauth2.Token, error) {
	return getAccountCredentialFromKeychain()
}
