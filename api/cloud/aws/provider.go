package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/riotgames/key-conjurer/api/cloud/base"
)

type Provider struct {
	stsClient *sts.STS
}

func NewProvider(region string) (*Provider, error) {
	config, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}
	return &Provider{stsClient: sts.New(config)}, nil
}

func (p *Provider) GetTemporaryCredentialsForUser(ctx context.Context, principalARN, roleARN, sAMLAssertion *string, ttlInHours int) (base.STSTokenResponse, error) {
	timeoutInSeconds := int64(3600 * ttlInHours)
	resp, err := p.stsClient.AssumeRoleWithSAMLWithContext(ctx, &sts.AssumeRoleWithSAMLInput{
		DurationSeconds: &timeoutInSeconds,
		PrincipalArn:    principalARN,
		RoleArn:         roleARN,
		SAMLAssertion:   sAMLAssertion,
	})
	if err != nil {
		return base.STSTokenResponse{}, err
	}
	credentials := resp.Credentials
	return base.STSTokenResponse{
		AccessKeyID:     credentials.AccessKeyId,
		SecretAccessKey: credentials.SecretAccessKey,
		SessionToken:    credentials.SessionToken,
		Expiration:      credentials.Expiration.Format(time.RFC3339),
	}, nil
}
