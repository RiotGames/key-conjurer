package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	ttl            uint
	timeRemaining  uint
	outputType     string
	awsCliPath     string
	tencentCliPath string
	roleName       string
	cloudFlag      string
	shell          string = shellTypeInfer
)

var (
	// outputTypeEnvironmentVariable indicates that keyconjurer will dump the credentials to stdout in Bash environment variable format
	outputTypeEnvironmentVariable = "env"
	// outputTypeAWSCredentialsFile indicates that keyconjurer will dump the credentials into the ~/.aws/credentials file.
	outputTypeAWSCredentialsFile = "awscli"
	// outputTypeTencentCredentialsFile indicates that keyconjurer will dump the credentials into the ~/.tencent/credentials file.
	outputTypeTencentCredentialsFile = "tencentcli"
	permittedOutputTypes             = []string{outputTypeAWSCredentialsFile, outputTypeEnvironmentVariable, outputTypeTencentCredentialsFile}
	permittedShellTypes              = []string{shellTypePowershell, shellTypeBash, shellTypeBasic, shellTypeInfer}
)

func init() {
	getCmd.Flags().UintVar(&ttl, "ttl", 1, "The key timeout in hours from 1 to 8.")
	getCmd.Flags().UintVarP(&timeRemaining, "time-remaining", "t", DefaultTimeRemaining, "Request new keys if there are no keys in the environment or the current keys expire within <time-remaining> minutes. Defaults to 60.")
	getCmd.Flags().StringVarP(&outputType, "out", "o", outputTypeEnvironmentVariable, "Format to save new credentials in. Supported outputs: env, awscli,tencentcli")
	getCmd.Flags().StringVarP(&shell, "shell", "", shellTypeInfer, "If output type is env, determines which format to output credentials in - by default, the format is inferred based on the execution environment. WSL users may wish to overwrite this to `bash`")
	getCmd.Flags().StringVarP(&awsCliPath, "awscli", "", "~/.aws/", "Path for directory used by the aws-cli tool. Default is \"~/.aws\".")
	getCmd.Flags().StringVarP(&tencentCliPath, "tencentcli", "", "~/.tencent/", "Path for directory used by the tencent-cli tool. Default is \"~/.tencent\".")
	getCmd.Flags().StringVar(&roleName, "role", "", "The name of the role to assume.")
	getCmd.Flags().StringVarP(&cloudFlag, "cloud", "", "aws", "Choose a cloud vendor. Default is aws. Can choose aws or tencent")
	getCmd.Flags().StringVar(&identityProvider, "identity-provider", defaultIdentityProvider, "The identity provider to use. Refer to `"+appname+" identity-providers` for more info.")
}

var getCmd = &cobra.Command{
	Use:   "get <accountName/alias>",
	Short: "Retrieves temporary Cloud(AWS|Tencent) API credentials.",
	Long: `Retrieves temporary Cloud(AWS|Tencent) API credentials for the specified account.  It sends a push request to the first Duo device it finds associated with your account.

	A role must be specified when using this command through the --role flag. You may list the roles you can assume through the roles command.`,
	// Example: appname + " get <accountName/alias>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := newClient()
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

		valid = false
		for _, permitted := range permittedShellTypes {
			if shell == permitted {
				valid = true
			}
		}

		if !valid {
			return invalidValueError(shell, permittedShellTypes)
		}

		creds, err := config.GetCredentials()
		if err != nil {
			return err
		}

		// make sure we enforce limit
		if ttl > 8 {
			ttl = 8
		}

		var label, applicationID = args[0], args[0]
		account, ok := config.FindAccount(applicationID)
		if ok {
			applicationID = account.ID
			label = account.Name
		} else {
			account = &Account{}
		}

		if account.MostRecentRole != "" && roleName == "" {
			roleName = account.MostRecentRole
		}

		if config.TimeRemaining != 0 && timeRemaining == DefaultTimeRemaining {
			timeRemaining = config.TimeRemaining
		}

		if cloudFlag == "" {
			cloudFlag = cloudAws
			if strings.Contains(account.Name, "Tencent") {
				cloudFlag = cloudTencent
			}
		}

		var credentials CloudCredentials
		credentials.LoadFromEnv(cloudFlag)
		if credentials.ValidUntil(*account, cloudFlag, time.Duration(timeRemaining)*time.Minute) {
			return echoCredentials(args[0], args[0], credentials, outputType, cloudFlag)
		}

		if !quiet {
			fmt.Fprintf(os.Stderr, "sending authentication request for account %q - you may be asked to authenticate with Duo\n", label)
		}

		if ttl == 1 && config.TTL != 0 {
			ttl = config.TTL
		}

		credentials, err = client.GetCredentials(ctx, &GetCredentialsOptions{
			Credentials:            creds,
			ApplicationID:          applicationID,
			RoleName:               roleName,
			TimeoutInHours:         uint8(ttl),
			AuthenticationProvider: identityProvider,
		})

		if err != nil {
			return err
		}

		account.MostRecentRole = roleName
		return echoCredentials(args[0], args[0], credentials, outputType, cloudFlag)
	}}

func echoCredentials(id, name string, credentials CloudCredentials, outputType, cloudFlag string) error {
	switch outputType {
	case outputTypeEnvironmentVariable:
		credentials.WriteFormat(os.Stdout, shell, cloudFlag)
		return nil
	case outputTypeAWSCredentialsFile, outputTypeTencentCredentialsFile:
		acc := Account{ID: id, Name: name}
		newCliEntry := NewCloudCliEntry(credentials, &acc)
		cliPath := awsCliPath
		if outputType == outputTypeTencentCredentialsFile {
			cliPath = tencentCliPath
		}
		return SaveCloudCredentialInCLI(cliPath, newCliEntry)
	default:
		return fmt.Errorf("%s is an invalid output type", outputType)
	}
}
