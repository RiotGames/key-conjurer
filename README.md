# KeyConjurer

![Key Conjurer Champion](docs/champion.png)

KeyConjurer is a project designed to get rid of permanent AWS credentials.

KeyConjurer is made of three parts:

- [webserver](./webserver/) - Lambda functions used by the CLI to gather data on
  protected resources.
- [cli](./cli/) - The CLI interface.
- [frontend](./frontend/) - A static webpage which informs users on how to
  download and use KeyConjurer.

KeyConjurer is designed to work with Okta as an IdP, supports AWS applications,
and is inspired in part by [okta-aws-cli](https://github.com/okta/okta-aws-cli).
The main difference from okta-aws-cli is that KeyConjurer does not require all
users to have access to the Okta administration API - Instead, we use a Lambda
function to access the protected resources required.

We use KeyConjurer a lot at Riot, but we can't guarantee any external support
for this project. It's use at your own risk. If you encounter a bug or have a
feature request, please feel free to raise a pull request or an issue against
this repository. You're also welcome to fork the code and modify it as you see
fit.

## Dependencies

- go 1.20+
- node 16.17.0+

### Administration

#### Okta setup

In order to use KeyConjurer, an Okta administrator must configure their tenant
appropriately:

- A new _native_ OIDC application must be created within your Okta tenant, and
  the following settings must be configured:
  - Scopes: `profile openid okta.apps.read`
  - Authorization Types: Hybrid Flow, Authorization Code, Token Exchange
  - Redirection URI: http://localhost:57468
  - We recommend you enable Federated Mode on this native application so that
    users don't need to be explicitly assigned to it.
- All AWS applications must have their Allowed Web SSO Client set to the _Client
  ID_ of the native OIDC application that was created. This can be configured by
  going to the Sign On tab for each individual Okta application or managing the
  application configuration in an IAC provider, like Terraform.

Okta configuration should be configured _out of band_ and is not provided in
this repository.

#### Lambda functions settings

A single lambda function is used to filter applications within the organization
to just the ones the user has access to. This function is required because
enumerating applications within Okta's API is currently considered an
administrative action, and as such, using a users access token to perform this
action requires the user to be an administrator on the Okta tenant.

The Lambda function is deployed as a Docker container. It's up to you to decide
how to launch the Docker container, but you'll need to specify two values:

| Flag                                      | Purpose                                                                                                                                                                                                                                                                                                |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `--okta-host`                             | The hostname of your Okta instance. This may also be set via `KEYCONJURER_OKTA_HOST`.                                                                                                                                                                                                                  |
| `--okta-token` **or** `--okta-token-file` | An API token for your Okta instance. This must have the `okta.apps.read` scope. You may set `--okta-token-file` instead of `--okta-token` if you're supplying secrets to the container via a volume. This may also be set via `KEYCONJURER_OKTA_TOKEN` and `KEYCONJURER_OKTA_TOKEN_FILE` respectively. |
