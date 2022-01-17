# Key Conjurer API
Key Conjurer provides temporary AWS API credentials! Key Conjurer uses OneLogin and Duo to retrieve STS tokens.

## Requirements
Go 1.13.4+

# Deployment

The deployment of the Key Conjurer API is taken care of by the main makefile and terraform files.

## API Deployment Requirements
1. Private API Gateway endpoint in your VPC
1. A version enabled S3 bucket
1. Valid AWS credentials in environment or credentials file
1. A KMS key (accessible via Lambda)


## Lambda Env Settings
### Environment Variables
| Variable          | Purpose                                                           |
|-------------------|-------------------------------------------------------------------|
| EncryptedSettings | A KMS encrypted json blob with settings (See below for more info) |


### Encrypted Settings
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

# Testing
## Automated Testing
The automated tests require the following conditions:

1. Valid AWS credentials set in environment or in `~/.aws/config`
1. The following environment variables set

| Variable   | Purpose                           |
|------------|-----------------------------------|
| AWS_REGION | The region where KeyConjurer lives|
| KMS_KEY_ID | The KMS key to use during testing |

From the root directory run:
```
cd ./keyconjurer
go test -v
```

## Manual Testing

### get_user_data
**Event Example**
```
{
  "username": "<your username>",
  "password": "<your password>",
  "client": "lambdaEvent",
  "clientVersion": "lambda",
  "shouldEncryptCreds": "true" || "false"
}
```

1. Visit the lambda.
2. Click on the `Select a test event...` field. Then select `Configure test events`.
3. Set your test event data to look like the example above. Then select `Save`.
4. Click the `Test` button.
5. View the results. A positive result will look like the response below.

**Positive Result Example**
```
{
  "Success": true,
  "Message": "success",
  "Data": {
    "devices": [
      {
        "deviceId": "123456",
        "deviceType": "A device type"
      },
      ...
    ],
    "apps": [
      {
        "id": 666666,
        "name": "AWS - account1"
      },
      {
        "id": 666667,
        "name": "AWS - account2"
      },
      ...
    ],
    "creds": "Encrypted Credentials"
  }
}
```

### get_aws_creds
**Event Example**
```
{
  "username": "<your username>",
  "password": "<your password>",
  "client": "lambdaEvent",
  "clientVersion": "lambda",
  "appId": "<app id>",
  "timeoutInHours": "1"
}
```
The `<app id>` is the AWS account that will have keys created.

1. Visit the lambda.
1. Click on the `Select a test event...` field. Then select `Configure test events`.
1. Set your test event data to look like the example above. Then select `Save`.
1. Click the `Test` button.
1. Accept the push request sent to your phone.
1. View the results. A positive result will look like the response below.

**Positive Result Example**
```
{
  "Success": true,
  "Message": "success",
  "Data": {
    "accessKeyId": "A key id",
    "secretAccessKey": "A secret key",
    "sessionToken": "A session token",
    "expiration": "2018-08-09T10:25:13Z"
  }
}
```


