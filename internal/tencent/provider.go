package tencent

import (
	"context"
	"fmt"
	"os"

	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	sts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sts/v20180813"
)

type Provider struct {
	stsClient *sts.Client
}

func NewProvider(region string) (*Provider, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "sts.tencentcloudapi.com"
	client, _ := sts.NewClient(&common.Credential{}, region, cpf)
	return &Provider{stsClient: client}, nil
}

func (p *Provider) GetTemporaryCredentialsForUser(ctx context.Context, principalARN, roleARN, sAMLAssertion *string, ttlInHours int, roleName string) (*sts.Credentials, *string, error) {
	timeoutInSeconds := int64(3600 * ttlInHours)
	req := sts.NewAssumeRoleWithSAMLRequest()
	req.RoleSessionName = common.StringPtr(fmt.Sprintf("riot-keyConjurer-%s", roleName))
	req.DurationSeconds = common.Uint64Ptr(uint64(timeoutInSeconds))
	req.PrincipalArn = principalARN
	req.RoleArn = roleARN
	req.SAMLAssertion = sAMLAssertion
	resp, err := p.stsClient.AssumeRoleWithSAMLWithContext(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Response.Credentials, resp.Response.Expiration, nil
}

// STS Client
type STSClient struct {
	client *sts.Client
}

// init New STS Client
func NewSTSClient(region string) (*STSClient, error) {
	creds, err := ChainedCredsToCli()
	if err != nil {
		return nil, err
	}
	profile := profile.NewClientProfile()
	profile.Language = "en-US"
	profile.HttpProfile.ReqTimeout = 90
	profile.HttpProfile.Endpoint = "sts.tencentcloudapi.com"
	if region == "" {
		region = regions.SiliconValley
	}
	client, err := sts.NewClient(creds, region, profile)
	if err != nil {
		return nil, err
	}
	return &STSClient{client: client}, nil
}

func (c *STSClient) GetCallerIdentity() (*sts.GetCallerIdentityResponse, error) {
	return c.client.GetCallerIdentity(sts.NewGetCallerIdentityRequest())
}
func (c *STSClient) AssumeRole(roleARN, roleSessionName string) (*sts.AssumeRoleResponse, error) {
	request := sts.NewAssumeRoleRequest()
	request.RoleArn = &roleARN
	request.RoleSessionName = &roleSessionName
	return c.client.AssumeRole(request)
}

// CAM Client
type CAMClient struct {
	client *cam.Client
}

// init New CAM Client
func NewCAMClient(region string) (*CAMClient, error) {
	creds, err := ChainedCredsToCli()
	if err != nil {
		return nil, err
	}
	profile := profile.NewClientProfile()
	profile.Language = "en-US"
	profile.HttpProfile.ReqTimeout = 90
	if region == "" {
		region = regions.SiliconValley
	}
	client, err := cam.NewClient(creds, region, profile)
	if err != nil {
		return nil, err
	}
	return &CAMClient{client: client}, nil
}

// API： GetRoleName
func (c *CAMClient) GetRoleName(roleId string) (roleName string, err error) {
	req := cam.NewGetRoleRequest()
	req.RoleId = &roleId
	roleRsp, err := c.client.GetRole(req)
	fmt.Println(roleRsp.ToJsonString())
	if err != nil {
		return "", err
	}
	return *(roleRsp.Response.RoleInfo.RoleName), nil
}

// client chainedCreds for Cli
func ChainedCredsToCli() (common.CredentialIface, error) {
	providerChain := []common.Provider{
		DefaultEnvProvider(),
	}
	return common.NewProviderChain(providerChain).GetCredential()
}

// for tools login to STS auth
type EnvProvider struct {
	secretIdENV  string
	secretKeyENV string
	tokenENV     string
}

// DefaultEnvProvider return a default provider
// The default environment variable name are TENCENTCLOUD_SECRET_ID and TENCENTCLOUD_SECRET_KEY and TOKEN
func DefaultEnvProvider() *EnvProvider {
	return &EnvProvider{
		secretIdENV:  "TENCENTCLOUD_SECRET_ID",
		secretKeyENV: "TENCENTCLOUD_SECRET_KEY",
		tokenENV:     "TENCENTCLOUD_TOKEN",
	}
}

// GetCredential
func (p *EnvProvider) GetCredential() (common.CredentialIface, error) {
	secretId, ok1 := os.LookupEnv(p.secretIdENV)
	secretKey, ok2 := os.LookupEnv(p.secretKeyENV)
	token, ok3 := os.LookupEnv(p.tokenENV)
	if !ok1 || !ok2 || !ok3 {
		return nil, envNotSet
	}
	if secretId == "" || secretKey == "" || token == "" {
		return nil, tcerr.NewTencentCloudSDKError(creErr,
			"Environmental variable ("+p.secretIdENV+" or "+
				p.secretKeyENV+" or "+p.secretKeyENV+") is empty", "")
	}
	return common.NewTokenCredential(secretId, secretKey, token), nil
}

var creErr = "ClientError.CredentialError"
var envNotSet = tcerr.NewTencentCloudSDKError(creErr, "could not find environmental variable", "")
