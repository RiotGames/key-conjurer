import React from "react";
import { Card, Image, Divider } from "semantic-ui-react";
import { binaryName } from "./../consts";
import keyConjurerLogo from "./../images/KeyConjurer.png";

const History = () => (
  <Card fluid>
    <Image src={keyConjurerLogo} />
    <Card.Content>
      <Card.Header>History</Card.Header>
      <Divider />
      <Card.Content>
        Publishing AWS API keys publicly (e.g. to Github) is a significant
        security risk to Riot and our players. On several occasions, Rioters
        have unfortunately done this and these leaked keys have been used to
        modify AWS infrastructures, though the worst case of having player
        data compromised has thankfully not been realised.
      </Card.Content>
      <Divider />
      <Card.Content>
        This service provides temporary AWS API keys. Log in to retrieve a
        list of AWS accounts available to you.
      </Card.Content>
      <Divider />
      <Card.Content>
        If you prefer the cli, we have you covered. Just download one of the
        following and move it into your $PATH:
        <br />
        <br />
        <a href={`/${binaryName}-darwin-amd64`}>
          {`${binaryName}`}-darwin (osx)
        </a>
        <br />
        <a href={`${binaryName}-darwin-arm64`}>
          {`${binaryName}`}-darwin (osx M1/M2)
        </a>
        <br />
        <a href={`${binaryName}-linux`}>
          {`${binaryName}`}-linux
        </a>
        <br />
        <a href={`${binaryName}-windows.exe`}>
          {`${binaryName}`}-windows
        </a>
      </Card.Content>
    </Card.Content>
  </Card>
);

export default History;
