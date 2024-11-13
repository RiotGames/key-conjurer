package command

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type SwitchCommand struct {
	OutputType  string `help:"Format to save new credentials in." default:"env" enum:"env,awscli" short:"o" default:"env" name:"out"`
	AWSCLIPath  string `help:"Path to the AWS CLI configuration directory." default:"~/.aws/" name:"awscli"`
	ShellType   string `name:"shell" help:"If output type is env, determines which format to output credentials in. WSL users may wish to overwrite this to \"bash\"." default:"infer" enum:"infer,basic,powershell,bash"`
	SessionName string `help:"The name of the role session name that will show up in CloudTrail logs." default:"KeyConjurer-AssumeRole"`
	AccountID   string `arg:"" placeholder:"account-id"`
}

func (SwitchCommand) Help() string {
	return `Attempt to AssumeRole into the given AWS with the current credentials. You only need to use this if you are a power user or network engineer with access to many accounts.

This is used when a "bastion" account exists which users initially authenticate into and then pivot from that account into other accounts.

This command will fail if you do not have active Cloud credentials.`
}

func (s SwitchCommand) RunContext(ctx context.Context) error {
	// We could read the environment variable for the assumed role ARN, but it might be expired which isn't very useful to the user.
	creds, err := getAWSCredentials(ctx, s.AccountID, s.SessionName)
	if err != nil {
		// If this failed, either there was a network error or the user is not authorized to assume into this role
		// This can happen if the user is not authenticated using the Bastion instance.
		return err
	}

	switch s.OutputType {
	case "env":
		creds.WriteFormat(os.Stdout, s.ShellType)
		return nil
	case "aws":
		acc := Account{ID: s.AccountID, Name: s.AccountID}
		newCliEntry := NewCloudCliEntry(creds, &acc)
		return SaveCloudCredentialInCLI(s.AWSCLIPath, newCliEntry)
	default:
		return fmt.Errorf("%s is an invalid output type", s.OutputType)
	}
}

func (s SwitchCommand) Run() error {
	return s.RunContext(context.Background())
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
		AccountID:       accountID,
		AccessKeyID:     *resp.Credentials.AccessKeyId,
		SecretAccessKey: *resp.Credentials.SecretAccessKey,
		SessionToken:    *resp.Credentials.SessionToken,
		Expiration:      resp.Credentials.Expiration.Format(time.RFC3339),
	}

	return creds, nil
}
