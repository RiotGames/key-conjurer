package vault

import hashivault "github.com/hashicorp/vault/api"

type Auth struct {
	LDAP    *LDAP
	IAM     *IAM
	Token   *Token
	AppRole *AppRole
}

func NewAuth(client *hashivault.Client) *Auth {
	return &Auth{
		LDAP: &LDAP{
			client: client},
		IAM: &IAM{
			client: client},
		Token: &Token{
			client: client},
		AppRole: &AppRole{
			client: client}}
}
