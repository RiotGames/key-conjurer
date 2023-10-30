import * as marked from 'marked';
import React from "react";
import { Tab, TabPane } from "semantic-ui-react";
import MacUsageDocument from '../articles/MacUsage.md';

const macUsageDocument = marked.parse(MacUsageDocument);

export const Usage = () => {
    const panes = [
        { menuItem: "Mac", render: () => <TabPane dangerouslySetInnerHTML={{ __html: macUsageDocument }} />},
        { menuItem: "Windows", render: () => <TabPane>Windows</TabPane> },
        { menuItem: "WSL", render: () => <TabPane>WSL</TabPane> },
        { menuItem: "Linux", render: () => <TabPane>Linux</TabPane> },
    ]

    return <Tab panes={panes} />
}
