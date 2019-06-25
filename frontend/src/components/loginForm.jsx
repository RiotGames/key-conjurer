import React, { Component } from 'react';
import { Form, Card, Message } from 'semantic-ui-react';

import { authenticate, updateUserInfo } from './../actions';
import { subscribe } from './../stores';


class LoginForm extends Component {
    state = {
	username: '',
	password: '',
	requestSent: false,
	loginRequest: false,
	encryptCreds: false,
	error: false,
        errorEvent: '',
	errorMessage: ''
    };

    handleTextChange =
	property =>
	(event, data) => {
	    const {username, password} = this.state;
	    updateUserInfo({
		username: ('username' === property) ? data.value : username,
		password: ('password' === property) ? data.value : password
	    });
	};

    handleCheckChange =
	name =>
	event => {
	    this.setState({encryptCreds: !this.state.encryptCreds});
	};


    authUser =
	event => {
	    const {
		username,
		password,
		encryptCreds
	    } = this.state;
	    this.setState({
		loginRequest:true,
		error: false,
		errorMessage: ''
	    }, _ => authenticate(username, password, encryptCreds));
	}

    componentDidMount() {
	if (localStorage.creds) {
	    this.setState({
		autoAuth: true,
		encryptCreds: true
	    }, _ => updateUserInfo({
		username: 'encrypted',
		password: localStorage.creds
	    }));

	}

	subscribe('userInfo', ({username, password}) => {
	    this.setState({
		username,
		password
	    }, _ => {
		if (this.state.autoAuth) {
		    this.setState({
			autoAuth: false
		    }, _ => this.authUser());
		}
	    });
	});
	subscribe('request', ({requestSent}) => this.setState({
	    requestSent,
	    loginRequest: (requestSent && this.state.loginRequest)
	}));
	subscribe('errors', ({error, event, message: errorMessage}) => {
	    if (event === 'login') {
		this.setState({error, errorMessage, errorEvent:event});
	    }
	});
    }

    render() {
	const {
	    username,
	    password,
	    requestSent,
	    encryptCreds,
	    error,
            errorEvent,
	    errorMessage
	} = this.state;

	return (
	    <Card fluid>
              <Card.Content>
	        <Card.Header>
	          Auth
	        </Card.Header>
                <Card.Description>
                  <Form loading={requestSent}>
                    <Form.Group widths='equal' inline>
                      <Form.Input
                        fluid
                        label='Username'
                        placeholder='Username'
                        value={username}
                        onChange={this.handleTextChange('username')}
                        autoFocus/>
                      <Form.Input
                        fluid
                        type='password'
                        label='Password'
                        placeholder='Password'
                        value={password}
                        onKeyDown={e => (e.key === 'Enter') && this.authUser()}
                        onChange={this.handleTextChange('password')} />
                    </Form.Group>
                    <Form.Group widths='equal' inline>
                      <Form.Checkbox
                        label='Save Creds'
                        checked={encryptCreds}
                        onChange={this.handleCheckChange('encryptCreds')}                              
                      />
                      <Form.Button
                        fluid
                        primary
                        onClick={this.authUser}>Authenticate</Form.Button>
                    </Form.Group>
                  </Form>
                </Card.Description>
                {(error && errorEvent==='login') ? <Message negative>{errorMessage}</Message> : ''}
              </Card.Content>
            </Card>

	);
    }
}

export default LoginForm;
