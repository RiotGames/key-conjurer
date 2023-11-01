# Download
Download the latest release from one of the following two links, depending on the chip of your Mac.


If you're unsure what chip your Mac has, click the Apple button in the top left of your laptop screen and then click <em>About this Mac</em>. In the window that pops up, there will be a line that indicates whether your Mac has a M1, M2 or Intel chip.

* Apple Silicon: [keyconjurer-darwin-arm64](./keyconjurer-darwin-arm64)
* Intel: [keyconjurer-darwin-amd64](./keyconjurer-darwin-amd64)

# Installation

Open a terminal and execute the following commands.

    $ mkdir -p ~/.bin/
    $ mv ~/Downloads/keyconjurer-* ~/.bin/keyconjurer
    $ chmod +x ~/.bin/keyconjurer

Next, add `~/.bin` to your `$PATH`. With the Zsh shell (pre-installed with most Macs), this can be accomplished in the Terminal with:

    $ echo 'export PATH="$PATH:$HOME/.bin"' >> ~/.zshrc

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

# Troubleshooting

## I get a security warning when I try to run KeyConjurer

Please follow the instructions to add an exception to your Mac security policy in this [Apple support article][apple-support-article].

This is a known bug. We do not have the facility to sign Mac binaries, so the binary we ship is unsigned and Mac will, by default, try to prevent you from running it.

[apple-support-article]: https://support.apple.com/guide/mac-help/open-a-mac-app-from-an-unidentified-developer-mh40616/mac
