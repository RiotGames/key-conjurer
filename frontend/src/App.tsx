import * as marked from "marked";
import React from "react";
import { Card, Image, Menu, Tab, TabPane } from "semantic-ui-react";
import keyConjurerLogo from "./images/KeyConjurer.png";
import styles from "./App.module.css";
import MacUsageDocument from "./articles/MacUsage.md";
import LinuxUsageDocument from "./articles/LinuxUsage.md";
import WindowsUsageDocument from "./articles/WindowsUsage.md";
import WSLUsageDocument from "./articles/WSLUsage.md";

const macUsageDocument = marked.parse(MacUsageDocument);
const linuxUsageDocument = marked.parse(LinuxUsageDocument);
const windowsUsageDocument = marked.parse(WindowsUsageDocument);
const wslUsageDocument = marked.parse(WSLUsageDocument);

export const App = () => (
  <>
    <Header />

    <div className={styles.Content}>
      <p className={styles.Para1}>
        KeyConjurer is an application for generating temporary session
        credentials for AWS and Tencent Cloud.
      </p>
      <div className={styles.History}>
        <History />
      </div>
      <div className={styles.Usage}>
        <Usage />
      </div>
    </div>
  </>
);

const Usage = () => {
  const panes = [
    {
      menuItem: "Mac",
      render: () => (
        <TabPane dangerouslySetInnerHTML={{ __html: macUsageDocument }} />
      ),
    },
    {
      menuItem: "Windows",
      render: () => (
        <TabPane
          dangerouslySetInnerHTML={{ __html: windowsUsageDocument }}
        />
      ),
    },
    {
      menuItem: "WSL",
      render: () => (
        <TabPane dangerouslySetInnerHTML={{ __html: wslUsageDocument }} />
      ),
    },
    {
      menuItem: "Linux",
      render: () => (
        <TabPane dangerouslySetInnerHTML={{ __html: linuxUsageDocument }} />
      ),
    },
  ];

  return <Tab panes={panes} />;
};

const History = () => (
  <Card fluid>
    <Image src={keyConjurerLogo} />
    <Card.Content>
      <Card.Header>History</Card.Header>
      <Card.Content>
        Publishing AWS API keys publicly (e.g. to Github) is a significant
        security risk to Riot and our players. On several occasions, Rioters
        have unfortunately done this and these leaked keys have been used to
        modify AWS infrastructures, though the worst case of having player data
        compromised has thankfully not been realised.
      </Card.Content>
    </Card.Content>
  </Card>
);

const appVersion = process.env.REACT_APP_VERSION;
const Header = () => (
  <Menu fluid color="grey">
    <Menu.Item header>Key Conjurer</Menu.Item>
    {appVersion && <Menu.Item position="right">{appVersion}</Menu.Item>}
  </Menu>
);
