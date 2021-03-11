import React from "react";
import { Card } from "semantic-ui-react";

const TroubleshootingCard = () => {
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
};

export default TroubleshootingCard;
