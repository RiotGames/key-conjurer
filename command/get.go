package command

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/riotgames/key-conjurer/oauth2"
	"github.com/riotgames/key-conjurer/okta"
)

func resolveApplicationInfo(cfg *Config, bypassCache bool, nameOrID string) (*Account, bool) {
	if bypassCache {
		return &Account{ID: nameOrID, Name: nameOrID}, true
	}
	return cfg.FindAccount(nameOrID)
}

type GetCommand struct {
	AccountNameOrID string `arg:""`
	TimeToLive      uint   `placeholder:"hours" help:"The key timeout in hours from 1 to 8." default:"1" name:"ttl"`
	TimeRemaining   uint   `placeholder:"minutes" help:"Request new keys if there are no keys in the environment or the current keys expire within <time-remaining> minutes." default:"5" short:"t"`
	AWSCLIPath      string `help:"Path to the AWS CLI configuration directory." default:"~/.aws/" name:"awscli"`
	ShellType       string `name:"shell" help:"If output type is env, determines which format to output credentials in. WSL users may wish to overwrite this to \"bash\"." default:"infer" enum:"infer,basic,powershell,bash"`
	URLOnly         bool   `help:"Print only the URL to visit rather than a user-friendly message." short:"u"`
	Browser         bool   `help:"Open the browser to the Okta URL. If false, a URL will be printed to the command line instead." default:"true" negatable:"" short:"b"`
	OutputType      string `help:"Format to save new credentials in." default:"env" enum:"env,awscli" short:"o" default:"env" name:"out"`
	Login           bool   `help:"Login to Okta before running the command if the tokens have expired."`
	RoleName        string `help:"The name of the role to assume." short:"r" name:"role"`
	SessionName     string `help:"The name of the role session name that will show up in CloudTrail logs." default:"KeyConjurer-AssumeRole"`
	Region          string `help:"The AWS region to use." env:"AWS_REGION" default:"us-west-2"`
	Cache           bool   `help:"Check the cache for accounts and send the application ID as-is to Okta. This is useful if you have an ID you know is an Okta application ID and it is not stored in your local account cache." default:"true" negatable:"" hidden:""`

	UsageFunc  func() error `kong:"-"`
	PrintErrln func(...any) `kong:"-"`
}

func (g GetCommand) Help() string {
	return `Retrieves temporary cloud API credentials for the specified account.

A role must be specified when using this command through the --role flag. You may list the roles you can assume through the roles command, and the accounts through the accounts command.`
}

func (g GetCommand) printUsage() error {
	return g.UsageFunc()
}

func (g GetCommand) RunContext(ctx context.Context, globals *Globals, cfg *Config) error {
	if HasTokenExpired(cfg.Tokens) {
		if !g.Login {
			return ErrTokensExpiredOrAbsent
		}

		var loginCommand LoginCommand
		if err := loginCommand.RunContext(ctx, globals, cfg); err != nil {
			return err
		}
	}

	var accountID string
	if g.AccountNameOrID != "" {
		accountID = g.AccountNameOrID
	} else if cfg.LastUsedAccount != nil {
		// No account specified. Can we use the most recent one?
		accountID = *cfg.LastUsedAccount
	} else {
		return g.printUsage()
	}

	account, ok := resolveApplicationInfo(cfg, !g.Cache, accountID)
	if !ok {
		return UnknownAccountError(g.AccountNameOrID, "--no-cache")
	}

	if g.RoleName == "" {
		if account.MostRecentRole == "" {
			g.PrintErrln("You must specify the --role flag with this command")
			return nil
		}
		g.RoleName = account.MostRecentRole
	}

	if cfg.TimeRemaining != 0 && g.TimeRemaining == DefaultTimeRemaining {
		g.TimeRemaining = cfg.TimeRemaining
	}

	credentials := LoadAWSCredentialsFromEnvironment()
	if !credentials.ValidUntil(account, time.Duration(g.TimeRemaining)*time.Minute) {
		newCredentials, err := g.fetchNewCredentials(ctx, *account, globals, cfg)
		if err != nil {
			return err
		}
		credentials = *newCredentials
	}

	if account != nil {
		account.MostRecentRole = g.RoleName
	}

	cfg.LastUsedAccount = &accountID
	return echoCredentials(accountID, accountID, credentials, g.OutputType, g.ShellType, g.AWSCLIPath)
}

func (g GetCommand) Run(globals *Globals, cfg *Config) error {
	return g.RunContext(context.Background(), globals, cfg)
}

func (g GetCommand) fetchNewCredentials(ctx context.Context, account Account, globals *Globals, cfg *Config) (*CloudCredentials, error) {
	oauth2Cfg, err := oauth2.DiscoverConfig(ctx, globals.OIDCDomain, globals.ClientID)
	if err != nil {
		return nil, err
	}

	samlResponse, assertionStr, err := okta.ExchangeTokenForAssertion(ctx, oauth2Cfg, cfg.Tokens.AccessToken, cfg.Tokens.IDToken, globals.OIDCDomain, account.ID)
	if err != nil {
		return nil, err
	}

	pair, ok := findRoleInSAML(g.RoleName, samlResponse)
	if !ok {
		return nil, UnknownRoleError(g.RoleName, g.AccountNameOrID)
	}

	if g.TimeToLive == 1 && cfg.TTL != 0 {
		g.TimeToLive = cfg.TTL
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(g.Region))
	if err != nil {
		return nil, err
	}

	stsClient := sts.NewFromConfig(awsCfg)
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
		AccessKeyID:     *resp.Credentials.AccessKeyId,
		Expiration:      resp.Credentials.Expiration.Format(time.RFC3339),
		SecretAccessKey: *resp.Credentials.SecretAccessKey,
		SessionToken:    *resp.Credentials.SessionToken,
	}, nil
}

func echoCredentials(id, name string, credentials CloudCredentials, outputType, shellType, cliPath string) error {
	switch outputType {
	case "env":
		credentials.WriteFormat(os.Stdout, shellType)
		return nil
	case "aws":
		acc := Account{ID: id, Name: name}
		newCliEntry := NewCloudCliEntry(credentials, &acc)
		return SaveCloudCredentialInCLI(cliPath, newCliEntry)
	default:
		return fmt.Errorf("%s is an invalid output type", outputType)
	}
}
