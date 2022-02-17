package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/spf13/cobra"
)

var roleSessionName string

func init() {
	switchCmd.Flags().StringVar(&roleSessionName, "role-session-name", "KeyConjurer-AssumeRole", "the name of the role session name that will show up in CloudTrail logs")
	switchCmd.Flags().StringVarP(&outputType, "out", "o", outputTypeEnvironmentVariable, "Format to save new credentials in. Supported outputs: env, awscli")
	switchCmd.Flags().StringVarP(&shell, "shell", "", shellTypeInfer, "If output type is env, determines which format to output credentials in - by default, the format is inferred based on the execution environment. WSL users may wish to overwrite this to `bash`")
	switchCmd.Flags().StringVarP(&awsCliPath, "awscli", "", "~/.aws/", "Path for directory used by the aws-cli tool. Default is \"~/.aws\".")
}

var switchCmd = cobra.Command{
	Use:   "switch <account-id>",
	Short: "Switch from the current AWS account into the one with the given Account ID.",
	Long: `Attempt to AssumeRole into the given AWS account with the current credentials. You only need to use this if you are a power user or network engineer with access to many accounts.

This is used when a "bastion" account exists which users initially authenticate into and then pivot from that account into other accounts.

This command will fail if you do not have active AWS credentials.
`,
	Example: `keyconjurer switch 123456798`,
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"switch-account"},
	RunE: func(comm *cobra.Command, args []string) error {
		valid := false
		for _, permitted := range permittedOutputTypes {
			if outputType == permitted {
				valid = true
			}
		}

		if !valid {
			return invalidValueError(outputType, permittedOutputTypes)
		}

		for _, permitted := range permittedShellTypes {
			if shell == permitted {
				valid = true
			}
		}

		if !valid {
			return invalidValueError(shell, permittedShellTypes)
		}

		// We could read the environment variable for the assumed role ARN, but it might be expired which isn't very useful to the user.
		ctx := context.Background()
		sess, err := session.NewSession(aws.NewConfig())
		if err != nil {
			return err
		}

		c := sts.New(sess)
		response, err := c.GetCallerIdentityWithContext(ctx, &sts.GetCallerIdentityInput{})
		if err != nil {
			return err
		}

		// We need to modify this to change the last section to be role/GL-SuperAdmin
		id, err := arn.Parse(*response.Arn)
		if err != nil {
			return err
		}

		parts := strings.Split(id.Resource, "/")
		arn := arn.ARN{
			AccountID: args[0],
			Partition: "aws",
			Service:   "iam",
			Resource:  fmt.Sprintf("role/%s", parts[1]),
			Region:    id.Region,
		}

		roleARN := arn.String()
		resp, err := c.AssumeRoleWithContext(ctx, &sts.AssumeRoleInput{
			RoleArn:         &roleARN,
			RoleSessionName: &roleSessionName,
		})

		if err != nil {
			// If this failed, either there was a network error or the user is not authorized to assume into this role
			// This can happen if the user is not authenticated using the Bastion instance.
			return err
		}

		creds := AWSCredentials{
			AccountID:       args[0],
			AccessKeyID:     *resp.Credentials.AccessKeyId,
			SecretAccessKey: *resp.Credentials.SecretAccessKey,
			SessionToken:    *resp.Credentials.SessionToken,
			Expiration:      resp.Credentials.Expiration.Format(time.RFC3339),
		}

		switch outputType {
		case outputTypeEnvironmentVariable:
			creds.WriteFormat(os.Stdout, shell)
			return nil
		case outputTypeAWSCredentialsFile:
			acc := Account{ID: args[0], Name: args[0]}
			newCliEntry := NewAWSCliEntry(&creds, &acc)
			return SaveAWSCredentialInCLI(awsCliPath, newCliEntry)
		default:
			return fmt.Errorf("%s is an invalid output type", outputType)
		}
	},
}
