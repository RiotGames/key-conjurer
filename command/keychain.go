package command

import (
	"encoding/json"
	"errors"
	"os/exec"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

// ErrKeychainLocked indicates that the operating system keychain is locked and needs to be unlocked.
//
// This usually only occurs on Darwin systems.
var ErrKeychainLocked = errors.New("keychain locked")

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
	return err != nil && errors.Is(err, ErrTokensExpiredOrAbsent)
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
	// On Darwin, keyring uses the 'security' binary, and that might exit with an error if it's not unlocked.
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 36 {
		return ErrKeychainLocked
	}
	return err
}

type keychainTokenSource struct{}

func (k *keychainTokenSource) Token() (*oauth2.Token, error) {
	return getAccountCredentialFromKeychain()
}
