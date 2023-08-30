package internal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/riotgames/key-conjurer/api/core"
	"github.com/riotgames/key-conjurer/internal/aws"
	"github.com/riotgames/key-conjurer/internal/tencent"
)

type Provider struct {
	Aws     *aws.Provider
	Tencent *tencent.Provider
}

func NewProvider(awsRegion, tencentRegion string) (*Provider, error) {
	awsProvider, err := aws.NewProvider(awsRegion)
	if err != nil {
		return nil, err
	}
	tencentProvider, err := tencent.NewProvider(tencentRegion)
	if err != nil {
		return nil, err
	}
	return &Provider{Aws: awsProvider, Tencent: tencentProvider}, nil
}

var ErrNoEntitlements = errors.New("user is not entitled to any roles")

type ErrRoleNotFound struct{ Name string }

func (e ErrRoleNotFound) Error() string {
	return fmt.Sprintf("role %s was not found or you do not have access to it", e.Name)
}

type RoleProviderPair struct {
	RoleARN     string
	ProviderARN string
}

const (
	awsRoleUrl     = "https://aws.amazon.com/SAML/Attributes/Role"
	tencentRoleUrl = "https://cloud.tencent.com/SAML/Attributes/Role"
	awsFlag        = 0
	tencentFlag    = 1
)

func ListRoles(response *saml.Response) []string {
	if response == nil {
		return nil
	}

	roleUrl := awsRoleUrl
	roleSubstr := "role/"
	if response.GetAttribute(roleUrl) == "" {
		roleUrl = tencentRoleUrl
		roleSubstr = "roleName/"
	}

	var names []string
	for _, v := range response.GetAttributeValues(roleUrl) {
		p := getARN(v)
		idx := strings.Index(p.RoleARN, roleSubstr)
		parts := strings.Split(p.RoleARN[idx:], "/")
		names = append(names, parts[1])
	}

	return names
}

func FindRole(roleName string, response *saml.Response) (RoleProviderPair, int, bool) {
	if response == nil {
		return RoleProviderPair{}, 0, false
	}

	cloud := awsFlag
	roleUrl := awsRoleUrl
	roleSubstr := "role/"
	if response.GetAttribute(roleUrl) == "" {
		cloud = tencentFlag
		roleUrl = tencentRoleUrl
		roleSubstr = "roleName/"
	}

	if roleName == "" && cloud == awsFlag {
		// This is for legacy support.
		// Legacy clients would always retrieve the first two ARNs in the list, which would be
		//   AWS:
		//       arn:cloud:iam::[account-id]:role/[onelogin_role]
		//       arn:cloud:iam::[account-id]:saml-provider/[saml-provider]
		// If we get weird breakages with Key Conjurer when it's deployed alongside legacy clients, this is almost certainly a culprit!
		pair := getARN(response.GetAttribute(roleUrl))
		return pair, cloud, false
	}

	var pairs []RoleProviderPair
	for _, v := range response.GetAttributeValues(roleUrl) {
		pairs = append(pairs, getARN(v))
	}

	if len(pairs) == 0 {
		return RoleProviderPair{}, cloud, false
	}

	var pair RoleProviderPair
	for _, p := range pairs {
		idx := strings.Index(p.RoleARN, roleSubstr)
		parts := strings.Split(p.RoleARN[idx:], "/")
		if strings.EqualFold(parts[1], roleName) {
			pair = p
		}
	}

	if pair.RoleARN == "" {
		return RoleProviderPair{}, cloud, false
	}

	return pair, cloud, true
}

func getARN(value string) RoleProviderPair {
	p := RoleProviderPair{}
	roles := strings.Split(value, ",")
	if len(roles) >= 2 {
		if strings.Contains(roles[0], "saml-provider/") {
			p.ProviderARN = roles[0]
			p.RoleARN = roles[1]
		} else {
			p.ProviderARN = roles[1]
			p.RoleARN = roles[0]
		}
	}
	return p
}

type STSTokenResponse struct {
	AccessKeyID     *string `json:"accessKeyId"`
	SecretAccessKey *string `json:"secretAccessKey"`
	SessionToken    *string `json:"sessionToken"`
	Expiration      string  `json:"expiration"`
}

func (p *Provider) GetTemporaryCredentialsForUser(ctx context.Context, roleName string, response *core.SAMLResponse, ttlInHours int) (STSTokenResponse, error) {
	pair, cloud, ok := FindRole(roleName, &response.Response)
	if !ok {
		return STSTokenResponse{}, errors.New("role not found")
	}

	switch cloud {
	case awsFlag:
		rsp, err := p.Aws.GetTemporaryCredentialsForUser(ctx, &pair.ProviderARN, &pair.RoleARN, response.GetBase64Encoded(), ttlInHours)
		creds := STSTokenResponse{
			AccessKeyID:     rsp.AccessKeyId,
			SecretAccessKey: rsp.SecretAccessKey,
			SessionToken:    rsp.SessionToken,
			Expiration:      rsp.Expiration.Format(time.RFC3339),
		}
		if err != nil {
			return STSTokenResponse{}, err
		}

		return creds, err
	case tencentFlag:
		rsp, exp, err := p.Tencent.GetTemporaryCredentialsForUser(ctx, &pair.ProviderARN, &pair.RoleARN, response.GetBase64Encoded(), ttlInHours, roleName)
		if err != nil {
			return STSTokenResponse{}, err
		}

		creds := STSTokenResponse{
			AccessKeyID:     rsp.TmpSecretId,
			SecretAccessKey: rsp.TmpSecretKey,
			SessionToken:    rsp.Token,
			Expiration:      *exp,
		}
		return creds, err
	}

	return STSTokenResponse{}, errors.New("unsupported cloud provider")
}
