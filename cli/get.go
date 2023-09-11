package main

import (
	"fmt"
	"os"
	"time"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/riotgames/key-conjurer/internal"
	"github.com/spf13/cobra"
)

var (
	FlagRegion = "region"
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
	getCmd.Flags().String(FlagRegion, "us-west-2", "The AWS region to use")
	getCmd.Flags().UintVar(&ttl, "ttl", 1, "The key timeout in hours from 1 to 8.")
	getCmd.Flags().UintVarP(&timeRemaining, "time-remaining", "t", DefaultTimeRemaining, "Request new keys if there are no keys in the environment or the current keys expire within <time-remaining> minutes. Defaults to 60.")
	getCmd.Flags().StringVarP(&outputType, "out", "o", outputTypeEnvironmentVariable, "Format to save new credentials in. Supported outputs: env, awscli,tencentcli")
	getCmd.Flags().StringVarP(&shell, "shell", "", shellTypeInfer, "If output type is env, determines which format to output credentials in - by default, the format is inferred based on the execution environment. WSL users may wish to overwrite this to `bash`")
	getCmd.Flags().StringVarP(&awsCliPath, "awscli", "", "~/.aws/", "Path for directory used by the aws-cli tool. Default is \"~/.aws\".")
	getCmd.Flags().StringVarP(&tencentCliPath, "tencentcli", "", "~/.tencent/", "Path for directory used by the tencent-cli tool. Default is \"~/.tencent\".")
	getCmd.Flags().StringVar(&roleName, "role", "", "The name of the role to assume.")
	getCmd.Flags().StringVarP(&cloudFlag, "cloud", "", "aws", "Choose a cloud vendor. Default is aws. Can choose aws or tencent")
}

func isMemberOfSlice(slice []string, val string) bool {
	for _, member := range slice {
		if member == val {
			return true
		}
	}

	return false
}

var getCmd = &cobra.Command{
	Use:   "get <accountName/alias>",
	Short: "Retrieves temporary Cloud(AWS|Tencent) API credentials.",
	Long: `Retrieves temporary Cloud(AWS|Tencent) API credentials for the specified account.  It sends a push request to the first Duo device it finds associated with your account.

	A role must be specified when using this command through the --role flag. You may list the roles you can assume through the roles command.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config := ConfigFromContext(cmd.Context())
		ctx := cmd.Context()
		if HasTokenExpired(config.Tokens) {
			cmd.PrintErrln("Your session has expired. Please login again.")
			return nil
		}
		client := NewHTTPClient()

		if !isMemberOfSlice(permittedOutputTypes, outputType) {
			return invalidValueError(outputType, permittedOutputTypes)
		}

		if !isMemberOfSlice(permittedShellTypes, shell) {
			return invalidValueError(shell, permittedShellTypes)
		}

		// make sure we enforce limit
		if ttl > 8 {
			ttl = 8
		}

		var applicationID = args[0]
		account, ok := config.FindAccount(applicationID)
		if ok {
			applicationID = account.ID
		}

		if account.MostRecentRole != "" && roleName == "" {
			roleName = account.MostRecentRole
		}

		if config.TimeRemaining != 0 && timeRemaining == DefaultTimeRemaining {
			timeRemaining = config.TimeRemaining
		}

		if roleName == "" {
			cmd.PrintErrln("You must specify the --role flag with this command")
			return nil
		}

		var credentials CloudCredentials
		credentials.LoadFromEnv(cloudFlag)
		if credentials.ValidUntil(*account, cloudFlag, time.Duration(timeRemaining)*time.Minute) {
			return echoCredentials(args[0], args[0], credentials, outputType, cloudFlag)
		}

		oidcDomain, _ := cmd.Flags().GetString(FlagOIDCDomain)
		clientID, _ := cmd.Flags().GetString(FlagClientID)
		oauthCfg, _, err := DiscoverOAuth2Config(cmd.Context(), client, oidcDomain, clientID)
		if err != nil {
			cmd.PrintErrf("could not discover oauth2  config: %s\n", err)
			return nil
		}

		tok, err := ExchangeAccessTokenForWebSSOToken(cmd.Context(), client, oauthCfg, config.Tokens, applicationID)
		if err != nil {
			cmd.PrintErrf("error exchanging token: %s\n", err)
			return nil
		}

		assertionBytes, err := ExchangeWebSSOTokenForSAMLAssertion(cmd.Context(), client, oidcDomain, tok)
		if err != nil {
			cmd.PrintErrf("failed to fetch SAML assertion: %s\n", err)
			return nil
		}

		assertionStr := string(assertionBytes)
		samlResponse, err := saml.ParseEncodedResponse(assertionStr)
		if err != nil {
			cmd.PrintErrf("could not parse assertion: %s\n", err)
			return nil
		}

		pair, _, ok := internal.FindRole(roleName, samlResponse)
		if !ok {
			cmd.PrintErrf("you do not have access to the role %s on application %s\n", roleName, args[0])
			return nil
		}

		if cloudFlag == cloudAws {
			region, _ := cmd.Flags().GetString(FlagRegion)
			session, _ := session.NewSession(&aws.Config{Region: aws.String(region)})
			stsClient := sts.New(session)
			timeoutInSeconds := int64(3600 * ttl)
			resp, err := stsClient.AssumeRoleWithSAMLWithContext(ctx, &sts.AssumeRoleWithSAMLInput{
				DurationSeconds: &timeoutInSeconds,
				PrincipalArn:    &pair.ProviderARN,
				RoleArn:         &pair.RoleARN,
				SAMLAssertion:   &assertionStr,
			})

			if err != nil {
				cmd.PrintErrf("failed to exchange credentials: %s", err)
				return nil
			}

			credentials = CloudCredentials{
				AccessKeyID:     *resp.Credentials.AccessKeyId,
				Expiration:      resp.Credentials.Expiration.Format(time.RFC3339),
				SecretAccessKey: *resp.Credentials.SecretAccessKey,
				SessionToken:    *resp.Credentials.SessionToken,
			}
		}

		if ttl == 1 && config.TTL != 0 {
			ttl = config.TTL
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
