import React from "react";
import { Card, Image, Divider } from "semantic-ui-react";
import keyConjurerLogo from "./../images/KeyConjurer.png";

const binaryName = process.env.REACT_APP_BINARY_NAME;


export const DownloadLinks = () => {
  return <>
    <a href={`${binaryName}-linux-amd64`}>
      {`${binaryName}`}-linux AMD64 (This is probably the one you want)
    </a>
    <br />
    <a href={`${binaryName}-linux-arm64`}>{`${binaryName}`}-linux ARM64</a>
    <br />
    <a href={`${binaryName}-windows.exe`}>{`${binaryName}`}-windows</a>
  </>;
}

export const History = () => (
  <Card fluid>
    <Image src={keyConjurerLogo} />
    <Card.Content>
      <Card.Header>History</Card.Header>
      <Divider />
      <Card.Content>
        Publishing AWS API keys publicly (e.g. to Github) is a significant
        security risk to Riot and our players. On several occasions, Rioters
        have unfortunately done this and these leaked keys have been used to
        modify AWS infrastructures, though the worst case of having player data
        compromised has thankfully not been realised.
      </Card.Content>
      {/* <Divider />
      <Card.Content>
        <DownloadLinks />
      </Card.Content> */}
    </Card.Content>
  </Card>
);
