import React from "react";
import { Grid } from "semantic-ui-react";
import { Header } from "./components/Header";
import { History } from "./components/History";

export const App = () => {
  return (
    <div>
      <Header />
      <Grid>
        <Grid.Row />
        <Grid.Row columns={4} centered>
          <Grid.Column>
            <History />
          </Grid.Column>
        </Grid.Row>
      </Grid>
    </div>
  );
};
