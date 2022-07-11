# Key Conjurer CLI
======
Go-based CLI interface for Key Conjurer API. This tool
enables you to retrieve temporary AWS API credentials. It can be used in automation.

## Important Changes
* Account names are now case insensitive. `keyconjurer get foo` and `keyconjurer get Foo` are equivalent calls.
* In an effort to homogonize our MFA clients, this version of keyconjurer only supports users of Duo
devices. OTP codes are no longer accepted.

## Installation
### How To Get keyconjurer

* Download the lastest release from the Key Conjurer frontend.
* Once downloaded, place the binary in a directory that's in your path (e.g. `/usr/local/bin` on MAC/Linux)

## QuickStart
Add `/usr/local/bin` to your path if it's not already.
```sh
$ keyconjurer login
username: <username>
password: <password>
Login Successful
$ keyconjurer accounts
123123  Foo
321321  Bar
...
$ keyconjurer get Foo
Sending Duo Push  # This line in printed to stderr so the user gets feedback
export AWS_ACCESS_KEY_ID=ASIAI...
export AWS_SECRET_ACCESS_KEY=72qKx...
export AWS_SESSION_TOKEN=FQoDY...
export AWS_SECURITY_TOKEN=FQoDY...
export TF_VAR_access_key=$AWS_ACCESS_KEY_ID
export TF_VAR_secret_key=$AWS_SECRET_ACCESS_KEY
export TF_VAR_token=$AWS_SESSION_TOKEN
export AWSKEY_ACCOUNT=123123
export AWSKEY_EXPIRATION=2018-02-22T18:56:26+00:00
$ $(keyconjurer get Foo) # Runs the export commands instead of printing them
```
The `Sending Duo Push` is printed to stderr so it will not interfere with stdout rediction and evaluation.

## Detailed Info
### Exported Variables
keyconjurer uses the following environment variables:

| Variable                | Use                                               |
|-------------------------|---------------------------------------------------|
| `AWS_ACCESS_KEY_ID`     | Standard AWS env variable                         |
| `AWS_SECRET_ACCESS_KEY` | Standard AWS env variable                         |
| `AWS_SESSION_TOKEN`     | Standard AWS env variable                         |
| `AWS_SECURITY_TOKEN`    | Deprecated AWS env variable used by some software |
| `TF_VAR_access_key`     | Teraform env variable                             |
| `TF_VAR_secret_key`     | Teraform env variable                             |
| `TF_VAR_token`          | Teraform env variable                             |
| `AWSKEY_ACCOUNT`        | Used by keyconjurer --time-remaining flag          |
| `AWSKEY_EXPIRATION`     | Used by keyconjurer --time-remaining flag          |

### Help
keyconjurer has a help menu built into it.
```sh
$ keyconjurer --help
keyconjurer retrieves temporary credentials from the Key Conjurer service.

To get started run the following commands:
keyconjurer login # You will get prompted for your AD credentials
keyconjurer accounts
keyconjurer get <accountName>

Usage:
  keyconjurer [command]

Available Commands:
  accounts    Prints the list of accounts you have access to.
  alias       Give an account a nickname.
  devices     Prints the list of accounts you have access to.
  get         Retrieves temporary AWS API credentials.
  help        Help about any command
  login       Get credentials for Key Conjurer
  set         Sets config values.
  unalias     Remove alias from account.

Flags:
      --keyconjurer-rc-path string   path to .keyconjurerrc file (default "~/.keyconjurerrc")
  -h, --help                    help for keyconjurer
      --version                 version for keyconjurer

Use "keyconjurer [command] --help" for more information about a command.
```

You can get individual command help with `keyconjurer <command> --help`.
```sh
$ keyconjurer get --help
Retrieves temporary AWS API credentials for the specified account.  It sends a push request to the first Duo device it finds associated with your account.

Usage:
  keyconjurer get <accountName/alias> [flags]

Examples:
keyconjurer get <accountName/alias>

Flags:
  -c, --creds-prompt          Prompt for username and password through stdin. Can be piped in using the following format "<username>\n<pasword>\n".
  -h, --help                  help for get
  -t, --time-remaining uint   Request new keys if there are no keys in the environment or the current keys expire within <time-remaining> minutes. Defaults to 60. (default 60)
      --ttl uint              The key timeout in hours from 1 to 8. (default 1)

Global Flags:
      --keyconjurer-rc-path string   path to .keyconjurerrc file (default "~/.keyconjurerrc")
```

### Login
```sh
keyconjurer login
```

This will prompt you for your active directory credentials, which will be stored
in an encrypted token in the file selected by `--keyconjurer-rc-path` which defaults to `~/.keyconjurerrc`.

#### Reasons to run this command
* First time setup
* You changed your AD creds
* You got access to a new AWS account
  * Authorization for AWS account access is handled by OneLogin

### Retrieving API Keys
To retrieve API keys for a given account run:
```sh
$ keyconjurer get <accountName/alias>
```
keyconjurer will automatically look for a Duo device to send a push request.

If keyconjurer cannot find a Duo device, it will default to the first device it
finds. To retrieve a API keys in this instance run:
```sh
$ keyconjurer get <accountName/alias>
```
Both the above commands will print the output directly to `stdout`.

#### Easy Exports
keyconjurer is unable to directly export variables into the calling shell. Children
processes cannot manipulate parent process' environment.  To get around this
you can wrap the `keyconjurer get` command in `$()` to immediately evaluate the result:
```
$ $(keyconjurer get <accountName/alias>)
echo $AWS_ACCESS_KEY_ID
ASIA...
```

#### --ttl
The ttl flag is used to retrieve multi hour tokens. Currently this supports
up to 8 hours. This only accepts integers.
```sh
$ keyconjurer get <accountName/alias> --ttl N
```

#### -t, --time-remaining
A common usecase in automation is requesting new keys only when the current set is
expiring within `N` minutes. This can be achieved by running:
```sh
$ keyconjurer get <accountName/alias> --time-remaining <N>
# or
$ keyconjurer get <accountName/alias> -t <N>
```
By default `N=5` so that any call will always request new keys if there is less than 5 minutes left.

keyconjurer checks to see if the account you are requesting keys for is not the account
you currently have keys for. If the accounts differ, new API keys are requested and a
push request is sent.

If the account you are requesting keys for is the account you currently have keys there
are typically four possible situations when making this call:

| Situations                     | `AWSKEY_ACCOUNT` not set | `AWSKEY_ACCOUNT` set |
|:------------------------------:|--------------------------|----------------------|
| `AWSKEY_EXPIRATION` delta <= N | Push request sent        | Push request Sent    |
| `AWSKEY_EXPIRATION` delta >= N | Push request sent        | Nothing happens      |


#### -c, --creds-prompt
Another usecase in automation is passing in credentials through `stdin`.  This can be
achieved by running:
```sh
$ keyconjurer get <accountName/alias> --creds-prompt
username: <username>
password: <password>
# or
$ keyconjurer get <accountName/alias> -c
username: <username>
password: <password>
```

You can pipe in the username and password in as long as its formatted `<username>\n<password\n`:
```sh
$ printf "username\npassword\n" | keyconjurer get -c Foo
```

### Accounts
To list all the accounts you have access to run:
```sh
$ keyconjurer accounts
```

Given you have valid credentials, to update your accounts run:
```sh
$ keyconjurer accounts --update
```

Some accounts may have complex names which are a pain to type.  You can give the account
an alias to refer to in the future:
```sh
$ keyconjurer alias <accountName> <alias>
```

You can unalias any account by executing:
```sh
$ keyconjurer unalias <accountName/alias>
```

### Setting defaults
Use the `keyconjurer set` command to permanent set the `ttl` or `time-remaining`.
```
$ keyconjurer set ttl 8
$ keyconjurer set time-remaining 120
```

To see all available options run `$ keyconjurer set --help`
