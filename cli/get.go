package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/spf13/cobra"
)

var (
	FlagRegion        = "region"
	FlagRoleName      = "role"
	FlagTimeRemaining = "time-remaining"
	FlagTimeToLive    = "ttl"
	FlagBypassCache   = "bypass-cache"
	FlagLogin         = "login"
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
	getCmd.Flags().Uint(FlagTimeToLive, 1, "The key timeout in hours from 1 to 8.")
	getCmd.Flags().UintP(FlagTimeRemaining, "t", DefaultTimeRemaining, "Request new keys if there are no keys in the environment or the current keys expire within <time-remaining> minutes. Defaults to 60.")
	getCmd.Flags().StringP(FlagRoleName, "r", "", "The name of the role to assume.")
	getCmd.Flags().String(FlagRoleSessionName, "KeyConjurer-AssumeRole", "the name of the role session name that will show up in CloudTrail logs")
	getCmd.Flags().StringP(FlagOutputType, "o", outputTypeEnvironmentVariable, "Format to save new credentials in. Supported outputs: env, awscli,tencentcli")
	getCmd.Flags().String(FlagShellType, shellTypeInfer, "If output type is env, determines which format to output credentials in - by default, the format is inferred based on the execution environment. WSL users may wish to overwrite this to `bash`")
	getCmd.Flags().String(FlagAWSCLIPath, "~/.aws/", "Path for directory used by the aws-cli tool. Default is \"~/.aws\".")
	getCmd.Flags().String(FlagTencentCLIPath, "~/.tencent/", "Path for directory used by the tencent-cli tool. Default is \"~/.tencent\".")
	getCmd.Flags().String(FlagCloudType, "aws", "Choose a cloud vendor. Default is aws. Can choose aws or tencent")
	getCmd.Flags().Bool(FlagBypassCache, false, "Do not check the cache for accounts and send the application ID as-is to Okta. This is useful if you have an ID you know is an Okta application ID and it is not stored in your local account cache.")
	getCmd.Flags().Bool(FlagLogin, false, "Login to Okta before running the command")
}

func isMemberOfSlice(slice []string, val string) bool {
	for _, member := range slice {
		if member == val {
			return true
		}
	}

	return false
}

func resolveApplicationInfo(cfg *Config, bypassCache bool, nameOrID string) (*Account, bool) {
	if bypassCache {
		return &Account{ID: nameOrID, Name: nameOrID}, true
	}
	return cfg.FindAccount(nameOrID)
}

var getCmd = &cobra.Command{
	Use:   "get <accountName/alias>",
	Short: "Retrieves temporary cloud API credentials.",
	Long: `Retrieves temporary cloud API credentials for the specified account.  It sends a push request to the first Duo device it finds associated with your account.

A role must be specified when using this command through the --role flag. You may list the roles you can assume through the roles command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := ConfigFromCommand(cmd)
		ctx := cmd.Context()
		oidcDomain, _ := cmd.Flags().GetString(FlagOIDCDomain)
		clientID, _ := cmd.Flags().GetString(FlagClientID)
		if HasTokenExpired(config.Tokens) {
			if ok, _ := cmd.Flags().GetBool(FlagLogin); ok {
				token, err := Login(ctx, oidcDomain, clientID, LoginOutputModeBrowser{})
				if err != nil {
					return err
				}
				if err := config.SaveOAuthToken(token); err != nil {
					return err
				}

			} else {
				return ErrTokensExpiredOrAbsent
			}
			return nil
		}

		ttl, _ := cmd.Flags().GetUint(FlagTimeToLive)
		timeRemaining, _ := cmd.Flags().GetUint(FlagTimeRemaining)
		outputType, _ := cmd.Flags().GetString(FlagOutputType)
		shellType, _ := cmd.Flags().GetString(FlagShellType)
		roleName, _ := cmd.Flags().GetString(FlagRoleName)
		cloudType, _ := cmd.Flags().GetString(FlagCloudType)
		awsCliPath, _ := cmd.Flags().GetString(FlagAWSCLIPath)
		tencentCliPath, _ := cmd.Flags().GetString(FlagTencentCLIPath)

		if !isMemberOfSlice(permittedOutputTypes, outputType) {
			return ValueError{Value: outputType, ValidValues: permittedOutputTypes}
		}

		if !isMemberOfSlice(permittedShellTypes, shellType) {
			return ValueError{Value: shellType, ValidValues: permittedShellTypes}
		}

		var accountID string
		if len(args) > 0 {
			accountID = args[0]
		} else if config.LastUsedAccount != nil {
			// No account specified. Can we use the most recent one?
			accountID = *config.LastUsedAccount
		} else {
			return cmd.Usage()
		}

		bypassCache, _ := cmd.Flags().GetBool(FlagBypassCache)
		account, ok := resolveApplicationInfo(config, bypassCache, accountID)
		if !ok {
			return UnknownAccountError(args[0], FlagBypassCache)
		}

		if roleName == "" {
			if account.MostRecentRole == "" {
				cmd.PrintErrln("You must specify the --role flag with this command")
				return nil
			}
			roleName = account.MostRecentRole
		}

		if config.TimeRemaining != 0 && timeRemaining == DefaultTimeRemaining {
			timeRemaining = config.TimeRemaining
		}

		var credentials CloudCredentials
		if cloudType == cloudAws {
			credentials = LoadAWSCredentialsFromEnvironment()
		} else if cloudType == cloudTencent {
			credentials = LoadTencentCredentialsFromEnvironment()
		}

		if credentials.ValidUntil(account, time.Duration(timeRemaining)*time.Minute) {
			return echoCredentials(accountID, accountID, credentials, outputType, shellType, awsCliPath, tencentCliPath)
		}

		samlResponse, assertionStr, err := DiscoverConfigAndExchangeTokenForAssertion(cmd.Context(), NewHTTPClient(), config.Tokens, oidcDomain, clientID, account.ID)
		if err != nil {
			return err
		}

		pair, ok := FindRoleInSAML(roleName, samlResponse)
		if !ok {
			return UnknownRoleError(roleName, args[0])
		}

		if ttl == 1 && config.TTL != 0 {
			ttl = config.TTL
		}

		if cloudType == cloudAws {
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

			if err, ok := tryParseTimeToLiveError(err); ok {
				return err
			}

			if err != nil {
				return AWSError{
					InnerError: err,
					Message:    "failed to exchange credentials",
				}
			}

			credentials = CloudCredentials{
				AccessKeyID:     *resp.Credentials.AccessKeyId,
				Expiration:      resp.Credentials.Expiration.Format(time.RFC3339),
				SecretAccessKey: *resp.Credentials.SecretAccessKey,
				SessionToken:    *resp.Credentials.SessionToken,
				credentialsType: cloudType,
			}
		} else {
			panic("not yet implemented")
		}

		if account != nil {
			account.MostRecentRole = roleName
		}
		config.LastUsedAccount = &accountID

		return echoCredentials(accountID, accountID, credentials, outputType, shellType, awsCliPath, tencentCliPath)
	}}

func echoCredentials(id, name string, credentials CloudCredentials, outputType, shellType, awsCliPath, tencentCliPath string) error {
	switch outputType {
	case outputTypeEnvironmentVariable:
		credentials.WriteFormat(os.Stdout, shellType)
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
