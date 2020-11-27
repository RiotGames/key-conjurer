package cmd

import (
	"context"
	"fmt"

	api "github.com/riotgames/key-conjurer/api/keyconjurer"
	"github.com/riotgames/key-conjurer/cli/keyconjurer"

	"github.com/spf13/cobra"
)

var (
	ttl           uint
	timeRemaining uint
	outputType    string
	awsCliPath    string
	roleName      string
)

var (
	// outputTypeEnvironmentVariable indicates that keyconjurer will dump the credentials to stdout in Bash environment variable format
	outputTypeEnvironmentVariable = "env"
	// outputTypeAWSCredentialsFile indicates that keyconjurer will dump the credentials into the ~/.aws/credentials file.
	outputTypeAWSCredentialsFile = "awscli"
)

var permittedOutputTypes = []string{outputTypeAWSCredentialsFile, outputTypeEnvironmentVariable}

func init() {
	getCmd.Flags().UintVar(&ttl, "ttl", 1, "The key timeout in hours from 1 to 8.")
	getCmd.Flags().UintVarP(&timeRemaining, "time-remaining", "t", keyconjurer.DefaultTimeRemaining, "Request new keys if there are no keys in the environment or the current keys expire within <time-remaining> minutes. Defaults to 60.")
	getCmd.Flags().StringVarP(&outputType, "out", "o", outputTypeEnvironmentVariable, "Format to save new credentials in. Supported outputs: env, awscli")
	getCmd.Flags().StringVarP(&awsCliPath, "awscli", "", "~/.aws/", "Path for directory used by the aws-cli tool. Default is \"~/.aws\".")
	// TODO: This flag should only be required if the user has not used this command before.
	// If the user has used this command before, we should use their previously selected role persisted in userdata
	getCmd.Flags().StringVar(&roleName, "role", "", "The name of the role to assume.")
	getCmd.Flags().StringVar(&authProvider, "auth-provider", api.AuthenticationProviderOkta, "The authentication provider to use.")
}

var getCmd = &cobra.Command{
	Use:     "get <accountName/alias>",
	Short:   "Retrieves temporary AWS API credentials.",
	Long:    "Retrieves temporary AWS API credentials for the specified account.  It sends a push request to the first Duo device it finds associated with your account.",
	Example: "keyconjurer get <accountName/alias>",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := newClient()
		if err != nil {
			return err
		}

		creds, err := loadCredentialsFromFile()
		if err != nil {
			return err
		}

		valid := false
		for _, permitted := range permittedOutputTypes {
			if outputType == permitted {
				valid = true
			}
		}

		if !valid {
			return invalidValueError(outputType, permittedOutputTypes)
		}

		// make sure we enforce limit
		if ttl > 8 {
			ttl = 8
		}

		credentials, err := client.GetCredentials(ctx, &keyconjurer.GetCredentialsOptions{
			Credentials: creds,
			// TODO: We need to turn args[0] into an application ID for our authentication provider.
			// This either needs to happen on the client or the server
			// If the user provides `okta-test-one`, for example, that won't work because that's not an application within Okta.

			ApplicationID:          args[0],
			RoleName:               roleName,
			TimeoutInHours:         uint8(ttl),
			AuthenticationProvider: authProvider,
		})

		if err != nil {
			return err
		}

		switch outputType {
		case outputTypeEnvironmentVariable:
			credentials.PrintCredsForEnv()
		case outputTypeAWSCredentialsFile:
			acc := keyconjurer.Account{ID: args[0], Name: args[0]}
			newCliEntry := keyconjurer.NewAWSCliEntry(credentials, &acc)
			return keyconjurer.SaveAWSCredentialInCLI(awsCliPath, newCliEntry)
		default:
			return fmt.Errorf("%s is an invalid output type", outputType)
		}

		return nil
	}}
