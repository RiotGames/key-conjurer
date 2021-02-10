import React, { Component } from "react";
import { Card } from "semantic-ui-react";

class TroubleshootingCard extends Component {
  render() {
    return (
      <Card fluid>
        <Card.Content>
          <Card.Meta>Troubleshooting</Card.Meta>
          <Card.Description>
            This app requires Javascript execution
          </Card.Description>
          <Card.Description>Disable adblockers for this site</Card.Description>
        </Card.Content>
      </Card>
    );
  }
}

export default TroubleshootingCard;
