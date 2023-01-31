import React from "react";
import { Grid } from "semantic-ui-react";
import Header from "./components/header";
import History from "./components/history";
import KeyCard from "./components/keyCard";
import KeyRequestForm from "./components/keyRequestForm";
import LoginForm from "./components/loginForm";
import TroubleshootingCard from "./components/troubleshootingCard";
import { updateUserInfo } from "./actions";

const App = () => {
  React.useEffect(() => {
    if (localStorage.    getItem("provider") === "onelogin") {
      // Force a user to log out if they are using OneLogin as their provider
      updateUserInfo({ username: "", password: "" });
      localStorage.removeItem("provider");
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
            <LoginForm idp="okta" />
            <KeyRequestForm idp="okta" />
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
