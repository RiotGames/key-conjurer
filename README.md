# Key Conjurer

![Key Conjurer Champion](docs/champion.png)

Key Conjurer is a project designed to get rid of permanent AWS credentials. This was important to us as it brought down two related risks; compromise of permanent credentials and compromise of a users machines. Luckily, AWS provides their Security Token Service which allows users/services to generate temporary and just-in-time credentials. However, STS just handles the AWS side of the equation and we needed the process of generating tokens to be linked with both our identity provider and MFA. And for that we now have Key Conjurer.

Key Conjurer is made of three parts:

- [api](./api/README.md) -- The lambda based API
- [cli](./cli/README.md) -- The CLI interface
- [frontend](./frontend/README.md) -- The web UI

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

1. Make an S3 Bucket:

```
aws s3api create-bucket --bucket <terraform state bucket> --region us-west-2 --create-bucket-configuration LocationConstraint=us-west-2
```

3. A VPC w/ Subnets to access service
4. Setup a `KMS` key

## Setup Build Environment

- go 1.13.4+
- npm 6.4.1+
- node 10.10.0+
- tfswitch

## Setting Up Your Variable Files

Create `prod.env` based on `example.env`.

## Serverless Settings

### Lambda Env Settings

#### Environment Variables

| Variable          | Purpose                                                           |
| ----------------- | ----------------------------------------------------------------- |
| EncryptedSettings | A KMS encrypted json blob with settings (See below for more info) |

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
| ---------------------- | --------------------------------------- |
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

You'll need to create a Terraform module which references KeyConjurer. We recommend you do this outside of the KeyConjurer folder itself and check your Terraform configuration into source control. An example module that uses KeyConjurer might look like this:


```hcl
resource "aws_acm_certificate" "api-cert" {
  domain_name       = "api.keyconjurer.example.com"
  validation_method = "EMAIL"
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_acm_certificate_validation" "api-cert" {
  certificate_arn = aws_acm_certificate.api-cert.arn
}

resource "aws_acm_certificate" "frontend-cert" {
  domain_name       = "keyconjurer.example.com"
  validation_method = "EMAIL"
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_acm_certificate_validation" "frontend-cert" {
  certificate_arn = aws_acm_certificate.frontend-cert.arn
}

module "keyconjurer-production" {
  source          = "./Key-Conjurer/terraform"
  api_cert        = aws_acm_certificate.development-cert.arn
  api_domain      = aws_acm_certificate.development-cert.domain_name
  frontend_cert   = aws_acm_certificate.frontend-cert.arn
  frontend_domain = aws_acm_certificate.frontend-cert.domain_name
  vpc_id          = "vpc-xxxxxx"
  subnets         = ["subnet-xxxxxx", "subnet-xxxxxx", "subnet-xxxxxx"]
  s3_tf_bucket    = "<the bucket you created in step 1>"
  kms_key_arn     = data.aws_kms_key.development.arn

  lambda_env = {
    VAULT_ADDR              = ""
    VAULT_ROLE_NAME         = "
    VAULT_SECRET_MOUNT_PATH = ""
    VAULT_SECRET_PATH       = ""
    VAULT_AWS_AUTH_PATH     = ""
    SETTINGS_PROVIDER       = "vault"
  }

  lb_security_group_ids = []
  depends_on = [
    aws_acm_certificate_validation.frontend-cert
    aws_acm_certificate_validation.api-cert
  ]
}
```

After modifying `example.env` to your liking, we would recommend renaming this to `prod.env`. You can then deploy KeyConjurer using the following steps:

```
$ pwd
/key-conjurer
$ make build
$ cd terraform
/key-conjurer
$ make upload
$ /your/key-conjurer/terraform/folder
$ terraform apply
```

During your initial deployment, you may need to verify the domain name you've created. This is left as an exercise to the reader; the only thing KeyConjurer requires is _two_ ACM certificates:

1. One for the frontend Cloudfront distribution
2. One for the Load Balancer.

## Future Deploys

Similar to the above steps:

```
$ pwd
/key-conjurer
$ make build
$ cd terraform
/key-conjurer
$ make upload
$ /your/key-conjurer/terraform/folder
$ terraform apply
```


## Noteworthy Info

* `frontend` serves the CLI tool. This means the binaries created in `cli` need to be uploaded to the same bucket that's used to serve the frontend.
* KeyConjurer's Terraform will create an ACL by default unless `create_waf_acl` is set to _false_ and a WAF ACL is provided using `waf_acl_id`. This default ACL will **block all connections**.
* Both a Load Balancer Security Group and a WAF are used to control connections to KeyConjurer. These both need to agree on the IP ranges to allow to KeyConjurer, otherwise you may end up in a situation where a user can access the frontend or use KeyConjurer from the CLI, but not both.

# Development

All pieces of Key Conjurer have been made extensible where possible and use configuration values to select the right plugin.

Adding any new supported authenticator, MFA provider, or cloud provider should be as easy as developing a struct that implemented the given interfaces and ensureing that the interface constructer understands how to initializer and return the new struct.

This section aims to provide details on non-obvious decisions that may impact how one develops and deploys their own plugin.

# Adding an Encryption Provider

KeyConjurer works by sending credentials from the CLI to the server. The server authenticates these credentials, locates all the accounts available for the user, and then returns all of this information, in addition to returning an _encrypted_ version of the credentials for storage by the client.

This is done so that the credentials may be stored on file on the user's machine and later accessed without having to worry about file permissions (as some platforms have better permission support than others).

Encryption was previously tied to the `Provider` interface listed above, but it is now decoupled. You may add an encryption provider by adhering to the [CryptoProvider][cryptoprovider] interface and then using that provider when creating your lambda handler. This will require modification of the codebase for now, and is not configurable via environment variable or compiler switch.

We'll happily accept a pull request making the use of encryption providers a choice at runtime.

[cryptoprovider]: ./api/core/crypto.go
