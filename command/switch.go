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
	"github.com/spf13/cobra"
)

var (
	FlagRoleSessionName = "role-session-name"
	FlagOutputType      = "out"
	FlagShellType       = "shell"
	FlagAWSCLIPath      = "awscli"
)

func init() {
	switchCmd.Flags().String(FlagRoleSessionName, "KeyConjurer-AssumeRole", "the name of the role session name that will show up in CloudTrail logs")
	switchCmd.Flags().StringP(FlagOutputType, "o", outputTypeEnvironmentVariable, "Format to save new credentials in. Supported outputs: env, awscli")
	switchCmd.Flags().String(FlagShellType, shellTypeInfer, "If output type is env, determines which format to output credentials in - by default, the format is inferred based on the execution environment. WSL users may wish to overwrite this to `bash`")
	switchCmd.Flags().String(FlagAWSCLIPath, "~/.aws/", "Path for directory used by the aws-cli tool. Default is \"~/.aws\".")
}

var switchCmd = cobra.Command{
	Use:   "switch <account-id>",
	Short: "Switch from the current AWS account into the one with the given Account ID.",
	Long: `Attempt to AssumeRole into the given AWS with the current credentials. You only need to use this if you are a power user or network engineer with access to many accounts.

This is used when a "bastion" account exists which users initially authenticate into and then pivot from that account into other accounts.

This command will fail if you do not have active Cloud credentials.
`,
	Example: "keyconjurer switch 123456798",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"switch-account"},
	RunE: func(cmd *cobra.Command, args []string) error {
		outputType, _ := cmd.Flags().GetString(FlagOutputType)
		shellType, _ := cmd.Flags().GetString(FlagShellType)
		awsCliPath, _ := cmd.Flags().GetString(FlagAWSCLIPath)
		if !slices.Contains(permittedOutputTypes, outputType) {
			return ValueError{Value: outputType, ValidValues: permittedOutputTypes}
		}

		if !slices.Contains(permittedShellTypes, shellType) {
			return ValueError{Value: shellType, ValidValues: permittedShellTypes}
		}

		// We could read the environment variable for the assumed role ARN, but it might be expired which isn't very useful to the user.
		var err error
		var creds CloudCredentials
		sessionName, _ := cmd.Flags().GetString(FlagRoleSessionName)

		creds, err = getAWSCredentials(args[0], sessionName)
		if err != nil {
			// If this failed, either there was a network error or the user is not authorized to assume into this role
			// This can happen if the user is not authenticated using the Bastion instance.
			return err
		}

		switch outputType {
		case outputTypeEnvironmentVariable:
			creds.WriteFormat(os.Stdout, shellType)
			return nil
		case outputTypeAWSCredentialsFile:
			acc := Account{ID: args[0], Name: args[0]}
			newCliEntry := NewCloudCliEntry(creds, &acc)
			return SaveCloudCredentialInCLI(awsCliPath, newCliEntry)
		default:
			return fmt.Errorf("%s is an invalid output type", outputType)
		}
	},
}

func getAWSCredentials(accountID, roleSessionName string) (creds CloudCredentials, err error) {
	ctx := context.Background()
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
