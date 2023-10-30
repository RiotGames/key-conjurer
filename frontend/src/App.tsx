import React from "react";
import { Grid } from "semantic-ui-react";
import { Header } from "./components/Header";
import { History } from "./components/History";
import { Usage } from "./components/Usage";

export const App = () => {
  return (
    <div>
      <Header />
      <Grid>
        <Grid.Row />
        <Grid.Row columns={3} centered>
          <Grid.Column>
            <History />
          </Grid.Column>
          <Grid.Column>
            <Usage />
          </Grid.Column>
        </Grid.Row>
      </Grid>
    </div>
  );
};
