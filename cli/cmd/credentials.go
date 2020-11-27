package cmd

import (
	"github.com/riotgames/key-conjurer/api/core"
	"github.com/riotgames/key-conjurer/cli/keyconjurer"
)

func loadCredentialsFromFile() (core.Credentials, error) {
	var userData keyconjurer.UserData
	if err := userData.LoadFromFile(keyConjurerRcPath); err != nil {
		// If the user has no credentials saved we prompt them for an error.
		// We do this because we do not want to save unencrypted credentials to the users machine
		// The alternative would be to always log the user in every time they use this endpoint which would give us encrypted credentials at the cost of adding overhead.
		// As a user will very likely only log in to one account at once, it's probably fine for us to do this instead.
		return core.Credentials{}, ErrNoCredentials
	}

	return userData.GetCredentials(), nil
}
