package command

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/urfave/cli/v3"
)

var (
	FlagRoleSessionName = "role-session-name"
	FlagOutputType      = "output"
	FlagShellType       = "shell"
	FlagAWSCLIPath      = "awscli"
)

var switchCmd = &cli.Command{
	Name:  "switch",
	Usage: "Switch from the current AWS account into the one with the given Account ID.",

	Arguments: []cli.Argument{
		&cli.StringArg{Name: "account"},
	},

	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  FlagRoleSessionName,
			Value: "KeyConjurer-AssumeRole",
			Usage: "the name of the role session name that will show up in CloudTrail logs",
		},
		&cli.StringFlag{
			Name:    FlagOutputType,
			Aliases: []string{"o"},
			Value:   outputTypeEnvironmentVariable,
			Usage:   "Format to save new credentials in. Supported outputs: env, awscli, json",
		},
		&cli.StringFlag{
			Name:  FlagShellType,
			Value: shellTypeInfer,
			Usage: "If output type is env, determines which format to output credentials in - by default, the format is inferred based on the execution environment. WSL users may wish to overwrite this to `bash`",
		},
		&cli.StringFlag{
			Name:  FlagAWSCLIPath,
			Value: "~/.aws/",
			Usage: "Path for directory used by the aws-cli tool. Default is \"~/.aws\".",
		},
	},

	UsageText: `Attempt to AssumeRole into the given AWS with the current credentials. You only need to use this if you are a power user or network engineer with access to many accounts.

This is used when a "bastion" account exists which users initially authenticate into and then pivot from that account into other accounts.

This command will fail if you do not have active Cloud credentials.
`,
	Aliases: []string{"switch-account"},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var switchCmd SwitchCommand
		if err := switchCmd.Parse(cmd, cmd.Args().Slice()); err != nil {
			return err
		}

		if err := switchCmd.Validate(); err != nil {
			return err
		}

		return switchCmd.Execute(ctx)
	},
}

type SwitchCommand struct {
	OutputType      string
	ShellType       string
	AWSCLIPath      string
	RoleSessionName string
	AccountID       string
}

func (s *SwitchCommand) Parse(flags flagSet, args []string) error {
	s.OutputType = flags.String(FlagOutputType)
	s.ShellType = flags.String(FlagShellType)
	s.AWSCLIPath = flags.String(FlagAWSCLIPath)
	s.RoleSessionName = flags.String(FlagRoleSessionName)
	if len(args) == 0 {
		return cli.Exit("account-id is required", 1)
	}

	s.AccountID = args[0]
	return nil
}

func (s SwitchCommand) Validate() error {
	if !slices.Contains(permittedOutputTypes, s.OutputType) {
		return ValueError{Value: s.OutputType, ValidValues: permittedOutputTypes}
	}

	if !slices.Contains(permittedShellTypes, s.ShellType) {
		return ValueError{Value: s.ShellType, ValidValues: permittedShellTypes}
	}

	return nil
}

func (s SwitchCommand) Execute(ctx context.Context) error {
	// We could read the environment variable for the assumed role ARN, but it might be expired which isn't very useful to the user.
	creds, err := getAWSCredentials(ctx, s.AccountID, s.RoleSessionName)
	if err != nil {
		// If this failed, either there was a network error or the user is not authorized to assume into this role
		// This can happen if the user is not authenticated using the Bastion instance.
		return err
	}

	switch s.OutputType {
	case outputTypeEnvironmentVariable:
		creds.WriteFormat(os.Stdout, s.ShellType)
		return nil
	case outputTypeAWSCredentialsFile:
		acc := Account{ID: s.AccountID, Name: s.AccountID}
		newCliEntry := NewCloudCliEntry(creds, acc, "")
		return SaveCloudCredentialInCLI(s.AWSCLIPath, newCliEntry)
	default:
		return fmt.Errorf("%s is an invalid output type", s.OutputType)
	}
}

func getAWSCredentials(ctx context.Context, accountID, roleSessionName string) (creds CloudCredentials, err error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return
	}

	c := sts.NewFromConfig(cfg)
	response, err := c.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return
	}

	// We need to modify this to change the last section to be role/GL-SuperAdmin
	id, err := arn.Parse(*response.Arn)
	if err != nil {
		return
	}

	parts := strings.Split(id.Resource, "/")
	arn := arn.ARN{
		AccountID: accountID,
		Partition: "aws",
		Service:   "iam",
		Resource:  fmt.Sprintf("role/%s", parts[1]),
		Region:    id.Region,
	}

	resp, err := c.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         aws.String(arn.String()),
		RoleSessionName: aws.String(roleSessionName),
	})

	if err != nil {
		return
	}

	creds = CloudCredentials{
		Version:         CredentialProcessVersion,
		AccountID:       accountID,
		AccessKeyID:     *resp.Credentials.AccessKeyId,
		SecretAccessKey: *resp.Credentials.SecretAccessKey,
		SessionToken:    *resp.Credentials.SessionToken,
		Expiration:      resp.Credentials.Expiration.Format(time.RFC3339),
	}

	return creds, nil
}
