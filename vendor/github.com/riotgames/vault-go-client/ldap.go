package vault

import (
	"errors"
	"fmt"
	"strings"

	hashivault "github.com/hashicorp/vault/api"
)

type LDAP struct {
	client *hashivault.Client
}

type LDAPLoginOptions struct {
	Username  string
	Password  string
	MountPath string
}

func (l *LDAP) Login(options LDAPLoginOptions) (*hashivault.Secret, error) {
	authSecret, err := l.ldapLogin(options)

	if err != nil {
		return nil, err
	}

	if authSecret.Auth == nil {
		return nil, errors.New("Vault LDAP Auth returned nil")
	}

	l.client.SetToken(authSecret.Auth.ClientToken)
	return authSecret, nil
}

func (l *LDAP) ldapLogin(options LDAPLoginOptions) (*hashivault.Secret, error) {
	ldapCreds := map[string]interface{}{
		"password": options.Password,
	}
	pathFormatString := "auth/ldap/login/%s"
	if options.MountPath != "" {
		pathFormatString = "auth/" + strings.Trim(options.MountPath, "/") + "/login/%s"
	}
	normalizedPath := fmt.Sprintf(pathFormatString, options.Username)

	authSecret, err := l.client.Logical().Write(normalizedPath, ldapCreds)

	if err != nil {
		return nil, err
	}

	if authSecret == nil {
		return nil, errors.New("empty response from vault ldap")
	}

	return authSecret, nil
}
