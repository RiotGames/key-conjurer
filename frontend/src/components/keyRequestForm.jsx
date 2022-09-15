import React, { Component } from "react";
import { Message, Form, Card } from "semantic-ui-react";
import * as PropTypes from "prop-types";
import { requestKeys } from "./../actions";
import { subscribe, update } from "./../stores";
import { documentationURL } from "../consts";

const timeouts = [1, 2, 3, 4, 5, 6, 7, 8];

const RoleInput = ({ onChange, value }) => {
  return (
    <>
      <Form.Input
        fluid
        label="Role"
        onChange={onChange}
        value={value}
        list="common-roles"
      />
      <datalist id="common-roles">
        <option value="GL-Power">GL-Power</option>
        <option value="GL-NetEng">GL-NetEng</option>
        <option value="GL-Admin">GL-Admin</option>
      </datalist>
      <Message info>
        The suggestions provided are there for convenience, but may not
        necessarily match up to the roles you have access to. You can find out
        what roles you have access to by checking out{" "}
        <a href={documentationURL}>this confluence page</a>.
      </Message>
    </>
  );
};


class KeyRequestForm extends Component {
  state = {
    accounts: [],
    keyRequest: false,
    password: "",
    requestSent: false,
    selectedAccount: undefined,
    timeout: parseInt(localStorage.getItem("timeout") || 1, 10),
    username: "",
    error: false,
    errorEvent: "",
    errorMessage: "",
    role: "",
  };

  handleChange = (name) => (event) => {
    if (name === "timeout") {
      localStorage.setItem("timeout", parseInt(event.currentTarget.value));
    }

    this.setState({ [name]: event.currentTarget.value });
  };

  setAccount = (event) => {
    this.setState({ selectedAccount: event.currentTarget.value });
  }

  componentDidMount() {
    subscribe("idpInfo", ({ apps: accounts }) => {
      // when idpInfo is triggered, we may have not selected an account yet, but
      // the DOM will visually indicate we have selected the first element.
      // The easiest way to fix this is to update the selected account to the first
      // available, if one has not already been selected.
      //
      // This doesn't get reproduced in tests because the tests can only test the
      // visual output to the browser, but this is happening within the internal state
      // of this component. Even when specifying that the browser should use the internal
      // state value, this still happens, which is very strange.
      this.setState((prevState) => {
        if (prevState.selectedAccount !== undefined) {
          // An account was selected by the user.
          if (accounts.length > 0 && accounts.every(acc => acc.id !== prevState.selectedAccount)) {
            // There are no accounts in the new account set that match the users current account, so we should unset.
            // This is unlikely to ever happen - a user would have to be removed from an account between two button presses.
            return { accounts, selectedAccount: undefined };
          }

          // The only case left that a user had selected an account, and is still entitled to it.
          return { accounts, selectedAccount: prevState.selectedAccount };
        }

        if (accounts.length > 0 && prevState.accounts.length === 0) {
          // The user was not entitled to any accounts, and now they are - we should pick the first one.
          // This resolves #18.
          return { accounts, selectedAccount: accounts[0].id };
        }

        // The user had no account preference, and there are no special cases for the accounts list,
        // so preserve the current selection.
        return { accounts }
      });
    });

    subscribe("userInfo", ({ username, password }) => {
      this.setState({
        username,
        password,
      });
    });
    subscribe("request", ({ requestSent }) => {
      this.setState({
        requestSent,
        keyRequest: requestSent && this.state.keyRequest,
      });
    });
    subscribe("errors", ({ error, event, message: errorMessage }) => {
      if (event === "keyRequest") {
        this.setState({ error, errorMessage, errorEvent: event });
      }
    });
  }

  handleSubmit = (_event) => {
    const { username, password, selectedAccount, timeout, role } = this.state;
    this.setState({
      keyRequest: true,
      error: false,
      errorEvent: "",
      errorMessage: "",
    });

    requestKeys({
      username,
      password,
      selectedAccount,
      timeout,
      idp: this.props.idp,
      role,
    });
  };

  render() {
    const {
      accounts,
      keyRequest,
      requestSent,
      timeout,
      error,
      errorEvent,
      errorMessage,
      role,
      selectedAccount
    } = this.state;

    const accountPlaceHolder = accounts.length ? "Accounts" : "No Accounts";
    // Dropdowns use raw HTML with Semantic UI React classes to make them accessible.
    // Elements must be accessible to be accessed from React Testing Library in tests
    // (and also to obey California law).
    //
    // Semantic UI React uses divs for everything which cannot be labeled and is thus
    // not accessible.
    return (
      <Card fluid>
        <Card.Content>
          <Card.Header>Key Request</Card.Header>
          <Card.Meta>Will MFA push on request</Card.Meta>
          <Card.Description>
            <Form loading={requestSent} onSubmit={this.handleSubmit}>
              <Form.Group widths='equal'>
                <Form.Field>
                  <label htmlFor="account">Account</label>
                  <select className='ui field dropdown selection' id="account" placeholder={accountPlaceHolder} onChange={this.setAccount} value={selectedAccount}>
                    {accounts.map(({ name, id }) =>
                      <option key={id} value={id}>{name}</option>
                    )}
                  </select>
                </Form.Field>

                <Form.Field>
                  <label htmlFor="timeout">TTL (hours)</label>
                  <select className='ui field dropdown selection' id="timeout" onChange={this.handleChange("timeout")} value={timeout}>
                    {timeouts.map((timeout) =>
                      <option key={timeout} value={timeout}>{timeout}</option>
                    )}
                  </select>
                </Form.Field>
              </Form.Group>
              <RoleInput value={role} onChange={this.handleChange("role")} />
              <Form.Button fluid primary type="submit">
                Request Keys
              </Form.Button>
            </Form>
            {keyRequest ? <Message positive>Sending Push</Message> : ""}
            {error && errorEvent === "keyRequest" ? (
              <Message negative>{errorMessage}</Message>
            ) : (
              ""
            )}
          </Card.Description>
        </Card.Content>
      </Card >
    );
  }
}

KeyRequestForm.propTypes = {
  idp: PropTypes.string.isRequired,
};

export default KeyRequestForm;
