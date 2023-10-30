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

        <h1>Usage</h1>
    </>
}