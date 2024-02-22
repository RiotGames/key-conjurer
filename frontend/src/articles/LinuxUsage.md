# Download

Download the latest release from one of the following two links, depending on your CPU architecture.

As Linux is a more advanced operating system, we assume that you're familiar with your particular distribution and how to find out which architecture you're on.

* Intel: [keyconjurer-linux-amd64](./keyconjurer-linux-amd64)
* ARM64: [keyconjurer-linux-arm64](./keyconjurer-linux-arm64)

# Installation

Open a terminal and execute the following commands.

    $ mkdir -p ~/.bin/
    $ mv ~/Downloads/keyconjurer-* ~/.bin/keyconjurer
    $ chmod +x ~/.bin/keyconjurer

Next, add `~/.bin` to your `$PATH`. With Bash, this can be accomplished in the Terminal with:

    $ echo "export PATH="$PATH:$HOME/.bin"' >> ~/.bashrc

Restart your shell.

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
