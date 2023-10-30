import React from "react";
import { Tab, TabPane } from "semantic-ui-react"

const binaryName = process.env.REACT_APP_BINARY_NAME;

export const Usage = () => {
    const panes = [
        { menuItem: "Mac", render: () => <TabPane><MacUsage /></TabPane> },
        { menuItem: "Windows", render: () => <TabPane>Windows</TabPane> },
        { menuItem: "WSL", render: () => <TabPane>WSL</TabPane> },
        { menuItem: "Linux", render: () => <TabPane>Linux</TabPane> },
    ]

    return <Tab panes={panes} />
}


const MacUsage = () => {
    return <>
        <h1>Download</h1>
        <p>
            Download the latest release from one of the following two links, depending on the chip of your Mac.
        </p>
        <p>If you're unsure what chip your Mac has, click the Apple button in the top left of your laptop screen and then click <em>About this Mac</em>. In the window that pops up, there will be a line that indicates whether your Mac has a M1, M2 or Intel chip.</p>
        <ul>
            <li>Apple Silicon: <a href={`/${binaryName}-darwin-arm64`}>{`${binaryName}`}-darwin-arm64</a></li>
            <li>Intel: <a href={`/${binaryName}-darwin-amd64`}>{`${binaryName}`}-darwin</a></li>
        </ul>

        <h1>Installation</h1>
        <p>
            Open a terminal and execute the following commands. <br />
            <code>
            $ mkdir -p ~/.bin/ <br />
            $ mv ~/Downloads/keyconjurer-* ~/.bin/keyconjurer <br />
            $ chmod +x ~/.bin/keyconjurer
            </code>
        </p>
        <p>
            Next, add <code>~/.bin</code> to your <code>`$PATH`</code>. With the Zsh shell (pre-installed with most Macs), this can be accomplished in the Terminal with:<br />
            <code>
                $ echo 'export PATH="$PATH:~/.bin"' &gt;&gt; ~/.zshrc
            </code><br />
            Restart your shell.
        </p>

        <h1>Daily Usage</h1>
        <p>
            Open a terminal and execute the following command to see the help menu:<br />
            <code>
                $ keyconjurer --help
            </code> <br />
        </p>

        <p>
            To use KeyConjurer, you need to log in using your YubiKey and Okta credentials. To do this, execute the following command:<br />
            <code>
                $ keyconjurer login
            </code>
        </p>
        <p>
            After you log in, you can generate credentials using:
            <br />
            <code>
                $ keyconjurer get [account-name] --role [role-name]
            </code>
        </p>
        <p>
            The accounts you have access to can be retrieved by using: <br />
            <code>
                $ keyconjurer accounts
            </code>
        </p>
        <p>
            The roles you have access to can be retrieved by using: <br />
            <code>
                $ keyconjurer roles [account-name]
            </code>
        </p>
        <p>
            KeyConjurer will function anywhere you have access to Okta. <b>You do not need to be on a specific VPN</b>. You may be required to have access to your YubiKey to access KeyConjurer.
        </p>

        <h1>Troubleshooting</h1>
        <h2>I get a security warning when I try to run KeyConjurer</h2>
        <p>
            Please follow the instructions to add an exception to your Mac security policy in this <a href="https://support.apple.com/guide/mac-help/open-a-mac-app-from-an-unidentified-developer-mh40616/mac">Apple support article</a>.

            This is a known bug. We do not have the facility to sign Mac binaries, so the binary we ship is unsigned and Mac will, by default, try to prevent you from running it.
        </p>
    </>
}