package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/riotgames/key-conjurer/api/core"
	"github.com/riotgames/key-conjurer/api/keyconjurer"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/spf13/cobra"
)

var errUnableToReadUsername = errors.New("unable to read username")

// getUsernameAndPassword prompts the user for their username and password via stdin
func getUsernameAndPassword(r io.Reader) (string, string, error) {
	scanner := bufio.NewScanner(r)
	fmt.Printf("username: ")
	username := ""
	if scanner.Scan() {
		username = scanner.Text()
	} else {
		return "", "", errUnableToReadUsername
	}

	fmt.Printf("password: ")
	bytes, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", fmt.Errorf("unable to get password: %w", err)
	}

	password := string(bytes)
	// Need to add our own newline
	fmt.Println()
	return username, password, nil
}

func promptForCredentials(r io.Reader) (core.Credentials, error) {
	username, password, err := getUsernameAndPassword(r)
	return core.Credentials{Username: username, Password: password}, err
}

func init() {
	loginCmd.Flags().StringVar(&identityProvider, "identity-provider", keyconjurer.AuthenticationProviderOkta, "The identity provider to use.")
}

var loginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Authenticate with KeyConjurer.",
	Long:    "Login using your AD creds. This stores encrypted credentials on the local system.",
	Example: "keyconjurer login",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := newClient()
		if err != nil {
			return err
		}

		creds, err := promptForCredentials(os.Stdin)
		if err != nil {
			return err
		}

		data, err := client.GetUserData(ctx, &GetUserDataOptions{
			Credentials:            creds,
			AuthenticationProvider: identityProvider,
		})

		if err != nil {
			return err
		}

		config.Creds = data.EncryptedCredentials
		var entries []Account
		for _, acc := range data.Apps {
			entries = append(entries, Account{ID: acc.ID, Name: acc.Name, Alias: generateDefaultAlias(acc.Name)})
		}

		config.Accounts.ReplaceWith(entries)
		return nil
	},
}
