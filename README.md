# KeyConjurer

![Key Conjurer Champion](docs/champion.png)

KeyConjurer is a project designed to get rid of permanent AWS credentials.

KeyConjurer is made of three parts:

- [lambda](./lambda/) - Lambda functions used by the CLI to gather data on protected resources.
- [cli](./cli/) - The CLI interface.
- [frontend](./frontend/) - A static webpage which informs users on how to download and use KeyConjurer.

KeyConjurer is designed to work with Okta as an IdP, supports AWS applications, and is inspired in part by [okta-aws-cli](https://github.com/okta/okta-aws-cli). The main difference from okta-aws-cli is that KeyConjurer does not require all users to have access to the Okta administration API - Instead, we use a Lambda function to access the protected resources required.

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
* All AWS applications must have their Allowed Web SSO Client set to the _Client ID_ of the native OIDC application that was created. This can be configured by going to the Sign On tab for each individual Okta application or managing the application configuration in an IAC provider, like Terraform.

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
