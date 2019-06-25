import React, { Component } from 'react';
import {Grid} from 'semantic-ui-react';

import Header from './components/header';
import History from './components/history';
import KeyCard from './components/keyCard';
import KeyRequestForm from './components/keyRequestForm';
import LoginForm from './components/loginForm';
import TroubleshootingCard from './components/troubleshootingCard';

class App extends Component {
    render() {
	return (
            <div>
              <Header />
              <Grid>
                <Grid.Row />
                <Grid.Row columns={4}>
                  <Grid.Column width={2}>
                  </Grid.Column>
                  <Grid.Column width={4}>
                    <History />
                  </Grid.Column>
                  <Grid.Column width={8}>
                    <LoginForm />
                    <KeyRequestForm />
                    <KeyCard />
                    <TroubleshootingCard />
                  </Grid.Column>
                  <Grid.Column width={2}/>
                </Grid.Row>
              </Grid>
            </div>
	);
    }
}

export default App;
