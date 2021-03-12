package aws

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RobotsAndPencils/go-saml"

	"github.com/riotgames/key-conjurer/api/core"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

type ErrRoleNotFound struct{ Name string }

func (e ErrRoleNotFound) Error() string {
	return fmt.Sprintf("role %s was not found or you do not have access to it", e.Name)
}

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

type STSTokenResponse struct {
	AccessKeyID     *string `json:"accessKeyId"`
	SecretAccessKey *string `json:"secretAccessKey"`
	SessionToken    *string `json:"sessionToken"`
	Expiration      string  `json:"expiration"`
}

type roleProviderPair struct {
	RoleARN     string
	ProviderARN string
}

func getRole(roleName string, response *saml.Response) (string, string, error) {
	if response == nil {
		return "", "", errors.New("Unable to get SAML assertion")
	}

	if roleName == "" {
		// This is for legacy support.
		// Legacy clients would always retrieve the first two ARNs in the list, which would be
		//   arn:aws:iam::[account-id]:role/[onelogin_role]
		//   arn:aws:iam::[account-id]:saml-provider/[saml-provider]
		// If we get weird breakages with Key Conjurer when it's deployed alongside legacy clients, this is almost certainly a culprit!
		roles := strings.Split(response.GetAttribute("https://aws.amazon.com/SAML/Attributes/Role"), ",")
		return roles[1], roles[0], nil
	}

	var pairs []roleProviderPair
	for _, v := range response.GetAttributeValues("https://aws.amazon.com/SAML/Attributes/Role") {
		parts := strings.Split(v, ",")
		pairs = append(pairs, roleProviderPair{RoleARN: parts[1], ProviderARN: parts[0]})
	}

	var pair roleProviderPair
	for _, p := range pairs {
		idx := strings.Index(p.RoleARN, "role/")
		parts := strings.Split(p.RoleARN[idx:], "/")
		if parts[1] == roleName {
			pair = p
		}
	}

	if pair.RoleARN == "" {
		return "", "", ErrRoleNotFound{Name: roleName}
	}

	return pair.ProviderARN, pair.RoleARN, nil
}

func (p *Provider) GetTemporaryCredentialsForUser(ctx context.Context, roleName string, response *core.SAMLResponse, ttlInHours int) (STSTokenResponse, error) {
	principalARN, roleARN, err := getRole(roleName, &response.Response)
	if err != nil {
		return STSTokenResponse{}, err
	}

	timeoutInSeconds := int64(3600 * ttlInHours)
	resp, err := p.stsClient.AssumeRoleWithSAMLWithContext(ctx, &sts.AssumeRoleWithSAMLInput{
		DurationSeconds: &timeoutInSeconds,
		PrincipalArn:    &principalARN,
		RoleArn:         &roleARN,
		SAMLAssertion:   response.GetBase64Encoded(),
	})

	if err != nil {
		return STSTokenResponse{}, err
	}

	credentials := resp.Credentials
	return STSTokenResponse{
		AccessKeyID:     credentials.AccessKeyId,
		SecretAccessKey: credentials.SecretAccessKey,
		SessionToken:    credentials.SessionToken,
		Expiration:      credentials.Expiration.Format(time.RFC3339),
	}, nil
}
