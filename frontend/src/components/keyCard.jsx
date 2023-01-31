import React, { Component } from "react";
import { Card, Button, Segment } from "semantic-ui-react";
import { CopyToClipboard } from "react-copy-to-clipboard";

import { subscribe } from "./../stores";

class KeyCard extends Component {
  state = {
    accessKeyId: "",
    secretAccessKey: "",
    sessionToken: "",
    subscribed: false,
  };

  constructor(props) {
    super(props);
    subscribe("awsKeys", (value) => {
      this.setState({
        ...this.state,
        ...value,
      });
    });
  }

  render() {
    const { accessKeyId, secretAccessKey, sessionToken, expiration } =
      this.state;

    const keys = `export AWS_ACCESS_KEY_ID=${accessKeyId}
export AWS_SECRET_ACCESS_KEY=${secretAccessKey}
export AWS_SESSION_TOKEN=${sessionToken}
export AWS_SECURITY_TOKEN=$AWS_SESSION_TOKEN`;

    return accessKeyId ? (
      <Card fluid>
        <Card.Content>
          <Card.Header>API Keys</Card.Header>
          <Card.Meta>Expires at: {expiration}</Card.Meta>
          <Segment>
            <Card.Description>
              <pre style={{ overflowX: "scroll" }}>{`${keys}`}</pre>
            </Card.Description>
          </Segment>
          <CopyToClipboard text={keys}>
            <Button fluid primary>
              Click to Copy
            </Button>
          </CopyToClipboard>
        </Card.Content>
      </Card>
    ) : (
      ""
    );
  }
}

export default KeyCard;
