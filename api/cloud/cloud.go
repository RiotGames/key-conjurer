package cloud

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/riotgames/key-conjurer/api/cloud/aws"
	"github.com/riotgames/key-conjurer/api/cloud/base"
	"github.com/riotgames/key-conjurer/api/cloud/tencent"
	"github.com/riotgames/key-conjurer/api/core"
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

type roleProviderPair struct {
	RoleARN     string
	ProviderARN string
}

const (
	awsRoleUrl     = "https://aws.amazon.com/SAML/Attributes/Role"
	tencentRoleUrl = "https://cloud.tencent.com/SAML/Attributes/Role"
	awsFlag        = 0
	tencentFlag    = 1
)

func getRole(roleName string, response *saml.Response) (string, string, int, error) {
	if response == nil {
		return "", "", 0, errors.New("unable to get SAML assertion")
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
		return pair.ProviderARN, pair.RoleARN, cloud, nil
	}

	var pairs []roleProviderPair
	for _, v := range response.GetAttributeValues(roleUrl) {
		pairs = append(pairs, getARN(v))
	}

	if len(pairs) == 0 {
		return "", "", cloud, ErrNoEntitlements
	}

	var pair roleProviderPair
	for _, p := range pairs {
		idx := strings.Index(p.RoleARN, roleSubstr)
		parts := strings.Split(p.RoleARN[idx:], "/")
		if strings.EqualFold(parts[1], roleName) {
			pair = p
		}
	}

	if pair.RoleARN == "" {
		return "", "", cloud, ErrRoleNotFound{Name: roleName}
	}

	return pair.ProviderARN, pair.RoleARN, cloud, nil
}

func getARN(value string) roleProviderPair {
	p := roleProviderPair{}
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

func (p *Provider) GetTemporaryCredentialsForUser(ctx context.Context, roleName string, response *core.SAMLResponse, ttlInHours int) (int, base.STSTokenResponse, error) {
	principalARN, roleARN, cloud, err := getRole(roleName, &response.Response)
	if err != nil {
		return 0, base.STSTokenResponse{}, err
	}
	switch cloud {
	case awsFlag:
		rsp, err := p.Aws.GetTemporaryCredentialsForUser(ctx, &principalARN, &roleARN, response.GetBase64Encoded(), ttlInHours)
		return awsFlag, rsp, err
	case tencentFlag:
		rsp, err := p.Tencent.GetTemporaryCredentialsForUser(ctx, &principalARN, &roleARN, response.GetBase64Encoded(), ttlInHours, roleName)
		return tencentFlag, rsp, err
	}
	return 0, base.STSTokenResponse{}, fmt.Errorf("can't find cloud vendors")
}
