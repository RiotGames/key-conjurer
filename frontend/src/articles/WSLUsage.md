**WSL support is experimental**. It may or may not work. This is not being actively developed because we don't use WSL but any issues or feedback will be addressed.

# Download

You can run `.exe` files within WSL, and you may also find that the Linux version works.

* Windows (x64) [keyconjurer-windows.exe](./keyconjurer-windows.exe)
* Linux (x64) [keyconjurer-linux-amd64](./keyconjurer-linux-amd64)

Add either one of these to your `$PATH`, and you should be able to run them just fine.

**Please note**: The Windows version of the binary will store your configuration files in `%USERPROFILE%` when run from Windows, and `$HOME` when run in WSL. WSL cannot read the Windows configuration and visa versa.

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

## Login Flow in WSL

When executing `keyconjurer login` in WSL for the first time, that requires an oauth flow with your browser, you might get the following error:

```
exec: "xdg-open,x-www-browser,www-browser,wslview": executable file not found in $PATH
```

It requires some dependencies to be installed:

```
sudo add-apt-repository ppa:wslutilities/wslu
sudo apt update
sudo apt install wslu
```

After that, you are able to login using the browser.
