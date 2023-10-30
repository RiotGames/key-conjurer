import React from "react";
import { Tab, TabPane } from "semantic-ui-react"

export const Usage = () => {
    const panes = [
        { menuItem: "Mac", render: () => <TabPane>Mac</TabPane> },
        { menuItem: "Windows", render: () => <TabPane>Windows</TabPane> },
        { menuItem: "WSL", render: () => <TabPane>WSL</TabPane> },
        { menuItem: "Linux", render: () => <TabPane>Linux</TabPane> },
    ]

    return <Tab panes={panes} />
}
