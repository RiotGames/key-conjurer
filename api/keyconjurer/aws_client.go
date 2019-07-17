package keyconjurer

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"keyconjurer-lambda/authenticators"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/sirupsen/logrus"
)

// AWSClient provides an interface into the required
//  AWS actions for KeyuConjurer
type AWSClient struct {
	kmsClient *kms.KMS
	stsClient *sts.STS
	kmsKeyID  string
	logger    *logrus.Entry
}

// NewAWSClient creates a new AWS Client in the region provided
//  and will use the provided logger for logging
func NewAWSClient(awsRegion string, logger *logrus.Entry) *AWSClient {
	config := session.New(&aws.Config{
		Region: aws.String(awsRegion)})

	awsClient := &AWSClient{
		kmsClient: kms.New(config),
		stsClient: sts.New(config),
		logger:    logger}

	return awsClient
}

// SetKMSKeyID stores the key ID that the AWSClient will use for
//  crypto operations
func (a *AWSClient) SetKMSKeyID(kmsKeyID string) {
	a.kmsKeyID = kmsKeyID
}

// Encrypt is a wrapper around KMS encrypt. It marshals the interface
//  provided before encrypting.
func (a *AWSClient) Encrypt(data interface{}) (string, error) {
	plaintext, err := json.Marshal(data)
	if err != nil {
		a.logger.Error("AWSClient", "Encrypt", "Failed to marshal data", err.Error())
		return "", ErrorJsonMarshalError
	}

	input := &kms.EncryptInput{
		KeyId:     aws.String(a.kmsKeyID),
		Plaintext: plaintext}
	ciphertext, err := a.kmsClient.Encrypt(input)

	if err != nil {
		a.logger.Error("failed to encrypt reason: ", err.Error())
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext.CiphertextBlob), nil
}

// Decrypt is a wrapper around KMS decrypt. It unmarshals the plaintext
//  into the provided interface.
func (a *AWSClient) Decrypt(ciphertext string, v interface{}) error {
	blob, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		a.logger.Error("unable to decode ciphertext reason: ", err.Error())
		return ErrorUnableToDecode
	}
	input := &kms.DecryptInput{
		CiphertextBlob: blob}

	result, err := a.kmsClient.Decrypt(input)
	if err != nil {
		a.logger.Error("aws client failed to decrypt reason: ", err.Error())
		return err
	}
	if err := json.Unmarshal(result.Plaintext, v); err != nil {
		a.logger.Error("unable to unmarshal reason: ", err.Error())
		return ErrorJsonUnmarshalError
	}
	return nil
}

// SelectRoleFromSaml will pull out the role ARN that will be assumed and the principal
//  that will be assuming the role
func (a *AWSClient) SelectRoleFromSaml(samleResponse authenticators.SamlResponse) (string, string, error) {
	if samleResponse == nil {
		a.logger.Error("saml assertion is nil pointer")
		return "", "", ErrorUnableToGetSamlAssertion
	}

	roleInfo := strings.Split(samleResponse.GetSamlResponse().GetAttribute("https://aws.amazon.com/SAML/Attributes/Role"), ",")
	if len(roleInfo) != 2 {
		a.logger.Error("saml assertion has too many roles")
		return "", "", ErrorSamlAssertionHasTooManyRoles
	}
	roleArn := roleInfo[0]
	principalArn := roleInfo[1]
	return roleArn, principalArn, nil
}

// AssumeRole is a wrapper around AWS sts.AssumeRoleWithSAMLInput. After assuming role it returns
//  the temporary credentials
func (a *AWSClient) AssumeRole(roleArn, principalArn string, samlResponse authenticators.SamlResponse, timeoutInHours int) (*sts.Credentials, error) {
	samlString := samlResponse.GetBase64String()

	timeoutInSeconds := int64(3600 * timeoutInHours)
	input := &sts.AssumeRoleWithSAMLInput{
		DurationSeconds: &timeoutInSeconds,
		PrincipalArn:    &principalArn,
		RoleArn:         &roleArn,
		SAMLAssertion:   &samlString}
	resp, err := a.stsClient.AssumeRoleWithSAML(input)
	if err != nil {
		a.logger.Error("unable to assume role")
		return nil, err
	}
	return resp.Credentials, nil
}
