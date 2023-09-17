# KeyConjurer

![Key Conjurer Champion](docs/champion.png)

KeyConjurer is a project designed to get rid of permanent AWS credentials.

KeyConjurer is made of three parts:

- [lambda](./lambda/) - Lambda functions used by the CLI to gather data on protected resources.
- [cli](./cli/) - The CLI interface.
- [frontend](./frontend/) - A static webpage which informs users on how to download and use KeyConjurer.

KeyConjurer is designed to work with Okta as an IdP, supports AWS and Tencent Cloud applications, and is inspired in part by [okta-aws-cli](https://github.com/okta/okta-aws-cli). The main difference from okta-aws-cli is that KeyConjurer does not require all users to have access to the Okta administration API - Instead, we use a Lambda function to access the protected resources required.

We use KeyConjurer a lot at Riot, but we can't guarantee any external support for this project. It's use at your own risk. If you encounter a bug or have a feature request, please feel free to raise a pull request or an issue against this repository. You're also welcome to fork the code and modify it as you see fit.


# Pre-Deployment Steps

## Platform Pre-Deployment Resources

1. Make an S3 Bucket:

```
aws s3api create-bucket --bucket <terraform state bucket> --region us-west-2 --create-bucket-configuration LocationConstraint=us-west-2
```

3. A VPC w/ Subnets to access service
4. Setup a `KMS` key

## Setup Build Environment

- go 1.20+
- node 16.17.0+
- terraform 1.3.7+

## Setting Up Your Variable Files

Create `prod.env` based on `example.env`.

### Configuration

#### Okta setup

In order to use KeyConjurer, an Okta administrator must configure their tenant appropriately:

* A new _native_ OIDC application must be created within your Okta tenant, and the following settings must be configured:
  * Scopes: `profile openid okta.apps.read`
  * Authorization Types: Hybrid Flow, Authorization Code, Token Exchange
  * Redirection URI: http://localhost:57468
  * We recommend you enable Federated Mode on this native application so that users don't need to be explicitly assigned to it.
* All AWS and tencent applications must have their Allowed Web SSO Client set to the _Client ID_ of the native OIDC application that was created. This can be configured by going to the Sign On tab for each individual Okta application or managing the application configuration in an IAC provider, like Terraform.

Okta configuration should be configured _out of band_ and is not provided in this repository.

#### Lambda functions settings

A single lambda function is used to filter applications within the organization to just the ones the user has access to. This function is required because enumerating applications within Okta's API is currently considered an administrative action, and as such, using a users access token to perform this action requires the user to be an administrator on the Okta tenant.

The lambda function has a couple of sensitive values. We use Vault at Riot to store sensitive values. The Lambda function must be configured to access Vault. Secrets can also be retrieved from environment variables directly, but we do not recommend it.

#### Vault

To use Vault, the following environment variables must be configured:

| Variable          | Purpose                                                           |
| ----------------- | ----------------------------------------------------------------- |
| VAULT_ROLE_NAME   | The name of the Vault role to use to acquire credentials          |
| VAULT_SECRET_MOUNT_PATH | The mount path of your Vault secrets mount                  |
| VAULT_SECRET_PATH | The path to the Vault secret containing your secrets              |
| VAULT_AWS_AUTH_PATH | The path to the mount on your Vault instance that handles IAM authentication |

The Vault secret should contain the following set of key-values - the values are examples and should be replaced as contextually appropriate:

```
okta_host=https://example.okta.com
okta_token={API TOKEN}
```

`{API_TOKEN}` must be replaced with an API token for Okta that has the `okta.apps.read` scope.

#### Environment Variables

We advise against using environment variables for secrets in AWS Lambda as they are persisted in plaintext. As such, your Okta API token may be leaked. If you would prefer to use environment variables, however, you must provide the following environment variables to your Lambda configuration:

| Variable          | Purpose                                                           |
| ----------------- | ----------------------------------------------------------------- |
| OKTA_HOST | The hostname of your Okta instance. We'd recommend using a vanity domain, such as https://singlesignon.example.com. |
| OKTA_TOKEN | A token from Okta that has the `okta.apps.read` scope. |
| SETTINGS_PROVIDER | This must be set to 'env' for the Lambda functions to read from the environment. |

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
