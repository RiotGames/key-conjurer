package cloudprovider

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/riotgames/key-conjurer/api/authenticators"
	"github.com/riotgames/key-conjurer/api/settings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/sirupsen/logrus"
)

type AWSProvider struct {
	kmsClient *kms.KMS
	stsClient *sts.STS
	kmsKeyID  string
	logger    *logrus.Entry
}

type STSTokenResponse struct {
	AccessKeyID     *string `json:"accessKeyId"`
	SecretAccessKey *string `json:"secretAccessKey"`
	SessionToken    *string `json:"sessionToken"`
	Expiration      string  `json:"expiration"`
}

func NewAWSProvider(settings *settings.Settings, logger *logrus.Entry) (*AWSProvider, error) {
	config := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(settings.AwsRegion)}))

	return &AWSProvider{
		kmsClient: kms.New(config),
		stsClient: sts.New(config),
		kmsKeyID:  settings.AwsKMSKeyID,
		logger:    logger,
	}, nil
}

func (p *AWSProvider) GetUserCredentials(username, password string) (*User, error) {
	var user User

	if username == "encrypted" {
		if err := p.DecryptUserInforamtion(password, &user); err != nil {
			p.logger.Info("creds decryption failure reason: ", err.Error())
			return nil, errors.New("Invalid username or password")
		}
	} else {
		user.Username = username
		user.Password = password
	}

	return &user, nil
}

func (p *AWSProvider) EncryptUserInformation(data interface{}) (string, error) {
	plaintext, err := json.Marshal(data)
	if err != nil {
		p.logger.Error("AWSClient", "Encrypt", "Failed to marshal data", err.Error())
		return "", errors.New("Unable to marshal json")
	}

	input := &kms.EncryptInput{
		KeyId:     aws.String(p.kmsKeyID),
		Plaintext: plaintext}
	ciphertext, err := p.kmsClient.Encrypt(input)

	if err != nil {
		p.logger.Error("failed to encrypt reason: ", err.Error())
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext.CiphertextBlob), nil
}

func (p *AWSProvider) DecryptUserInforamtion(ciphertext string, value interface{}) error {

	blob, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		p.logger.Error("unable to decode ciphertext reason: ", err.Error())
		return errors.New("Unable to base64 decode string")
	}
	input := &kms.DecryptInput{
		CiphertextBlob: blob}

	result, err := p.kmsClient.Decrypt(input)
	if err != nil {
		p.logger.Error("aws client failed to decrypt reason: ", err.Error())
		return err
	}
	if err := json.Unmarshal(result.Plaintext, value); err != nil {
		p.logger.Error("unable to unmarshal reason: ", err.Error())
		return errors.New("Unable to unmarshal json")
	}
	return nil
}

func (p *AWSProvider) GetTemporaryCredentialsForUser(samlAssertion authenticators.SamlResponse, ttlInHours int) (interface{}, error) {
	roleArn, principalArn, err := p.selectRoleFromSaml(samlAssertion)
	if err != nil {
		p.logger.Error("unable to select role from saml reason: ", err.Error())
		return nil, err
	}

	p.logger.Info("KeyConjurer", "Assuming role")
	credentials, err := p.assumeRole(roleArn, principalArn, samlAssertion, ttlInHours)
	if err != nil {
		p.logger.Error("unable to assume role reason: ", err.Error())
		return nil, err
	}

	return STSTokenResponse{
		AccessKeyID:     credentials.AccessKeyId,
		SecretAccessKey: credentials.SecretAccessKey,
		SessionToken:    credentials.SessionToken,
		Expiration:      credentials.Expiration.Format(time.RFC3339),
	}, nil
}

func (p *AWSProvider) selectRoleFromSaml(samleResponse authenticators.SamlResponse) (string, string, error) {
	if samleResponse == nil {
		p.logger.Error("saml assertion is nil pointer")
		return "", "", errors.New("Unable to get SAML assertion")
	}

	roleInfo := strings.Split(samleResponse.GetSamlResponse().GetAttribute("https://aws.amazon.com/SAML/Attributes/Role"), ",")
	if len(roleInfo) != 2 {
		p.logger.Error("saml assertion has too many roles")
		return "", "", errors.New("SAML assertion has too many roles")
	}
	roleArn := roleInfo[0]
	principalArn := roleInfo[1]
	return roleArn, principalArn, nil
}

// AssumeRole is a wrapper around AWS sts.AssumeRoleWithSAMLInput. After assuming role it returns
//  the temporary credentials
func (p *AWSProvider) assumeRole(roleArn, principalArn string, samlResponse authenticators.SamlResponse, timeoutInHours int) (*sts.Credentials, error) {
	samlString := samlResponse.GetBase64String()

	timeoutInSeconds := int64(3600 * timeoutInHours)
	input := &sts.AssumeRoleWithSAMLInput{
		DurationSeconds: &timeoutInSeconds,
		PrincipalArn:    &principalArn,
		RoleArn:         &roleArn,
		SAMLAssertion:   &samlString}
	resp, err := p.stsClient.AssumeRoleWithSAML(input)
	if err != nil {
		p.logger.Error("unable to assume role")
		return nil, err
	}

	return resp.Credentials, nil
}
