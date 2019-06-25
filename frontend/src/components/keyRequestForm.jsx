import React, { Component } from 'react';
import { Message, Ref, Form, Card } from 'semantic-ui-react';

import { requestKeys } from './../actions';
import { subscribe } from './../stores';

const timeouts = [1,2,3,4,5,6,7,8];

class KeyRequestForm extends Component {
    state = {
	accounts: [],
	keyRequest: false,
	password: '',
	requestSent: false,
	selectedAccount: '',
	timeout: parseInt(localStorage.getItem('timeout') || 1, 10),
	username: '',
        error: false,
        errorEvent: '',
        errorMessage:''
    };
    accountsField = React.createRef()


    handleChange =
	name =>
	(event, data) => {
	    if (name === 'timeout') {
		localStorage.setItem('timeout', data.value);
	    }
	    this.setState({[name]: data.value});
	}

    requestKeys =
	event => {
	    const {username, password, selectedAccount, timeout} = this.state;
	    this.setState({
                keyRequest: true,
		error: false,
                errorEvent:'',
		errorMessage: ''
            }, _ => requestKeys({username, password, selectedAccount, timeout}));

	};

    setAccount = (event, data) => this.setState({selectedAccount: data.value})

    componentDidMount() {
	subscribe('idpInfo', ({apps: accounts}) => {
	    this.setState({
		accounts
	    }, () => {
	        if (accounts.length) {
                    this.accountsField.current.lastChild.firstElementChild.focus();
	        } else {
		    this.setState({
		        accounts: [],
		        Keyrequest: false,
		        selectedAccount: ''
		    });
	        }});
	});
	subscribe('userInfo', ({username, password}) => {
	    this.setState({
		username,
		password
	    });
	});
	subscribe('request', ({requestSent}) => {
	    this.setState({
		requestSent,
		keyRequest: (requestSent && this.state.keyRequest)
	    });
	});
	subscribe('errors', ({error, event, message: errorMessage}) => {
	    if (event === 'keyRequest') {
		this.setState({error, errorMessage, errorEvent: event});
	    }
	});
    }

    render() {
	const {
	    accounts,
	    keyRequest,
	    requestSent,
	    timeout,
            error,
            errorEvent,
            errorMessage
	} = this.state;

	const timeoutOptions = timeouts.map(timeout => ({key: timeout, value: timeout, text: timeout}));
        const accountPlaceHolder = accounts.length ? 'Accounts' : 'No Accounts';
        const accountOptions = accounts.map(({name, id}) => ({key:id, value: id, text:name}));

	return (
	    <Card fluid>
              <Card.Content>
                <Card.Header>
                  Key Request 
                </Card.Header>
                <Card.Meta>
                  Will MFA push on request
                </Card.Meta>
                <Card.Description>
                  <Form loading={requestSent}>
                    <Form.Group>
                      <Ref innerRef={this.accountsField}>
                      <Form.Dropdown
                        fluid
                        search
                        selection
                        label='Account'
                        onChange={this.setAccount}
                        placeholder={`${accountPlaceHolder}`}
                        width={14}
                        options={accountOptions} />
                      </Ref>
                        <Form.Dropdown
                          fluid
                          selection
                          label='TTL'
                          onChange={this.handleChange('timeout')}
                          width={2}
                          defaultValue={timeout}
                          options={timeoutOptions}/>
                    </Form.Group>
                    <Form.Group widths='equal'>
                        <Form.Button
                          fluid
                          primary
                          onClick={this.requestKeys}>Request Keys</Form.Button>
                    </Form.Group>
                  </Form>
                  {keyRequest ? <Message positive>Sending Push</Message> : ''}
                  {(error && errorEvent==='keyRequest') ? <Message negative>{errorMessage}</Message> : ''}
                </Card.Description>
              </Card.Content>
	    </Card>
	);
    }
}

export default KeyRequestForm;
