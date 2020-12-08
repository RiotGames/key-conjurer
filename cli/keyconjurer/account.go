package keyconjurer

import (
	"strings"
)

// Account is used to store information related to the AWS OneLogin App/AWS Account
type Account struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

func (a *Account) normalizeName() string {
	return strings.Replace(a.Name, "AWS - ", "", -1)
}

func (a *Account) defaultAlias() {
	if a.Alias == "" {
		alias := strings.Replace(a.Name, "AWS - ", "", -1)
		alias = strings.Split(alias, " ")[0]
		a.Alias = strings.ToLower(alias)
	}
}

func (a *Account) isNameMatch(name string) bool {
	// Purposefully not checking the lowercase version of app.Alias
	//  as the user should match the alias provided
	if strings.ToLower(a.Name) == strings.ToLower(name) {
		return true
	}

	if strings.ToLower(a.normalizeName()) == strings.ToLower(name) {
		return true
	}

	if a.Alias == name {
		return true
	}

	return false
}

func (a *Account) setAlias(alias string) {
	if alias == "" {
		a.defaultAlias()
	} else {
		a.Alias = alias
	}
}

// App is being depricated in favor of Accounts
//   to keep underlying data structure names the same as cli names
type App struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

func (a *App) normalizeName() string {
	return strings.Replace(a.Name, "AWS - ", "", -1)
}

func (a *App) defaultAlias() {
	if a.Alias == "" {
		alias := strings.Replace(a.Name, "AWS - ", "", -1)
		alias = strings.Split(alias, " ")[0]
		a.Alias = alias
	}
}

func (a *App) isNameMatch(name string) bool {
	// Purposefully not checking the lowercase version of app.Alias
	//  as the user should match the alias provided
	return strings.ToLower(a.Name) == strings.ToLower(name) ||
		strings.ToLower(a.normalizeName()) == strings.ToLower(name) ||
		a.Alias == name
}

func (a *App) setAlias(alias string) {
	a.Alias = alias
}
