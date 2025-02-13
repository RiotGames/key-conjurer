package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"os"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/charmbracelet/huh"
	"github.com/riotgames/key-conjurer/pkg/oauth2cli"
	"github.com/spf13/cobra"
)

var (
	FlagRegion        = "region"
	FlagRoleName      = "role"
	FlagTimeRemaining = "time-remaining"
	FlagTimeToLive    = "ttl"
	FlagBypassCache   = "bypass-cache"
	FlagLogin         = "login"
	FlagInteractive   = "interactive"

	ErrNoRoles      = errors.New("no roles")
	ErrNoRole       = errors.New("no role")
	ErrNoAccountArg = errors.New("account name or alias is required")
)

var (
	// outputTypeEnvironmentVariable indicates that keyconjurer will dump the credentials to stdout in Bash environment variable format
	outputTypeEnvironmentVariable = "env"
	// outputTypeAWSCredentialsFile indicates that keyconjurer will dump the credentials into the ~/.aws/credentials file.
	outputTypeAWSCredentialsFile = "awscli"
	outputTypeJSON               = "json"
	permittedOutputTypes         = []string{outputTypeAWSCredentialsFile, outputTypeEnvironmentVariable, outputTypeJSON}
	permittedShellTypes          = []string{shellTypePowershell, shellTypeBash, shellTypeBasic, shellTypeInfer}
)

func init() {
	flags := getCmd.Flags()
	flags.String(FlagRegion, "us-west-2", "The AWS region to use")
	flags.Uint(FlagTimeToLive, 1, "The key timeout in hours from 1 to 8.")
	flags.UintP(FlagTimeRemaining, "t", DefaultTimeRemaining, "Request new keys if there are no keys in the environment or the current keys expire within <time-remaining> minutes. Defaults to 60.")
	flags.StringP(FlagRoleName, "r", "", "The name of the role to assume.")
	flags.String(FlagRoleSessionName, "KeyConjurer-AssumeRole", "the name of the role session name that will show up in CloudTrail logs")
	flags.StringP(FlagOutputType, "o", outputTypeEnvironmentVariable, "Format to save new credentials in. Supported outputs: env, awscli, json")
	flags.String(FlagShellType, shellTypeInfer, "If output type is env, determines which format to output credentials in - by default, the format is inferred based on the execution environment. WSL users may wish to overwrite this to `bash`")
	flags.Bool(FlagBypassCache, false, "Do not check the cache for accounts and send the application ID as-is to Okta. This is useful if you have an ID you know is an Okta application ID and it is not stored in your local account cache.")
	flags.Bool(FlagLogin, false, "Login to Okta before running the command")
	flags.String(FlagAWSCLIPath, "~/.aws/", "Path for directory used by the aws CLI")
	flags.BoolP(FlagURLOnly, "u", false, "Print only the URL to visit rather than a user-friendly message")
	flags.BoolP(FlagNoBrowser, "b", false, "Do not open a browser window, printing the URL instead")
	flags.Bool(FlagInteractive, false, "Use interactive prompts to supply information not otherwise supplied with flags")
}

func resolveApplicationInfo(cfg *Config, bypassCache bool, nameOrID string) (*Account, bool) {
	if bypassCache {
		return &Account{ID: nameOrID, Name: nameOrID}, true
	}
	return cfg.FindAccount(nameOrID)
}

type GetCommand struct {
	AccountIDOrName                                                           string
	TimeToLive                                                                uint
	TimeRemaining                                                             uint
	OutputType, ShellType, RoleName, AWSCLIPath, OIDCDomain, ClientID, Region string
	Login, URLOnly, NoBrowser, BypassCache, MachineOutput, Interactive        bool

	UsageFunc  func() error
	PrintErrln func(...any)
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
	g.UsageFunc = cmd.Usage
	g.PrintErrln = cmd.PrintErrln
	g.Interactive, _ = flags.GetBool(FlagInteractive)
	g.MachineOutput = ShouldUseMachineOutput(flags) || g.URLOnly
	if len(args) > 0 {
		g.AccountIDOrName = args[0]
	} else if g.Interactive {
		// We can resolve this at execution time with an interactive prompt.
		g.AccountIDOrName = ""
	} else {
		return ErrNoAccountArg
	}
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

func (g GetCommand) Execute(ctx context.Context, config *Config) error {
	var accountID string
	if g.AccountIDOrName != "" {
		accountID = g.AccountIDOrName
	} else if g.Interactive {
		acc, err := accountsInteractivePrompt(config.EnumerateAccounts(), nil)
		if err != nil {
			return err
		}
		accountID = acc.ID
	} else if config.LastUsedAccount != nil {
		// No account specified. Can we use the most recent one?
		accountID = *config.LastUsedAccount
	} else {
		return g.printUsage()
	}

	account, ok := resolveApplicationInfo(config, g.BypassCache, accountID)
	if !ok {
		return UnknownAccountError(g.AccountIDOrName, FlagBypassCache)
	}

	if config.TimeRemaining != 0 && g.TimeRemaining == DefaultTimeRemaining {
		g.TimeRemaining = config.TimeRemaining
	}

	credentials := LoadAWSCredentialsFromEnvironment()
	if !credentials.ValidUntil(account, time.Duration(g.TimeRemaining)*time.Minute) {
		newCredentials, err := g.fetchNewCredentials(ctx, account, config)
		if errors.Is(err, ErrTokensExpiredOrAbsent) && g.Login {
			loginCommand := LoginCommand{
				OIDCDomain:    g.OIDCDomain,
				ClientID:      g.ClientID,
				MachineOutput: g.MachineOutput,
				NoBrowser:     g.NoBrowser,
			}
			err = loginCommand.Execute(ctx, config)
			if err != nil {
				return err
			}
			newCredentials, err = g.fetchNewCredentials(ctx, account, config)
		}

		if errors.Is(err, ErrNoRoles) {
			g.PrintErrln("You don't have access to any roles on this account.")
			return nil
		}

		if errors.Is(err, ErrNoRole) {
			g.PrintErrln("You must specify a role with --role or using the interactive prompt.")
			return nil
		}

		if err != nil {
			return err
		}

		credentials = *newCredentials
	}

	config.LastUsedAccount = &accountID
	return echoCredentials(accountID, accountID, credentials, g.OutputType, g.ShellType, g.AWSCLIPath)
}

// fetchNewCredentials fetches new credentials for the given account.
//
// 'account' will have its MostRecentRole field updated to the role used if this call is successful.
func (g GetCommand) fetchNewCredentials(ctx context.Context, account *Account, cfg *Config) (*CloudCredentials, error) {
	samlResponse, assertionStr, err := oauth2cli.DiscoverConfigAndExchangeTokenForAssertion(ctx, &keychainTokenSource{}, g.OIDCDomain, g.ClientID, account.ID)
	if err != nil {
		return nil, err
	}

	roles := listRoles(samlResponse)
	if len(roles) == 0 {
		return nil, ErrNoRoles
	}

	if g.RoleName == "" {
		if account.MostRecentRole == "" || g.Interactive {
			g.RoleName, err = rolesInteractivePrompt(listRoles(samlResponse), account.MostRecentRole)
			if err != nil {
				return nil, ErrNoRole
			}
		} else {
			g.RoleName = account.MostRecentRole
		}
	}

	pair, ok := findRoleInSAML(g.RoleName, samlResponse)
	if !ok {
		return nil, UnknownRoleError(g.RoleName, g.AccountIDOrName)
	}
	account.MostRecentRole = g.RoleName

	if g.TimeToLive == 1 && cfg.TTL != 0 {
		g.TimeToLive = cfg.TTL
	}

	stsClient := sts.New(sts.Options{Region: g.Region})
	timeoutInSeconds := int32(3600 * g.TimeToLive)
	resp, err := stsClient.AssumeRoleWithSAML(ctx, &sts.AssumeRoleWithSAMLInput{
		DurationSeconds: aws.Int32(timeoutInSeconds),
		PrincipalArn:    aws.String(pair.ProviderARN),
		RoleArn:         aws.String(pair.RoleARN),
		SAMLAssertion:   aws.String(assertionStr),
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
		AccountID:       account.ID,
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

		if err := getCmd.Validate(); err != nil {
			return err
		}

		return getCmd.Execute(cmd.Context(), ConfigFromCommand(cmd))
	},
}

func echoCredentials(id, name string, credentials CloudCredentials, outputType, shellType, cliPath string) error {
	switch outputType {
	case outputTypeJSON:
		buf, err := json.Marshal(credentials)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, string(buf))
		return nil
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

func accountsInteractivePrompt(accounts iter.Seq[Account], selected *Account) (Account, error) {
	var opts []huh.Option[Account]
	for account := range accounts {
		opts = append(opts, huh.Option[Account]{
			Key:   account.Alias,
			Value: account,
		})
	}

	ctrl := huh.NewSelect[Account]().
		Options(opts...).
		Filtering(true).
		Title("account").
		Description("Choose an account using your arrow keys or by typing the account name and pressing return to confirm your selection.")

	if selected != nil {
		ctrl = ctrl.Value(selected)
	}

	err := huh.Run(ctrl)
	if err != nil {
		return Account{}, err
	}
	return ctrl.GetValue().(Account), nil
}

func rolesInteractivePrompt(roles []string, mostRecent string) (string, error) {
	opts := huh.NewOptions(roles...)
	ctrl := huh.NewSelect[string]().
		Options(opts...).
		Filtering(true).
		Title("role").
		Value(&mostRecent).
		Description("Choose a role using your arrow keys or by typing the role name and press the return key to confirm.")

	err := huh.Run(ctrl)
	if err != nil {
		return "", err
	}
	return ctrl.GetValue().(string), nil
}
