import React, { Component } from "react";
import { Message, Ref, Form, Card } from "semantic-ui-react";
import * as PropTypes from "prop-types";
import { requestKeys } from "./../actions";
import { subscribe } from "./../stores";
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
    selectedAccount: "",
    timeout: parseInt(localStorage.getItem("timeout") || 1, 10),
    username: "",
    error: false,
    errorEvent: "",
    errorMessage: "",
    role: "",
  };
  accountsField = React.createRef();

  handleChange = (name) => (event, data) => {
    if (name === "timeout") {
      localStorage.setItem("timeout", data.value);
    }
    this.setState({ [name]: data.value });
  };

  setAccount = (event, data) => this.setState({ selectedAccount: data.value });

  componentDidUpdate(_prevProps, prevState) {
    if (prevState.accounts.length !== this.state.accounts.length) {
      this.accountsField.current.lastChild.firstElementChild.focus();
    }
  }

  componentDidMount() {
    subscribe("idpInfo", ({ apps: accounts }) => {
      this.setState((prevState) => {
        return {
          accounts,
          selectedAccount:
            accounts.length === 0 ? "" : prevState.selectedAccount,
        };
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
    console.log(this.state);
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
    } = this.state;

    const timeoutOptions = timeouts.map((timeout) => ({
      key: timeout,
      value: timeout,
      text: timeout,
    }));
    const accountPlaceHolder = accounts.length ? "Accounts" : "No Accounts";
    const accountOptions = accounts.map(({ name, id }) => ({
      key: id,
      value: id,
      text: name,
    }));

    return (
      <Card fluid>
        <Card.Content>
          <Card.Header>Key Request</Card.Header>
          <Card.Meta>Will MFA push on request</Card.Meta>
          <Card.Description>
            <Form loading={requestSent} onSubmit={this.handleSubmit}>
              <Form.Group>
                <Ref innerRef={this.accountsField}>
                  <Form.Dropdown
                    fluid
                    search
                    selection
                    label="Account"
                    onChange={this.setAccount}
                    placeholder={accountPlaceHolder}
                    width={14}
                    options={accountOptions}
                  />
                </Ref>
                <Form.Dropdown
                  fluid
                  selection
                  label="TTL (hours)"
                  onChange={this.handleChange("timeout")}
                  width={2}
                  defaultValue={timeout}
                  options={timeoutOptions}
                />
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
      </Card>
    );
  }
}

KeyRequestForm.propTypes = {
  idp: PropTypes.string.isRequired,
};

export default KeyRequestForm;
