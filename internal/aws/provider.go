package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
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

func (p *Provider) GetTemporaryCredentialsForUser(ctx context.Context, principalARN, roleARN, sAMLAssertion *string, ttlInHours int) (*sts.Credentials, error) {
	timeoutInSeconds := int64(3600 * ttlInHours)
	resp, err := p.stsClient.AssumeRoleWithSAMLWithContext(ctx, &sts.AssumeRoleWithSAMLInput{
		DurationSeconds: &timeoutInSeconds,
		PrincipalArn:    principalARN,
		RoleArn:         roleARN,
		SAMLAssertion:   sAMLAssertion,
	})

	if err != nil {
		return nil, err
	}

	return resp.Credentials, nil
}
