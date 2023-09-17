package main

import (
	"strings"

	"github.com/RobotsAndPencils/go-saml"
)

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

func ListSAMLRoles(response *saml.Response) []string {
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

func FindRoleInSAML(roleName string, response *saml.Response) (RoleProviderPair, int, bool) {
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
