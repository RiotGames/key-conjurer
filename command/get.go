package command

import (
	"context"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/riotgames/key-conjurer/oauth2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	permittedOutputTypes         = []string{outputTypeAWSCredentialsFile, outputTypeEnvironmentVariable}
	permittedShellTypes          = []string{shellTypePowershell, shellTypeBash, shellTypeBasic, shellTypeInfer}
)

func init() {
	getCmd.Flags().String(FlagRegion, "us-west-2", "The AWS region to use")
	getCmd.Flags().Uint(FlagTimeToLive, 1, "The key timeout in hours from 1 to 8.")
	getCmd.Flags().UintP(FlagTimeRemaining, "t", DefaultTimeRemaining, "Request new keys if there are no keys in the environment or the current keys expire within <time-remaining> minutes. Defaults to 60.")
	getCmd.Flags().StringP(FlagRoleName, "r", "", "The name of the role to assume.")
	getCmd.Flags().String(FlagRoleSessionName, "KeyConjurer-AssumeRole", "the name of the role session name that will show up in CloudTrail logs")
	getCmd.Flags().StringP(FlagOutputType, "o", outputTypeEnvironmentVariable, "Format to save new credentials in. Supported outputs: env, awscli")
	getCmd.Flags().String(FlagShellType, shellTypeInfer, "If output type is env, determines which format to output credentials in - by default, the format is inferred based on the execution environment. WSL users may wish to overwrite this to `bash`")
	getCmd.Flags().Bool(FlagBypassCache, false, "Do not check the cache for accounts and send the application ID as-is to Okta. This is useful if you have an ID you know is an Okta application ID and it is not stored in your local account cache.")
	getCmd.Flags().Bool(FlagLogin, false, "Login to Okta before running the command")
	getCmd.Flags().String(FlagAWSCLIPath, "~/.aws/", "Path for directory used by the aws CLI")
	getCmd.Flags().BoolP(FlagURLOnly, "u", false, "Print only the URL to visit rather than a user-friendly message")
	getCmd.Flags().BoolP(FlagNoBrowser, "b", false, "Do not open a browser window, printing the URL instead")
}

func resolveApplicationInfo(cfg *Config, bypassCache bool, nameOrID string) (*Account, bool) {
	if bypassCache {
		return &Account{ID: nameOrID, Name: nameOrID}, true
	}
	return cfg.FindAccount(nameOrID)
}

type GetCommand struct {
	Config *Config

	Args                                                                      []string
	TimeToLive                                                                uint
	TimeRemaining                                                             uint
	OutputType, ShellType, RoleName, AWSCLIPath, OIDCDomain, ClientID, Region string
	Login, URLOnly, NoBrowser, BypassCache                                    bool

	UsageFunc  func() error
	PrintErrln func(...any)

	Flags   *pflag.FlagSet
	Command *cobra.Command
}

func (g *GetCommand) Parse(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	g.OIDCDomain, _ = flags.GetString(FlagOIDCDomain)
	g.ClientID, _ = flags.GetString(FlagClientID)
	g.TimeToLive, _ = flags.GetUint(FlagTimeToLive)
	g.TimeRemaining, _ = flags.GetUint(FlagTimeRemaining)
	g.OutputType, _ = flags.GetString(FlagOutputType)
	g.ShellType, _ = flags.GetString(FlagShellType)
	g.RoleName, _ = flags.GetString(FlagRoleName)
	g.AWSCLIPath, _ = flags.GetString(FlagAWSCLIPath)
	g.Login, _ = flags.GetBool(FlagLogin)
	g.URLOnly, _ = flags.GetBool(FlagURLOnly)
	g.NoBrowser, _ = flags.GetBool(FlagNoBrowser)
	g.BypassCache, _ = flags.GetBool(FlagBypassCache)
	g.Region, _ = flags.GetString(FlagRegion)
	g.Flags = flags
	g.Args = args
	g.UsageFunc = cmd.Usage
	g.PrintErrln = cmd.PrintErrln
	return nil
}

func (g GetCommand) Validate() error {
	if !slices.Contains(permittedOutputTypes, g.OutputType) {
		return ValueError{Value: g.OutputType, ValidValues: permittedOutputTypes}
	}

	if !slices.Contains(permittedShellTypes, g.ShellType) {
		return ValueError{Value: g.ShellType, ValidValues: permittedShellTypes}
	}
	return nil
}

func (g GetCommand) printUsage() error {
	return g.UsageFunc()
}

func (g GetCommand) Execute(ctx context.Context) error {
	if HasTokenExpired(g.Config.Tokens) {
		if g.Login {
			login := LoginCommand{
				Config:        g.Config,
				OIDCDomain:    g.OIDCDomain,
				ClientID:      g.ClientID,
				MachineOutput: ShouldUseMachineOutput(g.Flags) || g.URLOnly,
				NoBrowser:     g.NoBrowser,
			}

			if err := login.Execute(ctx); err != nil {
				return err
			}
		} else {
			return ErrTokensExpiredOrAbsent
		}
		return nil
	}

	var accountID string
	if len(g.Args) > 0 {
		accountID = g.Args[0]
	} else if g.Config.LastUsedAccount != nil {
		// No account specified. Can we use the most recent one?
		accountID = *g.Config.LastUsedAccount
	} else {
		return g.printUsage()
	}

	account, ok := resolveApplicationInfo(g.Config, g.BypassCache, accountID)
	if !ok {
		return UnknownAccountError(g.Args[0], FlagBypassCache)
	}

	if g.RoleName == "" {
		if account.MostRecentRole == "" {
			g.PrintErrln("You must specify the --role flag with this command")
			return nil
		}
		g.RoleName = account.MostRecentRole
	}

	if g.Config.TimeRemaining != 0 && g.TimeRemaining == DefaultTimeRemaining {
		g.TimeRemaining = g.Config.TimeRemaining
	}

	credentials := LoadAWSCredentialsFromEnvironment()
	if !credentials.ValidUntil(account, time.Duration(g.TimeRemaining)*time.Minute) {
		newCredentials, err := g.fetchNewCredentials(ctx, *account)
		if err != nil {
			return err
		}
		credentials = *newCredentials
	}

	if account != nil {
		account.MostRecentRole = g.RoleName
	}

	g.Config.LastUsedAccount = &accountID
	return echoCredentials(accountID, accountID, credentials, g.OutputType, g.ShellType, g.AWSCLIPath)
}

func (g GetCommand) fetchNewCredentials(ctx context.Context, account Account) (*CloudCredentials, error) {
	samlResponse, assertionStr, err := oauth2.DiscoverConfigAndExchangeTokenForAssertion(ctx, g.Config.Tokens.AccessToken, g.Config.Tokens.IDToken, g.OIDCDomain, g.ClientID, account.ID)
	if err != nil {
		return nil, err
	}

	pair, ok := findRoleInSAML(g.RoleName, samlResponse)
	if !ok {
		return nil, UnknownRoleError(g.RoleName, g.Args[0])
	}

	if g.TimeToLive == 1 && g.Config.TTL != 0 {
		g.TimeToLive = g.Config.TTL
	}

	session, _ := session.NewSession(&aws.Config{Region: aws.String(g.Region)})
	stsClient := sts.New(session)
	timeoutInSeconds := int64(3600 * g.TimeToLive)
	resp, err := stsClient.AssumeRoleWithSAMLWithContext(ctx, &sts.AssumeRoleWithSAMLInput{
		DurationSeconds: &timeoutInSeconds,
		PrincipalArn:    &pair.ProviderARN,
		RoleArn:         &pair.RoleARN,
		SAMLAssertion:   &assertionStr,
	})

	if err, ok := tryParseTimeToLiveError(err); ok {
		return nil, err
	}

	if err != nil {
		return nil, AWSError{
			InnerError: err,
			Message:    "failed to exchange credentials",
		}
	}

	return &CloudCredentials{
		AccessKeyID:     *resp.Credentials.AccessKeyId,
		Expiration:      resp.Credentials.Expiration.Format(time.RFC3339),
		SecretAccessKey: *resp.Credentials.SecretAccessKey,
		SessionToken:    *resp.Credentials.SessionToken,
	}, nil
}

var getCmd = &cobra.Command{
	Use:   "get <accountName/alias>",
	Short: "Retrieves temporary cloud API credentials.",
	Long: `Retrieves temporary cloud API credentials for the specified account.  It sends a push request to the first Duo device it finds associated with your account.

A role must be specified when using this command through the --role flag. You may list the roles you can assume through the roles command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var getCmd GetCommand
		if err := getCmd.Parse(cmd, args); err != nil {
			return err
		}

		return getCmd.Execute(cmd.Context())
	},
}

func echoCredentials(id, name string, credentials CloudCredentials, outputType, shellType, cliPath string) error {
	switch outputType {
	case outputTypeEnvironmentVariable:
		credentials.WriteFormat(os.Stdout, shellType)
		return nil
	case outputTypeAWSCredentialsFile:
		acc := Account{ID: id, Name: name}
		newCliEntry := NewCloudCliEntry(credentials, &acc)
		return SaveCloudCredentialInCLI(cliPath, newCliEntry)
	default:
		return fmt.Errorf("%s is an invalid output type", outputType)
	}
}
