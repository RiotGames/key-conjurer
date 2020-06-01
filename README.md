Key Conjurer
=====
![Key Conjurer Champion](docs/champion.png)

Key Conjurer is a project designed to get rid of permanent AWS credentials. This was important to us as it brought down two related risks; compromise of permanent credentials and compromise of a users machines. Luckily, AWS provides their Security Token Service which allows users/services to generate temporary and just-in-time credentials. However, STS just handles the AWS side of the equation and we needed the process of generating tokens to be linked with both our identity provider and MFA. And for that we now have Key Conjurer.

Key Conjurer is made of three parts:
* [api](./api/README.md) -- The lambda based API
* [cli](./cli/README.md) -- The CLI interface
* [frontend](./frontend/README.md) -- The web UI

Key Conjurer currently supports the following identity providers and mfa services:
  - Identity Providers:
    - onelogin
  - MFA:
    - duo

Key Conjurer now supports the ability to provide temporary crendentials different cloud providers as well as being deployed on different platforms. 

Currently supported credential providers are:
  - AWS STS

Current platforms supported for deployment are:
  - AWS

# Pre-Deployment Steps

## Platform Pre-Deployment Resources
### Generate/Initialize AWS Resources
1. Certificates - Make sure a certificate in ACM is requested with the desired hostname (the arn will be needed later)
```
aws acm request-certificate --domain-name <api domain> --validation-method EMAIL --region us-east-1
aws acm request-certificate --domain-name <frontend domain> --validation-method EMAIL --region us-east-1	
```
2. Make an S3 Bucket:
```
aws s3api create-bucket --bucket <terraform state bucket> --region us-west-2 --create-bucket-configuration LocationConstraint=us-west-2 
```
3. A VPC w/ Subnets to access service
4. Setup a `KMS` key

## Setup Build Environment
  - go 1.12.4+
  - npm 6.4.1+
  - node 10.10.0+
  - tfswitch

## Setting Up Your Variable Files
Create `prod.env` based on `example.env`.

## Serverless Settings
### Lambda Env Settings
#### Environment Variables
| Variable          | Purpose                                                           |
|-------------------|-------------------------------------------------------------------|
| EncryptedSettings | A KMS encrypted json blob with settings (See below for more info) |
| AWSRegion         | Used for KMS Region. Typically the same region KeyConjuer is in   |

##### Encrypted Settings
The encrypted settings are a JSON blob with the following keys.
```
{
  "awsKmsKeyId": "abc...",
  "oneLoginReadUserId": "def...",
  "oneLoginReadUserSecret": "ghi...",
  "oneLoginSamlId": "jkl...",
  "oneLoginSamlSecret": "lmn...",
  "oneLoginShard": "opq...",
  "oneLoginSubdomain": "rst..."
}
```
| Variable               | Purpose                                 |
|------------------------|-----------------------------------------|
| awsKmsKeyId            | The KMS key to encrypt information with |
| oneLoginReadUserId     | OneLogin key with read user permissions |
| oneLoginReadUserSecret | Secret key for oneLoginReadUserId       |
| oneLoginSamlId         | OneLogin key with SAML permissions      |
| oneLoginSamlSecret     | Secret key for oneLoginSamlId           |
| oneLoginShard          | OneLogin shard to talk with             |
| oneLoginSubdomain      | OneLogin subdomain                      |

They are encrypted so users with access to the lambdas cannot see the secrets

# Deploying
These steps assume you created `prod.env` as instructed above.
## First Deploy
```
source prod.env
make build deploy
```
*When Deploying to AWS* Ensure the IAM role provisioned by `terraform` has access to use the `KMS` key created above

## Future Deploys
```
source prod.env
make deploy
```

## Noteworthy Info
`frontend` serves the CLI tool. This means the binaries created in `cli`
need to be uploaded to the same bucket that's used to serve the frontend.

# Development

All pieces of Key Conjurer have been made extensible where possible and use configuration values to select the right plugin.

Adding any new supported authenticator, MFA provider, or cloud provider should be as easy as developing a struct that implemented the given interfaces and ensureing that the interface constructer understands how to initializer and return the new struct.

This section aims to provide details on non-obvious decisions that may impact how one develops and deploys their own plugin.

## Adding A Cloud Provider

Currently, Key Conjurer only has code written to support AWS but could provide support to any other cloud provider as long as it aligns with the following given interface (api/cloud-providers/provider.go):

```go
type Provider interface {
	GetUserCredentials(username, password string) (*User, error)
	DecryptUserInforamtion(ciphertext string, user interface{}) error
	EncryptUserInformation(data interface{}) (string, error)
	GetTemporaryCredentialsForUser(samlAssertion authenticators.SamlResponse, ttlInHours int) (interface{}, error)
}
```

The constructor for this iteraface is currently:

```go
func NewProvider(settings *settings.Settings, logger *logrus.Entry) (Provider, error) {
	if settings.AwsRegion != "" && settings.AwsKMSKeyID != "" {
		return NewAWSProvider(settings, logger)
	}

	return nil, errors.New("no matching credentials safe for settings")
}
```

Due to this snippet, there is a default order to provisioning of cloud providers assuming all configuration items are listed. This order is:

  1. AWS