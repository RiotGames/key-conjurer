import React from "react";
import { Grid } from "semantic-ui-react";
import Header from "./components/header";
import History from "./components/history";
import KeyCard from "./components/KeyCard";
import KeyRequestForm from "./components/KeyRequestForm";
import LoginForm from "./components/loginForm";
import TroubleshootingCard from "./components/troubleshootingCard";
import { updateUserInfo } from "./actions";

const App = () => {
  React.useEffect(() => {
    // We used to support OneLogin as a provider; this ensures it isn't kept around.
    if (localStorage.getItem("provider") !== "okta") {
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
            <LoginForm />
            <KeyRequestForm />
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
