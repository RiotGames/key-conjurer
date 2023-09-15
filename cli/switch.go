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
	"github.com/riotgames/key-conjurer/internal/tencent"
	"github.com/spf13/cobra"
)

var (
	FlagRoleSessionName = "role-session-name"
	FlagOutputType      = "out"
	FlagShellType       = "shell"
	FlagAWSCLIPath      = "awscli"
	FlagTencentCLIPath  = "tencentcli"
	FlagCloudType       = "cloud"
)

func init() {
	switchCmd.Flags().String(FlagRoleSessionName, "KeyConjurer-AssumeRole", "the name of the role session name that will show up in CloudTrail logs")
	switchCmd.Flags().StringP(FlagOutputType, "o", outputTypeEnvironmentVariable, "Format to save new credentials in. Supported outputs: env, awscli,tencentcli")
	switchCmd.Flags().String(FlagShellType, shellTypeInfer, "If output type is env, determines which format to output credentials in - by default, the format is inferred based on the execution environment. WSL users may wish to overwrite this to `bash`")
	switchCmd.Flags().String(FlagAWSCLIPath, "~/.aws/", "Path for directory used by the aws-cli tool. Default is \"~/.aws\".")
	switchCmd.Flags().String(FlagTencentCLIPath, "~/.tencent/", "Path for directory used by the tencent-cli tool. Default is \"~/.tencent\".")
	switchCmd.Flags().String(FlagCloudType, "aws", "Choose a cloud vendor. Default is aws. Can choose aws or tencent")
}

var switchCmd = cobra.Command{
	Use:   "switch <account-id>",
	Short: "Switch from the current Cloud (AWS or Tencent) account into the one with the given Account ID.",
	Long: `Attempt to AssumeRole into the given Cloud (AWS or Tencent) with the current credentials. You only need to use this if you are a power user or network engineer with access to many accounts.

This is used when a "bastion" account exists which users initially authenticate into and then pivot from that account into other accounts.

This command will fail if you do not have active Cloud credentials.
`,
	Example: "keyconjurer switch 123456798",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"switch-account"},
	RunE: func(cmd *cobra.Command, args []string) error {
		outputType, _ := cmd.Flags().GetString(FlagOutputType)
		shellType, _ := cmd.Flags().GetString(FlagShellType)
		cloudType, _ := cmd.Flags().GetString(FlagCloudType)
		awsCliPath, _ := cmd.Flags().GetString(FlagAWSCLIPath)
		if !isMemberOfSlice(permittedOutputTypes, outputType) {
			return invalidValueError(outputType, permittedOutputTypes)
		}

		if !isMemberOfSlice(permittedShellTypes, shellType) {
			return invalidValueError(shellType, permittedShellTypes)
		}

		// We could read the environment variable for the assumed role ARN, but it might be expired which isn't very useful to the user.
		var err error
		var creds CloudCredentials
		sessionName, _ := cmd.Flags().GetString(FlagRoleSessionName)
		switch strings.ToLower(cloudType) {
		case cloudAws:
			creds, err = getAWSCredentials(args[0], sessionName)
		case cloudTencent:
			creds, err = getTencentCredentials(args[0], sessionName)
		}

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

func getTencentCredentials(accountId, roleSessionName string) (creds CloudCredentials, err error) {
	region := os.Getenv("TENCENT_REGION")
	stsClient, err := tencent.NewSTSClient(region)
	if err != nil {
		return
	}

	response, err := stsClient.GetCallerIdentity()
	if err != nil {
		return
	}

	arn := response.Response.Arn
	roleId := ""
	if (*arn) != "" {
		arns := strings.Split(*arn, ":")
		if len(arns) >= 5 && len(strings.Split(arns[4], "/")) >= 2 {
			roleId = strings.Split(arns[4], "/")[1]
		}
	}
	if roleId == "" {
		err = fmt.Errorf("roleId is null")
		return
	}

	camClient, err := tencent.NewCAMClient(region)
	if err != nil {
		return
	}
	roleName, err := camClient.GetRoleName(roleId)
	if err != nil {
		return
	}
	resp, err := stsClient.AssumeRole(fmt.Sprintf("qcs::cam::uin/%s:roleName/%s", accountId, roleName), roleSessionName)
	if err != nil {
		return
	}

	creds = CloudCredentials{
		AccountID:       accountId,
		AccessKeyID:     *resp.Response.Credentials.TmpSecretId,
		SecretAccessKey: *resp.Response.Credentials.TmpSecretKey,
		SessionToken:    *resp.Response.Credentials.Token,
		Expiration:      *resp.Response.Expiration,
		credentialsType: cloudTencent,
	}

	return creds, nil
}

func getAWSCredentials(accountId, roleSessionName string) (creds CloudCredentials, err error) {
	ctx := context.Background()
	sess, err := session.NewSession(aws.NewConfig())
	if err != nil {
		return
	}

	c := sts.New(sess)
	response, err := c.GetCallerIdentityWithContext(ctx, &sts.GetCallerIdentityInput{})
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
		AccountID: accountId,
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
		return
	}

	creds = CloudCredentials{
		AccountID:       accountId,
		AccessKeyID:     *resp.Credentials.AccessKeyId,
		SecretAccessKey: *resp.Credentials.SecretAccessKey,
		SessionToken:    *resp.Credentials.SessionToken,
		Expiration:      resp.Credentials.Expiration.Format(time.RFC3339),
		credentialsType: cloudAws,
	}

	return creds, nil
}
