package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/riotgames/key-conjurer/pkg/oauth2cli"
	"github.com/urfave/cli/v3"
)

var (
	FlagRegion        = "region"
	FlagRoleName      = "role"
	FlagTimeRemaining = "time-remaining"
	FlagTimeToLive    = "ttl"
	FlagBypassCache   = "bypass-cache"
	FlagLogin         = "login"
	FlagProfileName   = "profile"
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

var getCmd = &cli.Command{
	Name:      "get",
	Usage:     "Retrieves temporary cloud API credentials for the specified account.",
	UsageText: `A role must be specified when using this command through the --role flag. You may list the roles you can assume through the roles command.`,
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var getCmd GetCommand
		if err := getCmd.Parse(cmd, cmd.Args().Slice()); err != nil {
			return err
		}

		if err := getCmd.Validate(); err != nil {
			return err
		}

		return getCmd.Execute(ctx, ConfigFromContext(ctx))
	},

	Arguments: []cli.Argument{
		&cli.StringArg{Name: "account"},
	},

	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "region",
			Usage: "The AWS region to use",
			Value: "us-west-2",
		},
		&cli.UintFlag{
			Name:  "ttl",
			Usage: "The key timeout in hours from 1 to 8",
			Value: 1,
		},
		&cli.UintFlag{
			Name:    "time-remaining",
			Usage:   "Request new keys if there are no keys in the environment or the current keys expire within <time-remaining> minutes. Defaults to 60",
			Aliases: []string{"t"},
			Value:   DefaultTimeRemaining,
		},
		&cli.StringFlag{
			Name:     "role-name",
			Usage:    "The name of the role to assume",
			Aliases:  []string{"r"},
			Required: true,
		},
		&cli.StringFlag{
			Name:  "role-session-name",
			Usage: "the name of the role session name that will show up in CloudTrail logs",
			Value: "KeyConjurer-AssumeRole",
		},
		&cli.StringFlag{
			Name:    "output-type",
			Usage:   "Format to save new credentials in. Supported outputs: env, awscli, json",
			Value:   outputTypeEnvironmentVariable,
			Aliases: []string{"o"},
		},
		&cli.StringFlag{
			Name:  "shell-type",
			Usage: "If output type is env, determines which format to output credentials in - by default, the format is inferred based on the execution environment. WSL users may wish to overwrite this to `bash`",
			Value: shellTypeInfer,
		},
		&cli.BoolFlag{
			Name:  "bypass-cache",
			Usage: "Do not check the cache for accounts and send the application ID as-is to Okta. This is useful if you have an ID you know is an Okta application ID and it is not stored in your local account cache.",
		},
		&cli.BoolFlag{
			Name:  "login",
			Usage: "Login to Okta before running the command",
		},
		&cli.StringFlag{
			Name:  "awscli-path",
			Usage: "Path for directory used by the aws CLI",
			Value: "~/.aws/",
		},
		&cli.BoolFlag{
			Name:    "url-only",
			Aliases: []string{"u"},
			Usage:   "Print only the URL to visit rather than a user-friendly message",
		},
		&cli.BoolFlag{
			Name:    "no-browser",
			Aliases: []string{"b"},
			Usage:   "Do not open a browser window, printing the URL instead",
		},
		&cli.StringFlag{
			Name:    "profile-name",
			Aliases: []string{"p"},
			Usage:   "The name of the awscli profile to save credentials. Only used if output type is awscli. Defaults to the account name.",
		},
	},
}

func resolveApplicationInfo(cfg *Config, bypassCache bool, nameOrID string) (Account, bool) {
	if bypassCache {
		return Account{ID: nameOrID, Name: nameOrID}, true
	}
	return cfg.FindAccount(nameOrID)
}

type GetCommand struct {
	AccountIDOrName                                                                        string
	TimeToLive                                                                             uint
	TimeRemaining                                                                          uint
	OutputType, ShellType, RoleName, AWSCLIPath, OIDCDomain, ClientID, Region, ProfileName string
	Login, URLOnly, NoBrowser, BypassCache, MachineOutput                                  bool
}

func (g *GetCommand) Parse(flags flagSet, args []string) error {
	g.OIDCDomain = flags.String(FlagOIDCDomain)
	g.ClientID = flags.String(FlagClientID)
	g.TimeToLive = flags.Uint(FlagTimeToLive)
	g.TimeRemaining = flags.Uint(FlagTimeRemaining)
	g.OutputType = flags.String(FlagOutputType)
	g.ShellType = flags.String(FlagShellType)
	g.RoleName = flags.String(FlagRoleName)
	g.AWSCLIPath = flags.String(FlagAWSCLIPath)
	g.Login = flags.Bool(FlagLogin)
	g.URLOnly = flags.Bool(FlagURLOnly)
	g.NoBrowser = flags.Bool(FlagNoBrowser)
	g.BypassCache = flags.Bool(FlagBypassCache)
	g.ProfileName = flags.String(FlagProfileName)
	g.Region = flags.String(FlagRegion)
	g.MachineOutput = ShouldUseMachineOutput(flags) || g.URLOnly
	if len(args) == 0 {
		return cli.Exit("account name or alias is required", 1)
	}
	g.AccountIDOrName = args[0]
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

func (g GetCommand) Execute(ctx context.Context, config *Config) error {
	var accountID string
	if g.AccountIDOrName != "" {
		accountID = g.AccountIDOrName
	} else if config.LastUsedAccount != nil {
		// No account specified. Can we use the most recent one?
		accountID = *config.LastUsedAccount
	} else {
		return errors.New("account name or alias is required")
	}

	account, ok := resolveApplicationInfo(config, g.BypassCache, accountID)
	if !ok {
		return UnknownAccountError(g.AccountIDOrName, FlagBypassCache)
	}

	if g.RoleName == "" {
		if account.MostRecentRole == "" {
			return errors.New("You must specify the --role flag with this command")
		}
		g.RoleName = account.MostRecentRole
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

		if err != nil {
			return err
		}

		credentials = *newCredentials
	}

	account.MostRecentRole = g.RoleName

	config.LastUsedAccount = &accountID
	return echoCredentials(account, g.ProfileName, credentials, g.OutputType, g.ShellType, g.AWSCLIPath)
}

func (g GetCommand) fetchNewCredentials(ctx context.Context, account Account, cfg *Config) (*CloudCredentials, error) {
	samlResponse, assertionStr, err := oauth2cli.DiscoverConfigAndExchangeTokenForAssertion(ctx, &keychainTokenSource{}, g.OIDCDomain, g.ClientID, account.ID)
	if err != nil {
		return nil, err
	}

	pair, ok := findRoleInSAML(g.RoleName, samlResponse)
	if !ok {
		return nil, UnknownRoleError(g.RoleName, g.AccountIDOrName)
	}

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

func echoCredentials(account Account, profileName string, credentials CloudCredentials, outputType, shellType, cliPath string) error {
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
		newCliEntry := NewCloudCliEntry(credentials, account, profileName)
		return SaveCloudCredentialInCLI(cliPath, newCliEntry)
	default:
		return fmt.Errorf("%s is an invalid output type", outputType)
	}
}
