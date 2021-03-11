import React from "react";
import { Card, FormCheckbox, Grid } from "semantic-ui-react";
import Header from "./components/header";
import History from "./components/history";
import KeyCard from "./components/keyCard";
import KeyRequestForm from "./components/keyRequestForm";
import LoginForm from "./components/loginForm";
import TroubleshootingCard from "./components/troubleshootingCard";
import { updateUserInfo } from "./actions";

const App = () => {
  const [onelogin, setOnelogin] = React.useState(false);
  const idp = onelogin ? "onelogin" : "okta";

  const handleIDPChange = (_event, { checked }) => {
    setOnelogin(checked);
    localStorage.setItem("provider", checked ? "onelogin" : "okta");
    // Changing IDP means we have to invalidate all of our credentials.
    updateUserInfo({
      username: "",
      password: "",
    });
  };

  React.useEffect(() => {
    if (localStorage.getItem("provider") === "onelogin") {
      setOnelogin(true);
    }
  }, []);

  return (
    <div>
      <Header />
      <Grid>
        <Grid.Row />
        <Grid.Row columns={4}>
          <Grid.Column width={2}></Grid.Column>
          <Grid.Column width={4}>
            <History />
          </Grid.Column>
          <Grid.Column width={8}>
            <Card fluid>
              <Card.Content>
                <FormCheckbox
                  toggle
                  label="I want to use OneLogin and I understand that this option will go away by May 2021."
                  checked={onelogin}
                  onChange={handleIDPChange}
                />
              </Card.Content>
            </Card>
            <LoginForm idp={idp} />
            <KeyRequestForm idp={idp} />
            <KeyCard />
            <TroubleshootingCard />
          </Grid.Column>
          <Grid.Column width={2} />
        </Grid.Row>
      </Grid>
    </div>
  );
};

export default App;
