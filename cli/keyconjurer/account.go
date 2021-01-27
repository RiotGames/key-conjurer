package keyconjurer

import (
	"strings"
)

// Account is used to store information related to the AWS OneLogin App/AWS Account
type Account struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

func (a *Account) NormalizeName() string {
	return strings.Replace(a.Name, "AWS - ", "", -1)
}

func (a *Account) DefaultAlias() {
	if a.Alias == "" {
		alias := strings.Replace(a.Name, "AWS - ", "", -1)
		alias = strings.Split(alias, " ")[0]
		a.Alias = strings.ToLower(alias)
	}
}

func (a *Account) IsNameMatch(name string) bool {
	// Purposefully not checking the lowercase version of app.Alias
	//  as the user should match the alias provided
	if strings.ToLower(a.Name) == strings.ToLower(name) {
		return true
	}

	if strings.ToLower(a.NormalizeName()) == strings.ToLower(name) {
		return true
	}

	if a.Alias == name {
		return true
	}

	return false
}

func (a *Account) SetAlias(alias string) {
	if alias == "" {
		a.DefaultAlias()
	} else {
		a.Alias = alias
	}
}
