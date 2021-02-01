import React, { Component } from "react";
import { Card, Image, Divider } from "semantic-ui-react";
import { keyConjurerDomain, binaryName } from "./../consts";

import keyConjurerLogo from "./../images/KeyConjurer.png";

class History extends Component {
  render() {
    // The download links will not use the proper file name when in dev
    //  only works in prod.
    return (
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
            <a
              href={`${keyConjurerDomain}/${binaryName}-darwin`}
              download={`${binaryName}`}
            >
              {`${binaryName}`}-darwin (osx)
            </a>
            <br />
            <a
              href={`${keyConjurerDomain}/${binaryName}-linux`}
              download={`${binaryName}`}
            >
              {`${binaryName}`}-linux
            </a>
            <br />
            <a
              href={`${keyConjurerDomain}/${binaryName}-windows.exe`}
              download={`${binaryName}.exe`}
            >
              {`${binaryName}`}-windows
            </a>
          </Card.Content>
        </Card.Content>
      </Card>
    );
  }
}

export default History;
