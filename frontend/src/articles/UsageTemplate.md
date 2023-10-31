# Download

TBD

# Installation

TBD

# Daily Usage

Open a terminal and execute the following command to see the help menu:

    $ keyconjurer --help

To use KeyConjurer, you need to log in using your YubiKey and Okta credentials. To do this, execute the following command:

    $ keyconjurer login

After you log in, you can generate temporary cloud credentials using:

    $ keyconjurer get [account-name] --role [role-name]

The accounts you have access to can be retrieved by using:

    $ keyconjurer accounts

The roles you have access to can be retrieved by using:

    $ keyconjurer roles [account-name]

KeyConjurer will function anywhere you have access to Okta. **You do not need to be on a specific VPN**. You may be required to have access to your YubiKey to access KeyConjurer.

# Troubleshooting

TBD
